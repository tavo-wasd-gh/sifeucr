package main

import (
	"net/http"

	"git.tavo.one/tavo/axiom/middleware"

	"sifeucr/handlers"
)

func routes(handler *handlers.Handler) *http.ServeMux {
	router := http.NewServeMux()

	protectedLax    := middleware.Stack(handler.ValidateSession(false))
	protectedStrict := middleware.Stack(handler.ValidateSession(true))

	router.HandleFunc("GET /", handler.IndexPage)
	router.Handle("GET /cuenta", middleware.With(protectedLax, handler.Dashboard))
	router.HandleFunc("POST /cuenta", handler.LoginForm)
	router.Handle("GET /panel", middleware.With(protectedLax, handler.Panel))
	router.HandleFunc("GET /proveedores", handler.SuppliersPage)
	router.HandleFunc("GET /fse", handler.FSEPage)

	// Presupuesto
	router.Handle("POST /budget/add", middleware.With(protectedStrict, handler.AddBudgetEntry))

	// Usuarios
	router.Handle("POST /user/add",    middleware.With(protectedStrict, handler.AddUser))
	router.Handle("POST /user/toggle", middleware.With(protectedStrict, handler.ToggleUser))

	// Cuentas
	router.Handle("POST /account/add",    middleware.With(protectedStrict, handler.AddAccount))
	router.Handle("POST /account/toggle", middleware.With(protectedStrict, handler.ToggleAccount))

	// TODO: Distribuciones
	// router.Handle("POST /dist/add",    middleware.With(protectedStrict, handler.AddDistribution))
	// router.Handle("POST /dist/toggle", middleware.With(protectedStrict, handler.ToggleDistribution))
	// router.Handle("PUT /dist/update",  middleware.With(protectedStrict, handler.UpdateDistribution))

	// TODO: Proveedores
	// router.Handle("POST /supplier/add",   middleware.With(protectedStrict, handler.AddSupplier))
	// router.Handle("PUT /supplier/update", middleware.With(protectedStrict, handler.UpdateSupplier))
	// router.Handle("POST /catalog/add",    middleware.With(protectedStrict, handler.AddCatalog))
	// router.Handle("PUT /catalog/update",  middleware.With(protectedStrict, handler.UpdateCatalog))

	// TODO: Solicitudes
	//     - Modificaciones Globales
	//     - Modificaciones Internas
	//     - Compras

	return router
}
