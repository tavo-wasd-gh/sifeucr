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

	"github.com/jmoiron/sqlx"
	"github.com/tavo-wasd-gh/gosmtp"
	"github.com/tavo-wasd-gh/sifeucr/config"
	"github.com/tavo-wasd-gh/sifeucr/database"
	"github.com/tavo-wasd-gh/sifeucr/auth"
	"github.com/tavo-wasd-gh/webtoolkit/cors"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
	"github.com/tavo-wasd-gh/webtoolkit/views"
)

type App struct {
	Production bool
	Views      map[string]*template.Template
	Log        *logger.Logger
	DB         *sqlx.DB
	Secret     string
}

//go:embed public/*
var publicFS embed.FS

//go:embed templates/*.html
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
	}

	// Handlers
	http.HandleFunc("/api/dashboard", app.handleDashboard)
	http.HandleFunc("/api/panel", app.handlePanel)
	http.HandleFunc("/api/login", app.handleLogin)

	// Serve files in public/
	staticFiles, err := fs.Sub(publicFS, "public")
	if err != nil {
		log.Fatalf("%v", logger.Errorf("failed to create sub filesystem: %v", err))
	}

	http.Handle("/", http.FileServer(http.FS(staticFiles)))

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

func (app *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, POST, OPTIONS", "Content-Type", false) {
		return
	}

	if r.Method == http.MethodGet {
		if err := views.Render(w, app.Views["login"], nil); err != nil {
			app.Log.Errorf("error rendering template: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			app.Log.Errorf("error parsing form: %v", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		account := r.FormValue("account")
		passwd := r.FormValue("passwd")

		if app.Production {
			s := smtp.Client("smtp.ucr.ac.cr", "587", passwd)
			if err := s.Validate(email); err != nil {
				app.Log.Errorf("error validating email: %v", err)
				http.Error(w, "", http.StatusUnauthorized)
				return
			}
		}

		user := database.User{
			Email: email,
			Account: database.Account{
				ID: account,
			},
		}

		if err := user.Login(app.DB); err != nil {
			app.Log.Errorf("error logging in: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		if len(user.AvailableAccounts) > 1 {
			if err := views.Render(w, app.Views["login"], user); err != nil {
				app.Log.Errorf("error rendering template: %v", err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
		}

		if user.Account.ID == "" {
			app.Log.Errorf("error setting Account.ID")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		auth.JwtSet(w,
			app.Production,
			"token",
			email,
			account,
			config.CookieTimeout(),
			app.Secret,
		)
	}
}

func (app *App) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	email, account, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		w.Header().Set("HX-Redirect", "/login")
		return
	}

	user := database.User{
		Email: email,
		Account: database.Account{
			ID: account,
		},
	}

	if err := user.Login(app.DB); err != nil {
		app.Log.Errorf("error logging in: %v", err)
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if err := views.Render(w, app.Views["dashboard"], user); err != nil {
		app.Log.Errorf("error rendering template: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (app *App) handlePanel(w http.ResponseWriter, r *http.Request) {
}
