package main

import (
	"context"
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tavo-wasd-gh/gosmtp"
	"github.com/tavo-wasd-gh/webapp/config"
	"github.com/tavo-wasd-gh/webapp/database"
	"github.com/tavo-wasd-gh/webtoolkit/auth"
	"github.com/tavo-wasd-gh/webtoolkit/cors"
	"github.com/tavo-wasd-gh/webtoolkit/forms"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
	"github.com/tavo-wasd-gh/webtoolkit/s3"
	"github.com/tavo-wasd-gh/webtoolkit/serve"
	"github.com/tavo-wasd-gh/webtoolkit/views"
)

const (
	AccessTokenKey  = "access_token"
	RefreshTokenKey = "refresh_token"
	AccessTokenTTL  = 10 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

const (
    DefaultMaxUploadSize = 10 << 20 // 10 MB
    DefaultUploadTimeout = 30 * time.Minute
)

type App struct {
	Production bool
	Views      map[string]*template.Template
	Log        *logger.Logger
	DB         *sqlx.DB
	// HTTP
	AllowOrigin string
	// JWT
	Secret       string
	AccessToken  string
	RefreshToken string
	// S3
	S3 *s3.Client
}

type JwtClaims struct {
	UserID    int
	AccountID int
}

//go:embed static/*
var publicFS embed.FS

//go:embed templates/*
var viewFS embed.FS

func main() {
	// Configure in config/config.go
	env, err := config.Init()
	if err != nil {
		log.Fatalf("%v", logger.Errorf("failed to initialize config: %v", err))
	}

	// Configure in config/views.go
	views, err := views.Init(viewFS, config.ViewMap, config.FuncMap)
	if err != nil {
		log.Fatalf("%v", logger.Errorf("failed to initialize views: %v", err))
	}

	// Configure in database/database.go
	db, err := database.Init(env.DBConnDvr, env.DBConnStr)
	if err != nil {
		log.Fatalf("%v", logger.Errorf("failed to initialize database: %v", err))
	}
	defer db.Close()

	s3 := s3.New("./data")

	var allowOrigin string

	if env.AllowOrigin != "" {
		allowOrigin = env.AllowOrigin
	} else {
		allowOrigin = "*"
	}

	app := &App{
		Production:  env.Production,
		Views:       views,
		Log:         &logger.Logger{Enabled: env.Debug},
		DB:          db,
		AllowOrigin: allowOrigin,
		Secret:      env.Secret,
		S3:          s3,
	}

	// Views (Auth required)
	http.HandleFunc("/api/user/toggle", app.handleUserToggle)
	http.HandleFunc("/api/user/add", app.handleAddUser)

	// Endpoints (Public)
	http.HandleFunc("/api/logout", app.handleLogout)

	// Pages (Auth required)
	http.HandleFunc("/panel", app.handlePanel)
	http.HandleFunc("/cuenta", app.handleDashboard)

	// Pages (Public)
	http.HandleFunc("/", app.handleIndex)
	http.HandleFunc("/proveedores", app.handleSuppliers)
	http.HandleFunc("/fse", app.handleFSE)

	// Set JWT
	http.HandleFunc("/api/login", app.handleLoginForm)

	// Serve files in static/
	staticFiles, err := fs.Sub(publicFS, "static")
	if err != nil {
		log.Fatalf("%v", logger.Errorf("failed to create sub filesystem: %v", err))
	}

	http.Handle(
		"/s/",
		serve.Compressed(
			http.StripPrefix("/s/", http.FileServer(http.FS(staticFiles))),
		),
	)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		app.Log.Printf("starting on :%s...", env.Port)

		if err := http.ListenAndServe(":"+env.Port, nil); err != nil {
			log.Fatalf("%v", logger.Errorf("failed to start server: %v", err))
		}
	}()

	<-stop
}

func (app *App) handleUserToggle(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	if _, _, err := app.validateSession(w, r, database.WriteAdvanced); err != nil {
		w.Header().Set("HX-Redirect", "/cuenta")
		return
	}

	type ToggleUser struct {
		ID int `form:"user_id" req:"1"`
	}

	var toggleUser ToggleUser

	if err := forms.ParseForm(r, &toggleUser); err != nil {
		app.Log.Errorf("error parsing toggle user form: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	var err error = nil
	_, err = database.ToggleUser(app.DB, toggleUser.ID)

	if err != nil {
		app.Log.Errorf("error toggling user state: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (app *App) handleAddUser(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	if _, _, err := app.validateSession(w, r, database.WriteAdvanced); err != nil {
		w.Header().Set("HX-Redirect", "/cuenta")
		return
	}

	type NewUser struct {
		ID     int
		Email  string `form:"email" req:"1"`
		Name   string `form:"name" req:"1"`
		Active bool
	}

	var newUser NewUser

	if err := forms.ParseForm(r, &newUser); err != nil {
		app.Log.Errorf("error parsing add user form: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	var err error = nil
	newUser.ID, newUser.Active, err = database.AddUser(app.DB, newUser.Email, newUser.Name)

	if err != nil {
		app.Log.Errorf("error adding user: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err := views.Render(w, r, app.Views["user"], newUser); err != nil {
		app.Log.Errorf("error rendering template %s: %v", "user", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (app *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	expiredAccessToken := &http.Cookie{
		Name:     AccessTokenKey,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   app.Production,
	}
	expiredRefreshToken := &http.Cookie{
		Name:     RefreshTokenKey,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   app.Production,
	}
	http.SetCookie(w, expiredAccessToken)
	http.SetCookie(w, expiredRefreshToken)
	w.WriteHeader(http.StatusOK)
}

func (app *App) handlePanel(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	_, _, err := app.validateSession(w, r, database.ReadAdvanced)
	if err != nil {
		app.Log.Errorf("error validating session: %v", err)
		http.Redirect(w, r, "/cuenta", http.StatusSeeOther)
		return
	}

	var panel database.Panel
	if err := panel.Load(app.DB); err != nil {
		app.Log.Errorf("error loading panel: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err := views.Render(w, r, app.Views["panel-page"], panel); err != nil {
		app.Log.Errorf("error rendering template %s: %v", "panel", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (app *App) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	userID, accountID, err := app.validateSession(w, r, database.Read)
	if err != nil {
		if err := views.Render(w, r, app.Views["login-page"], nil); err != nil {
			app.Log.Errorf("error rendering template %s: %v", "login", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		return
	}

	var dashboard database.Dashboard
	if err := dashboard.Load(app.DB, userID, accountID); err != nil {
		app.Log.Errorf("error loading data for user %d with account %d: %v", userID, accountID, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err := views.Render(w, r, app.Views["dashboard-page"], dashboard); err != nil {
		app.Log.Errorf("error rendering template %s: %v", "dashboard", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (app *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	// TODO: Load data for index
	//var data database.IndexData

	if err := views.Render(w, r, app.Views["index-page"], nil); err != nil {
		app.Log.Errorf("error rendering template %s: %v", "index", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (app *App) handleSuppliers(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	// TODO: Load data for suppliers

	if err := views.Render(w, r, app.Views["suppliers-page"], nil); err != nil {
		app.Log.Errorf("error rendering template %s: %v", "index", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (app *App) handleFSE(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	// TODO: Load data for FSE

	if err := views.Render(w, r, app.Views["fse-page"], nil); err != nil {
		app.Log.Errorf("error rendering template %s: %v", "index", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (app *App) handleLoginForm(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, app.AllowOrigin, "POST, OPTIONS", "Content-Type", false) {
		return
	}

	requiredPermission := database.Read

	type loginForm struct {
		Email    string `form:"email" req:"1"`
		Password string `form:"password" req:"1"`
		Account  string `form:"account"`
	}

	var login loginForm

	if err := forms.ParseForm(r, &login); err != nil {
		app.Log.Errorf("error parsing login form: %v", err)

		if err := views.Render(w, r, app.Views["login"], map[string]any{"Error": true}); err != nil {
			app.Log.Errorf("error rendering template %s: %v", "login", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		return
	}

	if strings.Contains(login.Email, "@") {
		if !strings.Contains(strings.ToLower(login.Email), "@ucr.ac.cr") {
			// Is an external provider

			if err := views.Render(w, r, app.Views["login"], map[string]any{"ExternalEmail": true}); err != nil {
				app.Log.Errorf("error rendering template %s: %v", "login", err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			return
		}
	} else {
		login.Email += "@ucr.ac.cr"
	}

	userID, err := database.UserIDByUserEmail(app.DB, login.Email)
	if err != nil {
		app.Log.Errorf("error looking for user %s: %v", login.Email, err)
		if err := views.Render(w, r, app.Views["login"], map[string]any{"Error": true}); err != nil {
			app.Log.Errorf("error rendering template %s: %v", "login", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		return
	}

	allowedAccounts, err := database.AllowedAccountsByUserID(app.DB, userID)
	if err != nil {
		app.Log.Errorf("no allowed accounts for user %s: %v", login.Email, err)
		if err := views.Render(w, r, app.Views["login"], map[string]any{"Error": true}); err != nil {
			app.Log.Errorf("error rendering template %s: %v", "login", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		return
	}

	var chosenAccountID int

	switch len(allowedAccounts) {
	case 0:
		// No allowed accounts, render error and return
		if err := views.Render(w, r, app.Views["login"], map[string]any{"Error": true}); err != nil {
			app.Log.Errorf("error rendering template %s: %v", "login", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		return

	case 1:
		// One allowed account, set and continue
		chosenAccountID = allowedAccounts[0].ID

	default:
		type multiple struct {
			MultipleAccounts bool
			AllowedAccounts  []database.Account
		}

		m := multiple{
			MultipleAccounts: true,
			AllowedAccounts:  allowedAccounts,
		}

		if err := views.Render(w, r, app.Views["login"], m); err != nil {
			app.Log.Errorf("error rendering template %s: %v", "login", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		return
	}

	if app.Production {
		s := smtp.Client("smtp.ucr.ac.cr", "587", login.Password)

		if err := s.Validate(login.Email); err != nil {
			app.Log.Errorf("error validating user %s: %v", login.Email, err)

			if err := views.Render(w, r, app.Views["login"], map[string]any{"Error": true}); err != nil {
				app.Log.Errorf("error rendering template %s: %v", "login", err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}

			return
		}
	}

	claims := JwtClaims{
		UserID:    userID,
		AccountID: chosenAccountID,
	}

	if err = auth.JwtSet(w, app.Production, "/", AccessTokenKey, claims, time.Now().Add(AccessTokenTTL), app.Secret); err != nil {
		app.Log.Errorf("error setting JWT cookie: %v", err)

		if err := views.Render(w, r, app.Views["login"], map[string]any{"Error": true}); err != nil {
			app.Log.Errorf("error rendering template %s: %v", "login", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		return
	}

	if err = auth.JwtSet(w, app.Production, "/", RefreshTokenKey, claims, time.Now().Add(RefreshTokenTTL), app.Secret); err != nil {
		app.Log.Errorf("error setting JWT cookie: %v", err)

		if err := views.Render(w, r, app.Views["login"], map[string]any{"Error": true}); err != nil {
			app.Log.Errorf("error rendering template %s: %v", "login", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		return
	}

	session := database.Session{
		UserID:    claims.UserID,
		AccountID: claims.AccountID,
	}

	if err := session.Validate(app.DB, requiredPermission); err != nil {
		app.Log.Errorf("error validating session: %v", err)
		if err := views.Render(w, r, app.Views["login"], nil); err != nil {
			app.Log.Errorf("error rendering template %s: %v", "login", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		return
	}

	var dashboard database.Dashboard
	if err := dashboard.Load(app.DB, userID, chosenAccountID); err != nil {
		app.Log.Errorf("error loading data for user %d with account %d: %v", userID, chosenAccountID, err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err := views.Render(w, r, app.Views["dashboard"], dashboard); err != nil {
		app.Log.Errorf("error rendering template %s: %v", "dashboard", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (app *App) validateSession(w http.ResponseWriter, r *http.Request, p database.PermissionInteger) (int, int, error) {
	var claims JwtClaims

	if err := auth.JwtValidate(r, "/", AccessTokenKey, app.Secret, &claims); err != nil {
		if err := auth.JwtValidate(r, "/", RefreshTokenKey, app.Secret, &claims); err != nil {
			return 0, 0, logger.Errorf("error validating jwt: %v", err)
		}

		if err := auth.JwtSet(w, app.Production, "/", AccessTokenKey, claims, time.Now().Add(AccessTokenTTL), app.Secret); err != nil {
			return 0, 0, logger.Errorf("error setting jwt: %v", err)
		}

		if err := auth.JwtSet(w, app.Production, "/", RefreshTokenKey, claims, time.Now().Add(RefreshTokenTTL), app.Secret); err != nil {
			return 0, 0, logger.Errorf("error setting jwt: %v", err)
		}
	}

	session := database.Session{
		UserID:    claims.UserID,
		AccountID: claims.AccountID,
	}

	if err := session.Validate(app.DB, p); err != nil {
		return 0, 0, logger.Errorf("error validating session: %v", err)
	}

	return session.UserID, session.AccountID, nil
}

func (app *App) uploadFile(w http.ResponseWriter, r *http.Request, formName, bucket, key string, allowedMimeTypes []string) error {
	r.Body = http.MaxBytesReader(w, r.Body, DefaultMaxUploadSize)
	_, cancel := context.WithTimeout(r.Context(), DefaultUploadTimeout)
	defer cancel()

	file, _, err := r.FormFile(formName)
	if err != nil {
		return logger.Errorf("error reading file: %v", err)
	}
	defer file.Close()

	err = s3.VerifyFileType(file, allowedMimeTypes)
	if err != nil {
		return logger.Errorf("error verifying mime type: %v", err)
	}

	err = app.S3.PutObject(bucket, key, file)
	if err != nil {
		return logger.Errorf("error uploading file: %v", err)
	}

	return nil
}
