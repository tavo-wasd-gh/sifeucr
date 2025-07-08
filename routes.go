package main

import (
	"net/http"

	"git.tavo.one/tavo/axiom/middleware"

	"sifeucr/config"
	"sifeucr/handlers"
)

func routes(handler *handlers.Handler) *http.ServeMux {
	router := http.NewServeMux()

	router.HandleFunc("GET /",            handler.IndexPage)
	router.HandleFunc("GET /proveedores", handler.SuppliersPage)
	router.HandleFunc("GET /fse",         handler.FSEPage)
	router.HandleFunc("POST /cuenta",     handler.LoginForm)
	router.Handle("GET /cuenta",
		middleware.With(middleware.Stack(handler.DashboardMiddleware),
		handler.Dashboard),
	)

	// Panel read middleware, requires:
	// - Disabled CSRF protection
	// - ReadAdvanced permission
	// - Redirect to /cuenta on error
	router.Handle(
		"GET /panel",
		middleware.With(middleware.Stack(
			handler.AuthenticationMiddleware(
				false,               // Enforce CSRF protection
				config.ReadAdvanced, // Required permission
				"/cuenta",           // Redirect on error
			),
		), handler.Panel),
	)

	// Panel modification middleware, requires:
	// - CSRF protection
	// - WriteAdvanced permission
	// - No redirect on error
	panelMod := middleware.Stack(
		handler.AuthenticationMiddleware(
			true,                 // Enforce CSRF protection
			config.WriteAdvanced, // Required permission
			"",                   // Redirect on error
		),
	)

	// Presupuesto
	router.Handle("POST /panel/budget/add", middleware.With(panelMod, handler.AddBudgetEntry))
	// Usuarios
	router.Handle("POST /panel/user/add",         middleware.With(panelMod, handler.AddUser))
	router.Handle("POST /panel/user/toggle/{id}", middleware.With(panelMod, handler.ToggleUser))
	// Cuentas
	router.Handle("POST /panel/account/add",         middleware.With(panelMod, handler.AddAccount))
	router.Handle("POST /panel/account/toggle/{id}", middleware.With(panelMod, handler.ToggleAccount))
	// Distribuciones
	router.Handle("POST /panel/dist/add",         middleware.With(panelMod, handler.AddDistribution))
	router.Handle("POST /panel/dist/toggle/{id}", middleware.With(panelMod, handler.ToggleDistribution))
	router.Handle( "PUT /panel/dist/update/{id}",  middleware.With(panelMod, handler.UpdateDistribution))
	// TODO: Proveedores
	// router.Handle("POST /supplier/add",   middleware.With(panelMod, handler.AddSupplier))
	// router.Handle("PUT /supplier/update", middleware.With(panelMod, handler.UpdateSupplier))
	// router.Handle("POST /catalog/add",    middleware.With(panelMod, handler.AddCatalog))
	// router.Handle("PUT /catalog/update",  middleware.With(panelMod, handler.UpdateCatalog))
	// TODO: Solicitudes
	//     - Modificaciones Globales
	//     - Modificaciones Internas
	//     - Compras

	return router
}
