package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"strconv"
	"syscall"
	"time"

	"github.com/tavo-wasd-gh/gocors"
	"github.com/tavo-wasd-gh/sifeucr/auth"
	"github.com/tavo-wasd-gh/sifeucr/database"
	"github.com/tavo-wasd-gh/sifeucr/views"
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
		"login":           "views/login.html",
		"dashboard":       "views/dashboard.html",
		"servicio":        "views/servicio.html",
		"servicio-form":   "views/servicio-form.html",
		"suministro":      "views/suministro.html",
		"suministro-form": "views/suministro-form.html",
		"bien":            "views/bien.html",
		"bien-form":       "views/bien-form.html",
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

	// Hacer Solicitudes
	http.HandleFunc("/api/servicios", app.handleServicioForm)
	http.HandleFunc("/api/suministros", app.handleSuministroForm)
	http.HandleFunc("/api/bienes", app.handleBienForm)

	// Leer Solicitudes
	http.HandleFunc("/api/servicios/", app.handleServicio)
	http.HandleFunc("/api/suministro/", app.handleSuministro)
	http.HandleFunc("/api/bienes/", app.handleBien)

	// Suscribir a Solicitudes
	http.HandleFunc("/api/suscribir/servicio", app.handleSuscribirServicio)
	http.HandleFunc("/api/suscribir/bien", app.handleSuscribirBien)

	// Marcar como ejecutado/recibido
	http.HandleFunc("/api/ejecutar/servicio", app.handleEjecutarServicio)
	http.HandleFunc("/api/recibir/suministro", app.handleRecibirSuministro)
	http.HandleFunc("/api/recibir/bien", app.handleRecibirBien)

	// Credenciales
	http.HandleFunc("/api/cuentas", app.handleCuentas)

	// ---
	http.HandleFunc("/api/dashboard", app.handleDashboard)
	http.HandleFunc("/logout", logoutHandler)
	http.Handle("/", http.FileServer(http.Dir("public")))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// app.log("main: privilegio SF: %d", auth.SF)
		// app.log("main: privilegio COES: %d", auth.COES)
		// app.log("main: privilegio Regular: %d", auth.Regular)
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

		if err := app.Render(w, "dashboard", u); err != nil {
			app.log("handleDashboard: error rendering view: %v", err)
			return
		}

		return
	}

	if r.Method == http.MethodPost {
		correo, cuenta, err := app.ValidateLoginForm(r, w)
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

		if err := app.Render(w, "dashboard", u); err != nil {
			app.log("handleDashboard: error rendering view: %v", err)
		}

		return
	}
}

func (app *App) handleServicio(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	if r.Method == http.MethodGet {
		correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
		if err != nil {
			app.log("handleServicios: error validating token: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		var id string

		path := r.URL.Path
		segments := strings.Split(path, "/")

		if len(segments) <= 2 {
			http.Error(w, "ID not found in URL", http.StatusBadRequest)
			return
		}

		id = segments[3]

		s, err := database.LeerServicio(app.DB, id, cuenta)
		if err != nil {
			app.log("handleServicios: error loading service: %v", err)
			http.Error(w, "failed to load servicio", http.StatusInternalServerError)
			return
		}

		s.UsuarioLoggeado = correo
		s.CuentaLoggeada = cuenta

		if err := app.Render(w, "servicio", s); err != nil {
			app.log("handleServicios: error rendering view: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		return
	}

	return
}

func (app *App) handleSuministro(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	if r.Method == http.MethodGet {
		correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
		if err != nil {
			app.log("handleSuministros: error validating token: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		var id string

		path := r.URL.Path
		segments := strings.Split(path, "/")

		if len(segments) <= 2 {
			http.Error(w, "ID not found in URL", http.StatusBadRequest)
			return
		}

		id = segments[3]

		s, err := database.LeerSuministro(app.DB, id, cuenta)
		if err != nil {
			app.log("handleSuministros: error loading sum: %v", err)
			http.Error(w, "failed to load sum", http.StatusInternalServerError)
			return
		}

		s.UsuarioLoggeado = correo

		if err := app.Render(w, "suministro", s); err != nil {
			app.log("handleSuministros: error rendering view: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		return
	}

	return
}

func (app *App) handleBien(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	if r.Method == http.MethodGet {
		correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
		if err != nil {
			app.log("handleBien: error validating token: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		var id string
		path := r.URL.Path
		segments := strings.Split(path, "/")
		if len(segments) < 4 {
			http.Error(w, "ID not found in URL", http.StatusBadRequest)
			return
		}
		id = segments[3]

		b, err := database.LeerBien(app.DB, id, cuenta)
		if err != nil {
			app.log("handleBien: error loading bien: %v", err)
			http.Error(w, "failed to load bien", http.StatusInternalServerError)
			return
		}

		b.UsuarioLoggeado = correo
		b.CuentaLoggeada = cuenta

		if err := app.Render(w, "bien", b); err != nil {
			app.log("handleBien: error rendering view: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}

func (app *App) handleServicioForm(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, POST, OPTIONS", "Content-Type", false) {
		return
	}

	if r.Method == http.MethodGet {
		correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
		if err != nil {
			app.log("handleServicioForm: error validating token: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		u, err := database.Login(app.DB, correo, cuenta)
		if err != nil {
			app.log("handleServicioForm: error logging in: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		if err := app.Render(w, "servicio-form", u); err != nil {
			app.log("handleServicioForm: error rendering view: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		return
	}

	if r.Method == http.MethodPost {
		servicio, err := app.validateServicioForm(r, w)
		if err != nil {
			app.log("handleServicioForm: error validating token: %v", "")

			w.Header().Set("HX-Redirect", "/dashboard")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		err = database.NuevoServicio(app.DB, servicio)
		if err != nil {
			app.log("handleServicioForm: error registering service: %v", err)
			http.Error(w, "", http.StatusBadRequest)
			fmt.Fprint(w, `<div class="card-header app-error">Error solicitando servicio</div>`)
			return
		}

		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusSeeOther)

		return
	}
}

func (app *App) handleSuministroForm(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, POST, OPTIONS", "Content-Type", false) {
		return
	}

	if r.Method == http.MethodGet {
		correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
		if err != nil {
			app.log("handleSuministroForm: error validating token: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		u, err := database.Login(app.DB, correo, cuenta)
		if err != nil {
			app.log("handleSuministroForm: error logging in: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		if err := app.Render(w, "suministro-form", u); err != nil {
			app.log("handleSuministroForm: error rendering view: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		return
	}

	if r.Method == http.MethodPost {
		suministro, err := app.validateSuministroForm(r, w)
		if err != nil {
			app.log("handleSuministroForm: error validating token: %v", "")

			w.Header().Set("HX-Redirect", "/dashboard")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		err = database.NuevoSuministro(app.DB, suministro)
		if err != nil {
			app.log("handleSuministroForm: error registering sum: %v", err)
			http.Error(w, "", http.StatusBadRequest)
			fmt.Fprint(w, `<div class="card-header app-error">Error solicitando servicio</div>`)
			return
		}

		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusSeeOther)

		return
	}
}

func (app *App) handleBienForm(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, POST, OPTIONS", "Content-Type", false) {
		return
	}

	if r.Method == http.MethodGet {
		correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
		if err != nil {
			app.log("handleBienForm: error validating token: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		u, err := database.Login(app.DB, correo, cuenta)
		if err != nil {
			app.log("handleBienForm: error logging in: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		if err := app.Render(w, "bien-form", u); err != nil {
			app.log("handleBienForm: error rendering view: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		return
	}

	if r.Method == http.MethodPost {
		bien, err := app.validateBienForm(r, w)
		if err != nil {
			app.log("handleBienForm: error validating form: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		err = database.NuevoBien(app.DB, bien)
		if err != nil {
			app.log("handleBienForm: error registering bien: %v", err)
			http.Error(w, "", http.StatusBadRequest)
			fmt.Fprint(w, `<div class="card-header app-error">Error solicitando bien</div>`)
			return
		}

		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusSeeOther)

		return
	}
}

func (app *App) handleSuscribirServicio(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("handleSuscribirServicio: error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	err = r.ParseForm()
	idMov := r.Form.Get("id")
	firmaForm := r.Form.Get("firma-suscribir")

	if err != nil || idMov == "" || firmaForm == "" {
		app.log("handleSuscribirServicio: error parsing form: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la solicitud</div>`)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	err = database.FirmarMovimientoServicios(app.DB, idMov, correo, cuenta, firmaForm)
	if err != nil {
		app.log("handleSuscribirServicio: error signing service: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la solicitud</div>`)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleSuscribirBien(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("handleSuscribirBien: error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	err = r.ParseForm()
	idMov := r.Form.Get("id")
	firmaForm := r.Form.Get("firma-suscribir")

	if err != nil || idMov == "" || firmaForm == "" {
		app.log("handleSuscribirBien: error parsing form: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la recepción</div>`)
		return
	}

	err = database.FirmarMovimientoBienes(app.DB, idMov, correo, cuenta, firmaForm)
	if err != nil {
		app.log("handleSuscribirBien: error signing bien: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la recepción</div>`)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleEjecutarServicio(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("handleSuscribirServicio: error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	err = r.ParseForm()
	idServ := r.Form.Get("id")
	fechaEjecutadoStr := r.Form.Get("fecha-ejecutado")
	acuseEjecutado := r.Form.Get("acuse-ejecutado")
	firmaAcuse := r.Form.Get("firma-acuse")

	if err != nil ||
	idServ == "" ||
	fechaEjecutadoStr == "" ||
	acuseEjecutado == "" ||
	firmaAcuse == "" {
		app.log("handleSuscribirServicio: error parsing form: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la solicitud</div>`)
		return
	}

	fechaEjecutado, err := time.Parse("2006-01-02T15:04", fechaEjecutadoStr)
	if err != nil {
		app.log("handleEjecutarServicio: error parsing form: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la solicitud</div>`)
		return
	}

	err = database.ConfirmarEjecutadoServicios(app.DB, idServ, correo, cuenta, fechaEjecutado, acuseEjecutado, firmaAcuse)
	if err != nil {
		app.log("handleEjecutarServicio: error in confirmation: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error confirmando la solicitud</div>`)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleRecibirSuministro(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("handleRecibirSuministro: error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	err = r.ParseForm()
	idSuministro := r.Form.Get("id")
	fechaRecibidoStr := r.Form.Get("fecha-recibido")
	acuseRecibido := r.Form.Get("acuse-recibido")
	firmaAcuse := r.Form.Get("firma-acuse")

	if err != nil ||
		idSuministro == "" ||
		fechaRecibidoStr == "" ||
		acuseRecibido == "" ||
		firmaAcuse == "" {
		app.log("handleRecibirSuministro: error parsing form: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la recepción</div>`)
		return
	}

	fechaRecibido, err := time.Parse("2006-01-02T15:04", fechaRecibidoStr)
	if err != nil {
		app.log("handleRecibirSuministro: error parsing form: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la recepción</div>`)
		return
	}

	err = database.ConfirmarEjecutadoSuministros(app.DB, idSuministro, correo, cuenta, fechaRecibido, acuseRecibido, firmaAcuse)
	if err != nil {
		app.log("handleRecibirSuministro: error in confirmation: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error confirmando la recepción</div>`)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleRecibirBien(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("handleRecibirBien: error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	err = r.ParseForm()
	idBien := r.Form.Get("id")
	fechaRecibidoStr := r.Form.Get("fecha-recibido")
	acuseRecibido := r.Form.Get("acuse-recibido")
	firmaAcuse := r.Form.Get("firma-acuse")

	if err != nil ||
		idBien == "" ||
		fechaRecibidoStr == "" ||
		acuseRecibido == "" ||
		firmaAcuse == "" {
		app.log("handleRecibirBien: error parsing form: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la recepción</div>`)
		return
	}

	fechaRecibido, err := time.Parse("2006-01-02T15:04", fechaRecibidoStr)
	if err != nil {
		app.log("handleRecibirBien: error parsing date: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error en la fecha de recepción</div>`)
		return
	}

	err = database.ConfirmarRecibidoBienes(app.DB, idBien, correo, cuenta, fechaRecibido, acuseRecibido, firmaAcuse)
	if err != nil {
		app.log("handleRecibirBien: error in confirmation: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error confirmando la recepción</div>`)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleCuentas(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, POST, OPTIONS", "Content-Type", false) {
		return
	}

	_, _, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("handleSuscribe: error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if r.Method == http.MethodGet {
		cuentas, err := database.ListaCuentas(app.DB)
		if err != nil {
			app.log("handleSuscribe: error fetching accounts: %v", err)
			http.Error(w, "Error fetching accounts", http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, `<div class="select-group card-header">
			<select name="suscriben">`)

		for _, cuenta := range cuentas {
			fmt.Fprintf(w, `<option value="%s">%s</option>`, cuenta.ID, cuenta.Nombre)
		}

		fmt.Fprint(w, `</select>
			<button type="button" style="margin:0.5em;" hx-on:click="this.closest('.select-group').remove()">Quitar</button>
			</div>`)
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

func (app *App) ValidateLoginForm(r *http.Request, w http.ResponseWriter) (string, string, error) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return "", "", err
	}

	cuenta := ""
	correo := r.FormValue("correo")
	passwd := r.FormValue("passwd")
	cuentaPedida := r.FormValue("cuenta")

	if !strings.Contains(correo, "@") {
		correo = correo + "@ucr.ac.cr"
	}

	cuentas, err := database.CuentasActivas(app.DB, correo)
	if err != nil {
		http.Error(w, "", http.StatusUnauthorized)
		return "", "", err
	}

	if len(cuentas) > 1 {
		found := false
		if cuentaPedida == "" {
			// TODO
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

func (app *App) validateServicioForm(r *http.Request, w http.ResponseWriter) (database.Servicio, error) {
	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("validateServicioForm: error validating token: %v", err)
		http.Error(w, "", http.StatusUnauthorized)
		return database.Servicio{}, err
	}
	
	if err := r.ParseForm() ; err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return database.Servicio{}, err
	}

	// Servicio
	emitido := time.Now()
	emisor := correo
	detalle := r.Form.Get("detalle")
	porEjecutarStr := r.Form.Get("por-ejecutar")
	justif := r.Form.Get("justif")

	// Movimiento
	firma := r.Form.Get("firma")
	suscriben := r.Form["suscriben"]

	if porEjecutarStr == "" || detalle == "" || justif == "" || firma == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return database.Servicio{}, err
	}

	cuentaSuscribe := false
	for _, cuentaID := range suscriben {
		if cuentaID == cuenta {
			cuentaSuscribe = true
		}
	}
	if !cuentaSuscribe {
		http.Error(w, "Missing default cuenta", http.StatusBadRequest)
		return database.Servicio{}, err
	}

	porEjecutar, err := time.Parse("2006-01-02T15:04", porEjecutarStr)
	if err != nil {
		http.Error(w, "Invalid datetime", http.StatusBadRequest)
		return database.Servicio{}, err
	}

	var s database.Servicio

	s.Emitido = emitido
	s.Emisor = emisor
	s.Detalle = detalle
	s.PorEjecutar = porEjecutar
	s.Justif = justif

	var mm []database.ServicioMovimiento

	cuentas, err := database.ListaCuentas(app.DB)
	if err != nil {
		app.log("handleServicioForm: error fetching accounts: %v", err)
		http.Error(w, "Error fetching accounts", http.StatusInternalServerError)
		return database.Servicio{}, err
	}

	cuentasStrgs := make([]string, 0, len(cuentas))
	for _, cuenta := range cuentas {
		cuentasStrgs = append(cuentasStrgs, cuenta.ID)
	}

	if err := validateSuscriben(suscriben, cuentasStrgs) ; err != nil {
		app.log("handleServicioForm: error fetching accounts: %v", err)
		http.Error(w, "Non-existent account", http.StatusUnauthorized)
		return database.Servicio{}, err
	}
	
	for _, cuentaSuscrita := range suscriben {
		var m database.ServicioMovimiento

		m.Cuenta = cuentaSuscrita

		if m.Cuenta == cuenta {
			m.Usuario = emisor
			m.Firma = firma
		}

		mm = append(mm, m)
	}

	s.Movimientos = mm

	return s, err
}

func (app *App) validateSuministroForm(r *http.Request, w http.ResponseWriter) (database.Suministros, error) {
	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("validateServicioForm: error validating token: %v", err)
		http.Error(w, "", http.StatusUnauthorized)
		return database.Suministros{}, err
	}

	if err := r.ParseForm(); err != nil {
		app.log("validateSuministroForm: error parsing form: %v", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return database.Suministros{}, err
	}

	justif := r.FormValue("justif")
	firma := r.FormValue("firma")
	nombres := r.Form["nombre[]"]
	articulos := r.Form["articulo[]"]
	agrupaciones := r.Form["agrupacion[]"]
	cantidades := r.Form["cantidad[]"]

	if justif == "" || firma == "" {
		app.log("validateSuministroForm: missing fields")
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return database.Suministros{}, err
	}

	if len(nombres) == 0 || len(articulos) == 0 || len(agrupaciones) == 0 || len(cantidades) == 0 {
		app.log("validateSuministroForm: missing fields")
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return database.Suministros{}, err
	}

	var desglose []database.SuministroDesglose
	for i := range nombres {
		cantidad, err := strconv.Atoi(cantidades[i])
		if err != nil {
			app.log("validateSuministroForm: error converting cantidad to int: %v", err)
			http.Error(w, "Invalid cantidad value", http.StatusBadRequest)
			return database.Suministros{}, err
		}

		desglose = append(desglose, database.SuministroDesglose{
			Nombre:     nombres[i],
			Articulo:   articulos[i],
			Agrupacion: agrupaciones[i],
			Cantidad:   cantidad,
		})
	}

	suministro := database.Suministros{
		Emitido:         time.Now(),
		Emisor:          correo,
		Cuenta:          cuenta,
		Justif:          justif,
		Firma:           firma,
		Desglose:        desglose,
	}

	return suministro, nil
}

func (app *App) validateBienForm(r *http.Request, w http.ResponseWriter) (database.Bien, error) {
	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("validateBienForm: error validating token: %v", err)
		http.Error(w, "", http.StatusUnauthorized)
		return database.Bien{}, err
	}
	
	if err := r.ParseForm(); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return database.Bien{}, err
	}

	emitido := time.Now()
	emisor := correo
	detalle := r.Form.Get("detalle")
	porRecibirStr := r.Form.Get("por-recibir")
	justif := r.Form.Get("justif")

	firma := r.Form.Get("firma")
	suscriben := r.Form["suscriben"]

	if porRecibirStr == "" || detalle == "" || justif == "" || firma == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return database.Bien{}, fmt.Errorf("missing required fields")
	}

	cuentaSuscribe := false
	for _, cuentaID := range suscriben {
		if cuentaID == cuenta {
			cuentaSuscribe = true
		}
	}
	if !cuentaSuscribe {
		http.Error(w, "Missing default cuenta", http.StatusBadRequest)
		return database.Bien{}, fmt.Errorf("missing default cuenta")
	}

	porRecibir, err := time.Parse("2006-01-02T15:04", porRecibirStr)
	if err != nil {
		http.Error(w, "Invalid datetime", http.StatusBadRequest)
		return database.Bien{}, err
	}

	var b database.Bien

	b.Emitido = emitido
	b.Emisor = emisor
	b.Detalle = detalle
	b.PorRecibir = porRecibir
	b.Justif = justif

	var mm []database.BienMovimiento

	cuentas, err := database.ListaCuentas(app.DB)
	if err != nil {
		app.log("validateBienForm: error fetching accounts: %v", err)
		http.Error(w, "Error fetching accounts", http.StatusInternalServerError)
		return database.Bien{}, err
	}

	cuentasStrgs := make([]string, 0, len(cuentas))
	for _, cuenta := range cuentas {
		cuentasStrgs = append(cuentasStrgs, cuenta.ID)
	}

	if err := validateSuscriben(suscriben, cuentasStrgs); err != nil {
		app.log("validateBienForm: error validating accounts: %v", err)
		http.Error(w, "Non-existent account", http.StatusUnauthorized)
		return database.Bien{}, err
	}
	
	for _, cuentaSuscrita := range suscriben {
		var m database.BienMovimiento

		m.Cuenta = cuentaSuscrita

		if m.Cuenta == cuenta {
			m.Usuario = emisor
			m.Firma = firma
		}

		mm = append(mm, m)
	}

	b.Movimientos = mm

	return b, err
}

func validateSuscriben(suscriben []string, cuentas []string) error {
	cuentaMap := make(map[string]bool)
	for _, cuenta := range cuentas {
		cuentaMap[cuenta] = true
	}

	for _, s := range suscriben {
		if !cuentaMap[s] {
			return fmt.Errorf("invalid suscriben: %s not found in cuentas", s)
		}
	}

	return nil
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   "",
		Expires: time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:    "/api",
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (app *App) log(format string, args ...interface{}) {
	if app.Debug {
		log.Println(fmt.Sprintf(format, args...))
	}
}
