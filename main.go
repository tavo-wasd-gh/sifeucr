package main

import (
	"log"
	"database/sql"
	"net/http"
	"html/template"
	"os"
	"os/signal"
	"syscall"

	"github.com/tavo-wasd-gh/sifeucr/db"
	"github.com/tavo-wasd-gh/sifeucr/views"
	"github.com/tavo-wasd-gh/gocors"
)

type App struct {
	Production bool
	Views      map[string]*template.Template
	DB         *sql.DB
}

func main() {
	var (
		production = os.Getenv("ENVIRONMENT") == "production"
		port       = os.Getenv("PORT")
		connStr    = os.Getenv("DB_CONNSTR")
		connDvr    = os.Getenv("DB_CONNDVR")
	)

	if port == "" || connStr == "" || connDvr == "" {
		log.Fatalf("Fatal: Missing env variables")
	}

	viewsMap := map[string]string{
		"Login":    "views/login.html",
		"Dashboard": "views/dashboard.html",
	}

	views, err := views.Init(viewsMap)
	if err != nil {
		log.Fatalf("Fatal: Failed to initialize templates: %v", err)
	}

	db, err := db.Init(connDvr, connStr)
	if err != nil {
		log.Fatalf("Fatal: Failed to initialize database: %v", err)
	}
	defer db.Close()

	app := &App{
		Production: production,
		Views:      views,
		DB:         db,
	}

	http.HandleFunc("/api/dashboard", app.handleDashboard)
	http.Handle("/", http.FileServer(http.Dir("public")))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Log: Running on :" + port + "...")
		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("Fatal: Failed to start on port %s: %v", port, err)
		}
	}()

	<-stop

	log.Println("Log: Shutting down...")
}

func (app *App) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}
}

func (app *App) Render(w http.ResponseWriter, name string, data interface{}) {
	tmpl, ok := app.Views[name]

	if !ok {
		http.Error(w, "View not found", http.StatusNotFound)
		return
	}

	if err := tmpl.Execute(w, data); err != nil {
		http.Error(w, "Failed to render view", http.StatusInternalServerError)
	}
}
