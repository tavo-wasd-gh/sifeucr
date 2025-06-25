package main

import (
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
	"github.com/tavo-wasd-gh/webtoolkit/serve"
	"github.com/tavo-wasd-gh/webtoolkit/views"
)

type App struct {
	Production bool
	Views      map[string]*template.Template
	Log        *logger.Logger
	DB         *sqlx.DB
	// HTTP
	AllowOrigin string
	// JWT
	Secret string
	Cookie string
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
		Cookie:      "session",
	}

	// Views (Auth required)
	http.HandleFunc("/api/user/toggle", app.handleUserToggle)
	http.HandleFunc("/api/user/add", app.handleAddUser)

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

	if _, _, err := app.validateSession(r, database.WriteAdvanced); err != nil {
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

	if _, _, err := app.validateSession(r, database.WriteAdvanced); err != nil {
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

func (app *App) handlePanel(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	_, _, err := app.validateSession(r, database.ReadAdvanced)
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

	userID, accountID, err := app.validateSession(r, database.Read)
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

	if err := auth.JwtSet(w, app.Production, "/", app.Cookie, claims, time.Now().Add(10*time.Second), app.Secret); err != nil {
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

func (app *App) validateSession(r *http.Request, p database.PermissionInteger) (int, int, error) {
	var claims JwtClaims
	if err := auth.JwtValidate(r, "/", app.Cookie, app.Secret, &claims); err != nil {
		return 0, 0, logger.Errorf("error validating jwt: %v", err)
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
