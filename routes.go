package main

import (
	"net/http"

	"git.tavo.one/tavo/axiom/middleware"

	"sifeucr/config"
	"sifeucr/handlers"
)

func routes(handler *handlers.Handler) *http.ServeMux {
	router := http.NewServeMux()

	// --- NAVBAR ---

	router.HandleFunc("GET /", handler.Static("index-page"))

	router.HandleFunc("GET /proveedores", handler.Static("suppliers-page"))
	router.HandleFunc("GET /fse", handler.Static("fse-page"))

	router.HandleFunc("POST /cuenta", handler.LoginForm)
	router.Handle(
		"GET /cuenta",
		middleware.With(
			middleware.Stack(handler.DashboardMiddleware),
			handler.Dashboard,
		),
	)
	router.HandleFunc("GET /cerrar", handler.Logout)

	// --- SUPPLIERS ---
	router.HandleFunc("POST /proveedores", handler.SendSupplierSummaryToken)
	router.HandleFunc("GET /proveedores/{id}/{token}", handler.LoadSupplierSummary)

	// --- PANEL ---

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
	// Proveedores
	router.Handle("POST /panel/supplier/add", middleware.With(panelMod, handler.AddSupplier))
	router.Handle("PUT /panel/supplier/update/{id}", middleware.With(panelMod, handler.UpdateSupplier))
	router.Handle("POST /panel/catalog/add", middleware.With(panelMod, handler.AddCatalog))
	router.Handle("POST /panel/item/add", middleware.With(panelMod, handler.AddItem))
	router.Handle("PUT /panel/item/update/{id}", middleware.With(panelMod, handler.UpdateItem))

	// Actualizaciones de solicitudes
	router.Handle("PATCH /panel/request/common/{id}", middleware.With(panelMod, handler.PatchRequestCommon))

	// --- FORMS ---

	// Protección de formularios
	getFormStack := middleware.Stack(
		handler.AuthenticationMiddleware(
			false,        // Do not enforce CSRF protection
			config.Write, // Requires Write Permission
			"/cuenta",    // Redirect on error
		),
		handler.PurchaseMiddleware(),
	)
	postFormStack := middleware.Stack(
		handler.AuthenticationMiddleware(
			true,         // Enforce CSRF protection
			config.Write, // Requires Write Permission
			"/cuenta",    // Redirect on error
		),
		handler.PurchaseMiddleware(),
	)

	router.Handle("GET /compra", middleware.With(
		getFormStack,
		handler.PurchaseFormPage,
	))

	router.Handle("POST /request/purchase", middleware.With(
		postFormStack,
		handler.NewPurchase,
	))

	// PROTECTED DOCUMENTS

	protectedPrintStack := middleware.Stack(
		handler.AuthenticationMiddleware(
			false,       // Do not enforce CSRF protection
			config.Read, // Requires Read Permission
			"/cuenta",   // Redirect on error
		),
		handler.ProtectedDocsMiddleware(),
	)

	router.Handle("GET /doc/{type}/{req}", middleware.With(
		protectedPrintStack,
		handler.PrintRequestHandler,
	))

	// SETUP

	router.HandleFunc("GET /panel/setup", handler.FirstTimeSetupPage)
	router.HandleFunc("POST /panel/setup", handler.FirstTimeSetup)

	return router
}
