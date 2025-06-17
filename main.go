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
	"github.com/tavo-wasd-gh/webapp/config"
	"github.com/tavo-wasd-gh/webapp/database"
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

//go:embed static/*
var publicFS embed.FS

//go:embed templates/*.html templates/**/*.html
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
	http.HandleFunc("/", app.handleIndex)
	http.HandleFunc("/dashboard", app.handleDashboard)

	// Serve files in static/
	staticFiles, err := fs.Sub(publicFS, "static")
	if err != nil {
		log.Fatalf("%v", logger.Errorf("failed to create sub filesystem: %v", err))
	}

	http.Handle("/s/", http.StripPrefix("/s/", http.FileServer(http.FS(staticFiles))))

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

func (app *App) handleIndex(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	if err := views.Render(w, app.Views["index"], nil); err != nil {
		app.Log.Errorf("error rendering template: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (app *App) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	if err := views.Render(w, app.Views["dashboard"], nil); err != nil {
		app.Log.Errorf("error rendering template: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
