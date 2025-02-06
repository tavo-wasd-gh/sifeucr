package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
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
	godotenv.Load()

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
		"donacion":        "views/donacion.html",
		"donacion-form":   "views/donacion-form.html",
		"ajuste":          "views/ajuste.html",
		"ajuste-form":     "views/ajuste-form.html",
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
	http.HandleFunc("/api/donaciones", app.handleDonacionForm)
	http.HandleFunc("/api/ajustes", app.handleAjusteForm)

	// Leer Solicitudes
	http.HandleFunc("/api/servicios/", app.handleServicio)
	http.HandleFunc("/api/suministro/", app.handleSuministro)
	http.HandleFunc("/api/bienes/", app.handleBien)
	http.HandleFunc("/api/donaciones/", app.handleDonacion)
	http.HandleFunc("/api/ajustes/", app.handleAjuste)

	// Suscribir a Solicitudes
	http.HandleFunc("/api/suscribir/servicio", app.handleSuscribirServicio)
	http.HandleFunc("/api/suscribir/bien", app.handleSuscribirBien)

	// Marcar como ejecutado/recibido
	http.HandleFunc("/api/ejecutar/servicio", app.handleEjecutarServicio)
	http.HandleFunc("/api/recibir/suministro", app.handleRecibirSuministro)
	http.HandleFunc("/api/recibir/bien", app.handleRecibirBien)

	// Aprobar COES
	http.HandleFunc("/api/aprobar/servicio/", app.handleAprobarServicioCOES)
	http.HandleFunc("/api/aprobar/suministro/", app.handleAprobarSuministroCOES)
	http.HandleFunc("/api/aprobar/bien/", app.handleAprobarBienCOES)
	http.HandleFunc("/api/aprobar/donacion/", app.handleAprobarDonacionCOES)

	// Movimientos
	http.HandleFunc("/api/movimientos/servicio/", app.handleMovimientosServicio)
	http.HandleFunc("/api/movimientos/bien/", app.handleMovimientosBien)

	// Solicitud GECO
	http.HandleFunc("/api/geco/servicio/", app.handleRegistrarSolicitudServicioGECO)
	http.HandleFunc("/api/geco/suministro/", app.handleRegistrarSolicitudSuministroGECO)
	http.HandleFunc("/api/geco/bien/", app.handleRegistrarSolicitudBienGECO)
	// OCS GECO
	http.HandleFunc("/api/orden/servicio/", app.handleServicioOCS)
	http.HandleFunc("/api/orden/bien/", app.handleBienOC)

	// Credenciales
	http.HandleFunc("/api/cuentas/suscriben", app.handleCuentasSuscriben)
	http.HandleFunc("/api/cuentas/donacion", app.handleCuentasDonacion)
	http.HandleFunc("/api/cuentas/ajuste", app.handleCuentasAjuste)

	// ---
	http.HandleFunc("/api/dashboard", app.handleDashboard)
	http.HandleFunc("/logout", logoutHandler)
	http.Handle("/", http.FileServer(http.Dir("public")))

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		// app.log("privilegio SF: %d", auth.SF)
		// app.log("privilegio COES: %d", auth.COES)
		// app.log("privilegio Regular: %d", auth.Regular)
		app.log("starting on :%s...", port)

		if err := http.ListenAndServe(":"+port, nil); err != nil {
			log.Fatalf("fatal: failed to start on port %s: %v", port, err)
		}
	}()

	<-stop

	app.log("shutting down...")
}

func (app *App) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, POST, OPTIONS", "Content-Type", false) {
		return
	}

	if r.Method == http.MethodGet {
		correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
		if err != nil {
			app.log("error validating token: %v", err)
			app.Render(w, "login", nil)
			return
		}

		u, err := database.Login(app.DB, correo, cuenta)
		if err != nil {
			app.log("error logging in: %v", err)
			app.Render(w, "login", nil)
			return
		}

		if err := app.Render(w, "dashboard", u); err != nil {
			app.log("error rendering view: %v", err)
			return
		}

		return
	}

	if r.Method == http.MethodPost {
		correo, cuenta, err := app.ValidateLoginForm(r, w)
		if err != nil {
			app.log("error validating form: %v", err)
			return
		}

		u, err := database.Login(app.DB, correo, cuenta)
		if err != nil {
			app.log("error logging in: %v", err)
			app.Render(w, "login", nil)
			return
		}

		if err := app.Render(w, "dashboard", u); err != nil {
			app.log("error rendering view: %v", err)
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
			app.log("error validating token: %v", err)
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
			app.log("error loading service: %v", err)
			http.Error(w, "failed to load servicio", http.StatusInternalServerError)
			return
		}

		s.UsuarioLoggeado = correo
		s.CuentaLoggeada = cuenta

		if err := app.Render(w, "servicio", s); err != nil {
			app.log("error rendering view: %v", err)
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
			app.log("error validating token: %v", err)
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
			app.log("error loading sum: %v", err)
			http.Error(w, "failed to load sum", http.StatusInternalServerError)
			return
		}

		s.UsuarioLoggeado = correo

		if err := app.Render(w, "suministro", s); err != nil {
			app.log("error rendering view: %v", err)
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
			app.log("error validating token: %v", err)
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
			app.log("error loading bien: %v", err)
			http.Error(w, "failed to load bien", http.StatusInternalServerError)
			return
		}

		b.UsuarioLoggeado = correo
		b.CuentaLoggeada = cuenta

		if err := app.Render(w, "bien", b); err != nil {
			app.log("error rendering view: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}

func (app *App) handleDonacion(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	if r.Method == http.MethodGet {
		correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
		if err != nil {
			app.log("error validating token: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			w.WriteHeader(http.StatusSeeOther)
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

		d, err := database.LeerDonacion(app.DB, id, cuenta)
		if err != nil {
			app.log("error loading donacion: %v", err)
			http.Error(w, "failed to load donacion", http.StatusInternalServerError)
			return
		}

		d.UsuarioLoggeado = correo
		d.CuentaLoggeada = cuenta

		if err := app.Render(w, "donacion", d); err != nil {
			app.log("error rendering view: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}

func (app *App) handleAjuste(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, OPTIONS", "Content-Type", false) {
		return
	}

	if r.Method == http.MethodGet {
		correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
		if err != nil {
			app.log("error validating token: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			w.WriteHeader(http.StatusSeeOther)
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

		ajuste, err := database.LeerAjuste(app.DB, id, cuenta)
		if err != nil {
			app.log("error loading ajuste: %v", err)
			http.Error(w, "failed to load ajuste", http.StatusInternalServerError)
			return
		}

		ajuste.UsuarioLoggeado = correo
		ajuste.CuentaLoggeada = cuenta

		if err := app.Render(w, "ajuste", ajuste); err != nil {
			app.log("error rendering view: %v", err)
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
			app.log("error validating token: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			w.WriteHeader(http.StatusSeeOther)
			return
		}

		u, err := database.Login(app.DB, correo, cuenta)
		if err != nil {
			app.log("error logging in: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		if err := app.Render(w, "servicio-form", u); err != nil {
			app.log("error rendering view: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		return
	}

	if r.Method == http.MethodPost {
		servicio, err := app.validateServicioForm(r, w)
		if err != nil {
			app.log("error validating token: %v", "")
			w.Header().Set("HX-Redirect", "/dashboard")
			w.WriteHeader(http.StatusSeeOther)
			return
		}

		err = database.NuevoServicio(app.DB, servicio)
		if err != nil {
			app.log("error registering service: %v", err)
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
			app.log("error validating token: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		u, err := database.Login(app.DB, correo, cuenta)
		if err != nil {
			app.log("error logging in: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		if err := app.Render(w, "suministro-form", u); err != nil {
			app.log("error rendering view: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		return
	}

	if r.Method == http.MethodPost {
		suministro, err := app.validateSuministroForm(r, w)
		if err != nil {
			app.log("error validating token: %v", "")
			w.Header().Set("HX-Redirect", "/dashboard")
			w.WriteHeader(http.StatusSeeOther)
			return
		}

		err = database.NuevoSuministro(app.DB, suministro)
		if err != nil {
			app.log("error registering sum: %v", err)
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
			app.log("error validating token: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		u, err := database.Login(app.DB, correo, cuenta)
		if err != nil {
			app.log("error logging in: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		if err := app.Render(w, "bien-form", u); err != nil {
			app.log("error rendering view: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		return
	}

	if r.Method == http.MethodPost {
		bien, err := app.validateBienForm(r, w)
		if err != nil {
			app.log("error validating form: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		err = database.NuevoBien(app.DB, bien)
		if err != nil {
			app.log("error registering bien: %v", err)
			http.Error(w, "", http.StatusBadRequest)
			fmt.Fprint(w, `<div class="card-header app-error">Error solicitando bien</div>`)
			return
		}

		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusSeeOther)

		return
	}
}

func (app *App) handleDonacionForm(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, POST, OPTIONS", "Content-Type", false) {
		return
	}

	if r.Method == http.MethodGet {
		correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
		if err != nil {
			app.log("error validating token: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		u, err := database.Login(app.DB, correo, cuenta)
		if err != nil {
			app.log("error logging in: %v", err)
			http.Error(w, "", http.StatusUnauthorized)
			return
		}

		if err := app.Render(w, "donacion-form", u); err != nil {
			app.log("error rendering view: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}

		return
	}

	if r.Method == http.MethodPost {
		donacion, err := app.validateDonacionForm(r, w)
		if err != nil {
			app.log("error validating form: %v", err)
			fmt.Fprint(w, `<div class="card-header app-error">Error solicitando donación</div>`)
			return
		}

		err = database.NuevoDonacion(app.DB, donacion)
		if err != nil {
			app.log("error registering donacion: %v", err)
			fmt.Fprint(w, `<div class="card-header app-error">Error solicitando donación</div>`)
			return
		}

		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusSeeOther)

		return
	}
}

func (app *App) handleAjusteForm(w http.ResponseWriter, r *http.Request) {
	const errResp = `<div class="card-header app-error">Error solicitando ajuste</div>`
	if !cors.Handler(w, r, "*", "GET, POST, OPTIONS", "Content-Type", false) {
		return
	}

	if r.Method == http.MethodGet {
		correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
		if err != nil {
			app.log("error validating token: %v", err)
			w.Header().Set("HX-Redirect", "/dashboard")
			return
		}

		u, err := database.Login(app.DB, correo, cuenta)
		if err != nil {
			app.log("error logging in: %v", err)
			fmt.Fprint(w, errResp)
			return
		}

		if err := app.Render(w, "ajuste-form", u); err != nil {
			app.log("error rendering view: %v", err)
			fmt.Fprint(w, errResp)
			return
		}

		return
	}

	if r.Method == http.MethodPost {
		ajuste, err := app.validateAjusteForm(r, w)
		if err != nil {
			app.log("error validating form: %v", err)
			fmt.Fprint(w, errResp)
			return
		}

		err = database.NuevoAjuste(app.DB, ajuste)
		if err != nil {
			app.log("error registering ajuste: %v", err)
			fmt.Fprint(w, errResp)
			return
		}

		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}
}

func (app *App) validateAjusteForm(r *http.Request, w http.ResponseWriter) (database.Ajuste, error) {
	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		return database.Ajuste{}, err
	}

	if err := r.ParseForm(); err != nil {
		return database.Ajuste{}, err
	}

	if cuenta != "CC" && cuenta != "SF" {
		http.Error(w, "", http.StatusBadRequest)
		return database.Ajuste{}, fmt.Errorf("Solamente CC y SF pueden emitir ajustes")
	}

	emitido := time.Now()
	cuentaEmisora := cuenta
	cuentaDestino := r.Form.Get("cuenta")
	partida := r.Form.Get("partida")
	detalle := r.Form.Get("detalle")
	montoStr := r.Form.Get("monto-bruto")

	if cuentaDestino == "" || partida == "" || detalle == "" || montoStr == "" {
		return database.Ajuste{}, fmt.Errorf("missing required fields")
	}

	montoBruto, err := strconv.ParseFloat(montoStr, 64)
	if err != nil {
		http.Error(w, "Invalid monto format", http.StatusBadRequest)
		return database.Ajuste{}, err
	}

	validPartidas := map[string]bool{
		"servicios":   true,
		"suministros": true,
		"bienes":      true,
	}
	if !validPartidas[partida] {
		http.Error(w, "Invalid partida value", http.StatusBadRequest)
		return database.Ajuste{}, fmt.Errorf("invalid partida value")
	}

	ajuste := database.Ajuste{
		Emitido:       emitido,
		Emisor:        correo,
		CuentaEmisora: cuentaEmisora,
		Cuenta:        cuentaDestino,
		Partida:       partida,
		Detalle:       detalle,
		MontoBruto:    montoBruto,
	}

	return ajuste, nil
}

func (app *App) validateDonacionForm(r *http.Request, w http.ResponseWriter) (database.Donacion, error) {
	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		http.Error(w, "", http.StatusUnauthorized)
		return database.Donacion{}, err
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return database.Donacion{}, err
	}

	emitido := time.Now()
	cuentaSalida := r.Form.Get("cuenta-salida")
	partidaSalida := r.Form.Get("partida-salida")
	cuentaEntrada := r.Form.Get("cuenta-entrada")
	partidaEntrada := r.Form.Get("partida-entrada")
	detalle := r.Form.Get("detalle")
	montoBrutoStr := r.Form.Get("monto-bruto")

	validPartidas := map[string]bool{
		"servicios":   true,
		"suministros": true,
		"bienes":      true,
	}

	if !validPartidas[partidaSalida] || !validPartidas[partidaEntrada] {
		return database.Donacion{}, fmt.Errorf("invalid partida value")
	}

	if cuentaSalida == "" || cuentaEntrada == "" || detalle == "" || montoBrutoStr == "" {
		return database.Donacion{}, fmt.Errorf("missing required fields")
	}

	montoBruto, err := strconv.ParseFloat(montoBrutoStr, 64)
	if err != nil {
		return database.Donacion{}, err
	}

	donacion := database.Donacion{
		Emitido:        emitido,
		Emisor:         correo,
		Cuenta:         cuenta,
		CuentaSalida:   cuentaSalida,
		PartidaSalida:  partidaSalida,
		CuentaEntrada:  cuentaEntrada,
		PartidaEntrada: partidaEntrada,
		Detalle:        detalle,
		MontoBruto:     montoBruto,
	}

	return donacion, nil
}

func (app *App) handleSuscribirServicio(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	err = r.ParseForm()
	idMov := r.Form.Get("id")
	firmaForm := r.Form.Get("firma-suscribir")

	if err != nil || idMov == "" || firmaForm == "" {
		app.log("error parsing form: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la solicitud</div>`)
		return
	}

	err = database.FirmarMovimientoServicios(app.DB, idMov, correo, cuenta, firmaForm)
	if err != nil {
		app.log("error signing service: %v", err)
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
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusSeeOther)
		return
	}

	err = r.ParseForm()
	idMov := r.Form.Get("id")
	firmaForm := r.Form.Get("firma-suscribir")

	if err != nil || idMov == "" || firmaForm == "" {
		app.log("error parsing form: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la recepción</div>`)
		return
	}

	err = database.FirmarMovimientoBienes(app.DB, idMov, correo, cuenta, firmaForm)
	if err != nil {
		app.log("error signing bien: %v", err)
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
		app.log("error validating token: %v", err)
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
		app.log("error parsing form: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la solicitud</div>`)
		return
	}

	fechaEjecutado, err := time.Parse("2006-01-02T15:04", fechaEjecutadoStr)
	if err != nil {
		app.log("error parsing form: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la solicitud</div>`)
		return
	}

	err = database.ConfirmarEjecutadoServicios(app.DB, idServ, correo, cuenta, fechaEjecutado, acuseEjecutado, firmaAcuse)
	if err != nil {
		app.log("error in confirmation: %v", err)
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
		app.log("error validating token: %v", err)
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
		app.log("error parsing form: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la recepción</div>`)
		return
	}

	fechaRecibido, err := time.Parse("2006-01-02T15:04", fechaRecibidoStr)
	if err != nil {
		app.log("error parsing form: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la recepción</div>`)
		return
	}

	err = database.ConfirmarEjecutadoSuministros(app.DB, idSuministro, correo, cuenta, fechaRecibido, acuseRecibido, firmaAcuse)
	if err != nil {
		app.log("error in confirmation: %v", err)
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
		app.log("error validating token: %v", err)
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
		app.log("error parsing form: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error firmando la recepción</div>`)
		return
	}

	fechaRecibido, err := time.Parse("2006-01-02T15:04", fechaRecibidoStr)
	if err != nil {
		app.log("error parsing date: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error en la fecha de recepción</div>`)
		return
	}

	err = database.ConfirmarRecibidoBienes(app.DB, idBien, correo, cuenta, fechaRecibido, acuseRecibido, firmaAcuse)
	if err != nil {
		app.log("error in confirmation: %v", err)
		fmt.Fprint(w, `<div class="card-header app-error">Error confirmando la recepción</div>`)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
	w.WriteHeader(http.StatusSeeOther)
}

func (app *App) handleCuentasSuscriben(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, POST, OPTIONS", "Content-Type", false) {
		return
	}

	_, _, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}

	if r.Method == http.MethodGet {
		cuentas, err := database.ListaCuentas(app.DB)
		if err != nil {
			app.log("error fetching accounts: %v", err)
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

func (app *App) handleCuentasDonacion(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, POST, OPTIONS", "Content-Type", false) {
		return
	}

	_, _, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}

	if r.Method == http.MethodGet {
		cuentas, err := database.ListaCuentas(app.DB)
		if err != nil {
			app.log("error fetching accounts: %v", err)
			http.Error(w, "Error fetching accounts", http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, `<div class="card-header" id="cuenta-entrada-container">
			<select name="cuenta-entrada">`)

		for _, cuenta := range cuentas {
			fmt.Fprintf(w, `<option value="%s">%s</option>`, cuenta.ID, cuenta.Nombre)
		}

		fmt.Fprint(w, `</select></div>`)

		return
	}
}

func (app *App) handleCuentasAjuste(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "GET, POST, OPTIONS", "Content-Type", false) {
		return
	}

	_, _, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		w.WriteHeader(http.StatusSeeOther)
		return
	}

	if r.Method == http.MethodGet {
		cuentas, err := database.ListaCuentas(app.DB)
		if err != nil {
			app.log("error fetching accounts: %v", err)
			http.Error(w, "Error fetching accounts", http.StatusInternalServerError)
			return
		}

		fmt.Fprint(w, `<div class="card-header" id="cuenta-container">
			<select name="cuenta">`)

		for _, cuenta := range cuentas {
			fmt.Fprintf(w, `<option value="%s">%s</option>`, cuenta.ID, cuenta.Nombre)
		}

		fmt.Fprint(w, `</select></div>`)

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

	if !strings.Contains(correo, "@ucr.ac.cr") {
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
		app.log("error validating token: %v", err)
		http.Error(w, "", http.StatusUnauthorized)
		return database.Servicio{}, err
	}

	if err := r.ParseForm(); err != nil {
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
		app.log("error fetching accounts: %v", err)
		http.Error(w, "Error fetching accounts", http.StatusInternalServerError)
		return database.Servicio{}, err
	}

	cuentasStrgs := make([]string, 0, len(cuentas))
	for _, cuenta := range cuentas {
		cuentasStrgs = append(cuentasStrgs, cuenta.ID)
	}

	if err := validateSuscriben(suscriben, cuentasStrgs); err != nil {
		app.log("error fetching accounts: %v", err)
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
		app.log("error validating token: %v", err)
		http.Error(w, "", http.StatusUnauthorized)
		return database.Suministros{}, err
	}

	if err := r.ParseForm(); err != nil {
		app.log("error parsing form: %v", err)
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return database.Suministros{}, err
	}

	justif := r.FormValue("justif")
	firma := r.FormValue("firma")
	nombres := r.Form["nombre[]"]
	articulos := r.Form["articulo[]"]
	agrupaciones := r.Form["agrupacion[]"]
	cantidades := r.Form["cantidad[]"]
	montos := r.Form["monto[]"]

	if justif == "" || firma == "" {
		app.log("missing fields")
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return database.Suministros{}, fmt.Errorf("missing required fields")
	}

	n := len(nombres)
	if n == 0 || len(articulos) != n || len(agrupaciones) != n || len(cantidades) != n || len(montos) != n {
		app.log("mismatched field lengths")
		http.Error(w, "Mismatched field lengths", http.StatusBadRequest)
		return database.Suministros{}, fmt.Errorf("mismatched field lengths")
	}

	var desglose []database.SuministroDesglose
	for i := 0; i < n; i++ {
		cantidad, err := strconv.Atoi(cantidades[i])
		if err != nil {
			app.log("error converting cantidad to int: %v", err)
			http.Error(w, "Invalid cantidad value", http.StatusBadRequest)
			return database.Suministros{}, err
		}

		montoUnitario, err := strconv.ParseFloat(montos[i], 64)
		if err != nil {
			app.log("error converting monto to float: %v", err)
			http.Error(w, "Invalid monto value", http.StatusBadRequest)
			return database.Suministros{}, err
		}

		desglose = append(desglose, database.SuministroDesglose{
			Nombre:        nombres[i],
			Articulo:      articulos[i],
			Agrupacion:    agrupaciones[i],
			Cantidad:      cantidad,
			MontoUnitario: montoUnitario,
		})
	}

	suministro := database.Suministros{
		Emitido:  time.Now(),
		Emisor:   correo,
		Cuenta:   cuenta,
		Justif:   justif,
		Firma:    firma,
		Desglose: desglose,
	}

	return suministro, nil
}

func (app *App) validateBienForm(r *http.Request, w http.ResponseWriter) (database.Bien, error) {
	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
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
		app.log("error fetching accounts: %v", err)
		http.Error(w, "Error fetching accounts", http.StatusInternalServerError)
		return database.Bien{}, err
	}

	cuentasStrgs := make([]string, 0, len(cuentas))
	for _, cuenta := range cuentas {
		cuentasStrgs = append(cuentasStrgs, cuenta.ID)
	}

	if err := validateSuscriben(suscriben, cuentasStrgs); err != nil {
		app.log("error validating accounts: %v", err)
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

func (app *App) handleAprobarServicioCOES(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	_, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}

	if cuenta != "COES" {
		app.log("error approving service: %v", err)
		http.Error(w, "Failed to approve service", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	segments := strings.Split(path, "/")
	if len(segments) < 5 {
		http.Error(w, "ID not found in URL", http.StatusBadRequest)
		return
	}
	id := segments[4]

	err = database.AprobarServicioCOES(app.DB, id)
	if err != nil {
		app.log("error approving service: %v", err)
		http.Error(w, "Failed to approve service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
}

func (app *App) handleAprobarSuministroCOES(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	_, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}

	if cuenta != "COES" {
		app.log("error approving service: %v", err)
		http.Error(w, "Failed to approve service", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	segments := strings.Split(path, "/")
	if len(segments) < 5 {
		http.Error(w, "ID not found in URL", http.StatusBadRequest)
		return
	}
	id := segments[4]

	err = database.AprobarSuministroCOES(app.DB, id)
	if err != nil {
		app.log("error approving service: %v", err)
		http.Error(w, "Failed to approve service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
}
func (app *App) handleAprobarBienCOES(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	_, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}

	if cuenta != "COES" {
		app.log("error approving service: %v", err)
		http.Error(w, "Failed to approve service", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	segments := strings.Split(path, "/")
	if len(segments) < 5 {
		http.Error(w, "ID not found in URL", http.StatusBadRequest)
		return
	}
	id := segments[4]

	err = database.AprobarBienCOES(app.DB, id)
	if err != nil {
		app.log("error approving service: %v", err)
		http.Error(w, "Failed to approve service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
}

func (app *App) handleAprobarDonacionCOES(w http.ResponseWriter, r *http.Request) {
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	_, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}

	if cuenta != "COES" {
		app.log("error approving service: %v", err)
		http.Error(w, "Failed to approve service", http.StatusUnauthorized)
		return
	}

	path := r.URL.Path
	segments := strings.Split(path, "/")
	if len(segments) < 5 {
		http.Error(w, "ID not found in URL", http.StatusBadRequest)
		return
	}
	id := segments[4]

	err = database.AprobarDonacionCOES(app.DB, id)
	if err != nil {
		app.log("error approving service: %v", err)
		http.Error(w, "Failed to approve service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
}

func (app *App) handleRegistrarSolicitudServicioGECO(w http.ResponseWriter, r *http.Request) {
	const errResp = `<div class="card-header app-error">error registrando servicio</div>`
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}

	segments := strings.Split(r.URL.Path, "/")
	if len(segments) < 5 {
		app.log("not enough segments")
		fmt.Fprint(w, errResp)
		return
	}
	id := segments[4]

	if err := r.ParseForm(); err != nil {
		app.log("error parsing form")
		fmt.Fprint(w, errResp)
		return
	}

	solicitudGECO := r.FormValue("solicitud-geco")

	servicio, err := database.ServicioPorID(app.DB, correo, cuenta, id)
	if err != nil {
		app.log("failed to get service: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	err = servicio.RegistrarSolicitudGECO(app.DB, solicitudGECO)
	if err != nil {
		app.log("failed to register service: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
}

func (app *App) handleRegistrarSolicitudSuministroGECO(w http.ResponseWriter, r *http.Request) {
	const errResp = `<div class="app-error">Error registrando suministros</div>`
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		fmt.Fprint(w, errResp)
		return
	}

	segments := strings.Split(r.URL.Path, "/")
	if len(segments) < 5 {
		app.log("not enough segments in URL")
		fmt.Fprint(w, errResp)
		return
	}
	id := segments[4]

	if err := r.ParseForm(); err != nil {
		app.log("error parsing form: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	solicitudGECO := r.FormValue("solicitud-geco")
	montoBrutoTotalStr := r.FormValue("monto-bruto-total")

	montoBrutoTotal, err := strconv.ParseFloat(montoBrutoTotalStr, 64)
	if err != nil {
		app.log("invalid monto bruto total: %v", err)
		fmt.Fprint(w, `<div class="app-error">Monto bruto total inválido</div>`)
		return
	}

	suministro, err := database.SuministroPorID(app.DB, correo, cuenta, id)
	if err != nil {
		app.log("failed to get suministro: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	err = suministro.RegistrarSolicitudGECO(app.DB, solicitudGECO, montoBrutoTotal)
	if err != nil {
		app.log("failed to register solicitud GECO: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
}

func (app *App) handleRegistrarSolicitudBienGECO(w http.ResponseWriter, r *http.Request) {
	const errResp = "Error registrando bien"
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}

	segments := strings.Split(r.URL.Path, "/")
	if len(segments) < 5 {
		app.log("not enough segments")
		fmt.Fprint(w, errResp)
		return
	}
	id := segments[4]

	if err := r.ParseForm(); err != nil {
		app.log("error parsing form")
		fmt.Fprint(w, errResp)
		return
	}

	solicitudGECO := r.FormValue("solicitud-geco")

	bien, err := database.BienPorID(app.DB, correo, cuenta, id)
	if err != nil {
		app.log("failed to get service: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	err = bien.RegistrarSolicitudGECO(app.DB, solicitudGECO)
	if err != nil {
		app.log("failed to register service: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
}

func (app *App) handleServicioOCS(w http.ResponseWriter, r *http.Request) {
	const errResp = "Error registrando OCS"
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}

	segments := strings.Split(r.URL.Path, "/")
	if len(segments) < 5 {
		app.log("not enough segments")
		fmt.Fprint(w, errResp)
		return
	}
	id := segments[4]

	if err := r.ParseForm(); err != nil {
		app.log("error parsing form: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	gecoOCS := r.FormValue("orden-geco")
	provNom := r.FormValue("prov-nom")
	provCed := r.FormValue("prov-ced")
	provDirec := r.FormValue("prov-direc")
	provEmail := r.FormValue("prov-email")
	provTel := r.FormValue("prov-tel")
	provBanco := r.FormValue("prov-banco")
	provIBAN := r.FormValue("prov-iban")
	provJustif := r.FormValue("prov-justif")

	montoBruto, err := strconv.ParseFloat(r.FormValue("prov-monto-bruto"), 64)
	if err != nil {
		app.log("invalid monto bruto: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	montoIVA, err := strconv.ParseFloat(r.FormValue("prov-iva"), 64)
	if err != nil {
		app.log("invalid monto iva: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	montoDesc, _ := strconv.ParseFloat(r.FormValue("prov-monto-desc"), 64)

	servicio, err := database.ServicioPorID(app.DB, correo, cuenta, id)
	if err != nil {
		app.log("failed to get service: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	err = servicio.RegistrarOCS(
		app.DB, gecoOCS, provNom, provCed, provDirec, provEmail, provTel,
		provBanco, provIBAN, provJustif, montoBruto, montoIVA, montoDesc,
	)

	if err != nil {
		app.log("failed to register OCS: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
}

func (app *App) handleBienOC(w http.ResponseWriter, r *http.Request) {
	const errResp = "Error registrando OC"
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}

	segments := strings.Split(r.URL.Path, "/")
	if len(segments) < 5 {
		app.log("not enough segments")
		fmt.Fprint(w, errResp)
		return
	}
	id := segments[4]

	if err := r.ParseForm(); err != nil {
		app.log("error parsing form: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	gecoOCS := r.FormValue("orden-geco")
	provNom := r.FormValue("prov-nom")
	provCed := r.FormValue("prov-ced")
	provDirec := r.FormValue("prov-direc")
	provEmail := r.FormValue("prov-email")
	provTel := r.FormValue("prov-tel")
	provBanco := r.FormValue("prov-banco")
	provIBAN := r.FormValue("prov-iban")
	provJustif := r.FormValue("prov-justif")

	montoBruto, err := strconv.ParseFloat(r.FormValue("prov-monto-bruto"), 64)
	if err != nil {
		app.log("invalid monto bruto: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	montoIVA, err := strconv.ParseFloat(r.FormValue("prov-iva"), 64)
	if err != nil {
		app.log("invalid monto iva: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	montoDesc, _ := strconv.ParseFloat(r.FormValue("prov-monto-desc"), 64)

	bien, err := database.BienPorID(app.DB, correo, cuenta, id)
	if err != nil {
		app.log("failed to get service: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	err = bien.RegistrarOC(
		app.DB, gecoOCS, provNom, provCed, provDirec, provEmail, provTel,
		provBanco, provIBAN, provJustif, montoBruto, montoIVA, montoDesc,
	)

	if err != nil {
		app.log("failed to register OC: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
}

func (app *App) handleMovimientosBien(w http.ResponseWriter, r *http.Request) {
	const errResp = "Error estableciendo montos: Deben sumar el monto bruto total de la solicitud"
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}

	segments := strings.Split(r.URL.Path, "/")
	if len(segments) < 5 {
		app.log("not enough segments")
		fmt.Fprint(w, errResp)
		return
	}
	id := segments[4]

	if err := r.ParseForm(); err != nil {
		app.log("error parsing form")
		fmt.Fprint(w, errResp)
		return
	}

	cuentas := r.Form["cuenta[]"]
	montosStr := r.Form["monto[]"]

	if len(cuentas) != len(montosStr) {
		app.log("mismatch between cuentas and montos length")
		fmt.Fprint(w, errResp)
		return
	}

	montos := make(map[string]float64)
	for i, cuenta := range cuentas {
		monto, err := strconv.ParseFloat(montosStr[i], 64)
		if err != nil {
			app.log("invalid monto value: %v", err)
			fmt.Fprint(w, errResp)
			return
		}
		montos[cuenta] = monto
	}

	bien, err := database.BienPorID(app.DB, correo, cuenta, id)
	if err != nil {
		app.log("failed to get service: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	err = bien.EstablecerMontos(app.DB, montos)
	if err != nil {
		app.log("failed to set mov: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
}

func (app *App) handleMovimientosServicio(w http.ResponseWriter, r *http.Request) {
	const errResp = "Error estableciendo montos: Deben sumar el monto bruto total de la solicitud"
	if !cors.Handler(w, r, "*", "POST, OPTIONS", "Content-Type", false) {
		return
	}

	correo, cuenta, err := auth.JwtValidate(r, "token", app.Secret)
	if err != nil {
		app.log("error validating token: %v", err)
		w.Header().Set("HX-Redirect", "/dashboard")
		return
	}

	segments := strings.Split(r.URL.Path, "/")
	if len(segments) < 5 {
		app.log("not enough segments")
		fmt.Fprint(w, errResp)
		return
	}
	id := segments[4]

	if err := r.ParseForm(); err != nil {
		app.log("error parsing form")
		fmt.Fprint(w, errResp)
		return
	}

	cuentas := r.Form["cuenta[]"]
	montosStr := r.Form["monto[]"]

	if len(cuentas) != len(montosStr) {
		app.log("mismatch between cuentas and montos length")
		fmt.Fprint(w, errResp)
		return
	}

	montos := make(map[string]float64)
	for i, cuenta := range cuentas {
		monto, err := strconv.ParseFloat(montosStr[i], 64)
		if err != nil {
			app.log("invalid monto value: %v", err)
			fmt.Fprint(w, errResp)
			return
		}
		montos[cuenta] = monto
	}

	servicio, err := database.ServicioPorID(app.DB, correo, cuenta, id)
	if err != nil {
		app.log("failed to get service: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	err = servicio.EstablecerMontos(app.DB, montos)
	if err != nil {
		app.log("failed to set mov: %v", err)
		fmt.Fprint(w, errResp)
		return
	}

	w.Header().Set("HX-Redirect", "/dashboard")
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/api",
	})

	http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
}

func (app *App) log(format string, args ...interface{}) {
	if app.Debug {
		pc, _, _, _ := runtime.Caller(1)
		funcName := runtime.FuncForPC(pc).Name()

		log.Printf("%s: %s", funcName, fmt.Sprintf(format, args...))
	}
}
