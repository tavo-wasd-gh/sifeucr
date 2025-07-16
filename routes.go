package main

import (
	"net/http"

	"git.tavo.one/tavo/axiom/middleware"

	"sifeucr/config"
	"sifeucr/handlers"
)

func routes(handler *handlers.Handler) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /", handler.IndexPage)
	router.HandleFunc("GET /proveedores", handler.SuppliersPage)
	router.HandleFunc("GET /fse", handler.FSEPage)
	router.HandleFunc("POST /cuenta", handler.LoginForm)
	router.Handle("GET /cuenta",
		middleware.With(middleware.Stack(handler.DashboardMiddleware),
			handler.Dashboard),
	)
	router.HandleFunc("GET /cerrar", handler.Logout)

	router.HandleFunc("GET /panel/setup", handler.FirstTimeSetupPage)
	router.HandleFunc("POST /panel/setup", handler.FirstTimeSetup)

	// Panel read middleware
	router.Handle(
		"GET /panel",
		middleware.With(middleware.Stack(
			handler.AuthenticationMiddleware(
				false,               // Do not enforce CSRF protection
				config.ReadAdvanced, // Requires ReadAdvanced Permission
				"/cuenta",           // Redirect on error
			),
		), handler.Panel),
	)

	// Panel modification middleware
	panelMod := middleware.Stack(
		handler.AuthenticationMiddleware(
			true,                 // Enforce CSRF protection
			config.WriteAdvanced, // Requires WriteAdvanced Permission
			"",                   // Do not redirect on error
		),
	)

	// TODO: Check active status on all handlers and middlewares where active is relevant
	// Presupuesto
	router.Handle("POST /panel/budget/add", middleware.With(panelMod, handler.AddBudgetEntry))
	// Usuarios
	router.Handle("POST /panel/user/add", middleware.With(panelMod, handler.AddUser))
	router.Handle("POST /panel/user/toggle/{id}", middleware.With(panelMod, handler.ToggleUser))
	// Cuentas
	router.Handle("POST /panel/perm/add", middleware.With(panelMod, handler.AddPermission))
	router.Handle("PUT /panel/perm/toggle/{permName}/{id}", middleware.With(panelMod, handler.TogglePermission))
	// Permisos
	router.Handle("POST /panel/account/add", middleware.With(panelMod, handler.AddAccount))
	router.Handle("POST /panel/account/toggle/{id}", middleware.With(panelMod, handler.ToggleAccount))
	// Periodos
	router.Handle("POST /panel/period/add", middleware.With(panelMod, handler.AddPeriod))
	router.Handle("POST /panel/period/toggle/{id}", middleware.With(panelMod, handler.TogglePeriod))
	router.Handle("PUT /panel/period/update/{id}", middleware.With(panelMod, handler.UpdatePeriod))
	// Distribuciones
	router.Handle("POST /panel/dist/add", middleware.With(panelMod, handler.AddDistribution))
	router.Handle("POST /panel/dist/toggle/{id}", middleware.With(panelMod, handler.ToggleDistribution))
	router.Handle("PUT /panel/dist/update/{id}", middleware.With(panelMod, handler.UpdateDistribution))
	// TODO: Proveedores
	// router.Handle("POST /supplier/add",   middleware.With(panelMod, handler.AddSupplier))
	// router.Handle("PUT /supplier/update", middleware.With(panelMod, handler.UpdateSupplier))
	// router.Handle("POST /catalog/add",    middleware.With(panelMod, handler.AddCatalog))
	// router.Handle("PUT /catalog/update",  middleware.With(panelMod, handler.UpdateCatalog))

	// TODO: Solicitudes
	// Check read/write and readother/writeother permissions depending on the required permissions
	//     - Modificaciones Globales
	//     - Modificaciones Internas
	//     - Compras

	return router
}
