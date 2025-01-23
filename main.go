package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
	"os/signal"
	"syscall"

	"github.com/tavo-wasd-gh/gocors"
	"github.com/tavo-wasd-gh/sifeucr/database"
	"github.com/tavo-wasd-gh/sifeucr/views"
	"github.com/tavo-wasd-gh/sifeucr/auth"
)

type App struct {
	Production bool
	Debug      bool
	Secret     string
	Views      map[string]*template.Template
	DB         *sql.DB
}

func main() {
	var (
		production = os.Getenv("PRODUCTION") == "1"
		debug      = os.Getenv("DEBUG") == "1"
		secret     = os.Getenv("SECRET")
		port       = os.Getenv("PORT")
		connStr    = os.Getenv("DB_CONNSTR")
		connDvr    = os.Getenv("DB_CONNDVR")
	)

	if port == "" || connStr == "" || connDvr == "" || secret == "" {
		log.Fatalf("main: fatal: missing env variables")
	}

	viewsMap := map[string]string{
		"login":     "views/login.html",
		"dashboard": "views/dashboard.html",
	}

	views, err := views.Init(viewsMap)
	if err != nil {
		log.Fatalf("main: fatal: failed to initialize templates: %v", err)
	}

	db, err := database.Init(connDvr, connStr)
	if err != nil {
		log.Fatalf("main: fatal: failed to initialize database: %v", err)
	}
	defer db.Close()

	app := &App{
		Production: production,
		Debug:      debug,
		Secret:     secret,
		Views:      views,
		DB:         db,
	}

	http.HandleFunc("/api/dashboard", app.handleDashboard)
	http.Handle("/", http.FileServer(http.Dir("public")))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		app.log("main: privilegio SF: %d", auth.SF)
		app.log("main: privilegio COES: %d", auth.COES)
		app.log("main: privilegio Regular: %d", auth.Regular)
		app.log("main: starting on :%s...", port)

		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("main: fatal: failed to start on port %s: %v", port, err)
		}
	}()

	<-stop

	app.log("main: shutting down...")
}

func (app *App) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, POST, OPTIONS", "Content-Type", false) {
		return
	}

	if r.Method == http.MethodGet {
		correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
		if err != nil {
			app.log("handleDashboard: error validating token: %v", err)
			app.Render(w, "login", nil)
			return
		}

		u, err := database.Login(app.DB, correo, cuenta)
		if err != nil {
			app.log("handleDashboard: error logging in: %v", err)
			app.Render(w, "login", nil)
			return
		}

		if err := app.Render(w, "dashboard", u) ; err != nil {
			app.log("handleDashboard: error rendering view: %v", err)
			return
		}

		return
	}

	if r.Method == http.MethodPost {
		correo, cuenta, err := app.ValidateForm(r, w)
		if err != nil {
			app.log("handleDashboard: error validating form: %v", err)
			return
		}

		u, err := database.Login(app.DB, correo, cuenta)
		if err != nil {
			app.log("handleDashboard: error logging in: %v", err)
			app.Render(w, "login", nil)
			return
		}

		if err := app.Render(w, "dashboard", u) ; err != nil {
			app.log("handleDashboard: error rendering view: %v", err)
		}

		return
	}
}

func (app *App) Render(w http.ResponseWriter, name string, data interface{}) error {
	tmpl, ok := app.Views[name]

	if !ok {
		return fmt.Errorf("no template '%s' available", name)
	}

	if data == nil {
		data = map[string]interface{}{}
	}

	if err := tmpl.Execute(w, data); err != nil {
		return fmt.Errorf("error executing template '%s': %v", name, err)
	}

	return nil
}

func (app *App) ValidateForm(r *http.Request, w http.ResponseWriter) (string, string, error) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return "", "", err
	}

	cuenta := ""
	correo := r.FormValue("correo")
	passwd := r.FormValue("passwd")
	cuentaPedida := r.FormValue("cuenta")

	cuentas, err := database.CuentasActivas(app.DB, correo) 
	if err != nil {
		err = nil

		correo = correo + "@ucr.ac.cr"
		cuentas, err = database.CuentasActivas(app.DB, correo)
		if err != nil {
			http.Error(w, "", http.StatusUnauthorized)
			return "", "", err
		}
	}

	if len(cuentas) > 1 {
		found := false
		if cuentaPedida == "" {
			// Falta el caso en el que hay multiples cuentas, mostrar
			// login con opciones de las cuentas activas
			app.Render(w, "login", nil)
			return "", "", fmt.Errorf("more than 1 available account")
		} else {
			for _, i := range cuentas {
				if i == cuentaPedida {
					found = true
					cuenta = cuentaPedida
					break
				}
			}
			if !found {
				http.Error(w, "", http.StatusUnauthorized)
				return "", "", fmt.Errorf("user not authorized for this account")
			}
		}
	} else if len(cuentas) == 1 {
		cuenta = cuentas[0]
	} else {
		http.Error(w, "", http.StatusUnauthorized)
		return "", "", fmt.Errorf("no accounts available")
	}

	if !auth.IsStudent(app.Production, correo, passwd) {
		http.Error(w, "", http.StatusUnauthorized)
		return "", "", err
	}

	auth.JwtSet(w,
		app.Production,
		"token",
		correo,
		cuenta,
		time.Now().Add(1*time.Hour),
		app.Secret,
	)

	return correo, cuenta, nil
}

func (app *App) log(format string, args ...interface{}) {
	if app.Debug {
		log.Println(fmt.Sprintf(format, args...))
	}
}
