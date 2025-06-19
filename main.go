package main

import (
	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tavo-wasd-gh/webapp/config"
	"github.com/tavo-wasd-gh/webapp/database"
	"github.com/tavo-wasd-gh/webtoolkit/auth"
	"github.com/tavo-wasd-gh/webtoolkit/cors"
	"github.com/tavo-wasd-gh/webtoolkit/forms"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
	"github.com/tavo-wasd-gh/webtoolkit/views"
)

type App struct {
	Production bool
	Views      map[string]*template.Template
	Log        *logger.Logger
	DB         *sqlx.DB
	// JWT
	Secret string
	Cookie string
}

type JwtClaims struct {
	Email string
	// TODO: Add logic to know which account a user is going to use
	//Account string
}

//go:embed static/*
var publicFS embed.FS

//go:embed templates/*
var viewFS embed.FS

func main() {
	// Configure environment in config/config.go
	env, err := config.Init()
	if err != nil {
		log.Fatalf("%v", logger.Errorf("failed to initialize config: %v", err))
	}

	// Configure views in config/views.go
	views, err := views.Init(viewFS, config.ViewMap, config.FuncMap)
	if err != nil {
		log.Fatalf("%v", logger.Errorf("failed to initialize views: %v", err))
	}

	// Defaults to "sqlite3" and "./db.db" if not set, modify in database/database.go
	db, err := database.Init(env.DBConnDvr, env.DBConnStr)
	if err != nil {
		log.Fatalf("%v", logger.Errorf("failed to initialize database: %v", err))
	}
	defer db.Close()

	app := &App{
		Production: env.Production,
		Views:      views,
		Log:        &logger.Logger{Enabled: env.Debug},
		DB:         db,
		Secret:     env.Secret,
		Cookie:     "session",
	}

	// Pages
	http.HandleFunc("/cuenta", app.handleDashboard)
	http.HandleFunc("/", app.handleStaticTemplate("index-page"))

	// Serve files in static/
	staticFiles, err := fs.Sub(publicFS, "static")
	if err != nil {
		log.Fatalf("%v", logger.Errorf("failed to create sub filesystem: %v", err))
	}
	http.Handle("/s/", http.StripPrefix("/s/", http.FileServer(http.FS(staticFiles))))

	// Views
	http.HandleFunc("/api/login", app.handleLoginForm)

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

// -----------------------
// STATIC PAGES RENDERERS
// -----------------------

func (app *App) handleStaticTemplate(name string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
			return
		}

		if err := views.Render(w, app.Views[name], nil); err != nil {
			app.Log.Errorf("error rendering template %s: %v", name, err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}

// -----------------------
// DYNAMIC PAGES RENDERERS
// -----------------------

func (app *App) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	var claims JwtClaims

	if err := auth.JwtValidate(r, "/", app.Cookie, app.Secret, &claims); err != nil {
		// app.Log.Errorf("error validating JWT: %v", err) // DEBUG

		if err := views.Render(w, app.Views["login-page"], nil); err != nil {
			app.Log.Errorf("error rendering template %s: %v", "login", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		return
	}

	// TODO: Load data for dashboard
	//var data database.DashboardData

	if err := views.Render(w, app.Views["dashboard-page"], nil); err != nil {
		app.Log.Errorf("error rendering template %s: %v", "dashboard", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

// ------------------
// HTML API ENDPOINTS
// ------------------

func (app *App) handleLoginForm(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	type loginForm struct {
		Email    string `form:"email"`
		Password string `form:"password"`
	}

	var login loginForm

	if err := forms.ParseForm(r, &login); err != nil {
		app.Log.Errorf("error parsing login form: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	// TODO: Validate form

	claims := JwtClaims{
		Email: login.Email,
		//Account: "SF",
	}

	if err := auth.JwtSet(w, app.Production, "/", app.Cookie, claims, time.Now().Add(time.Hour), app.Secret); err != nil {
		app.Log.Errorf("error setting JWT cookie: %v", err)

		if err := views.Render(w, app.Views["login"], map[string]any{"Error": true}); err != nil {
			app.Log.Errorf("error rendering template %s: %v", "login", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		return
	}

	if err := views.Render(w, app.Views["dashboard"], nil); err != nil {
		app.Log.Errorf("error rendering template %s: %v", "dashboard", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
