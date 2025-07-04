package main

import (
	"net/http"

	"git.tavo.one/tavo/axiom/middleware"

	"sifeucr/handlers"
)

func routes(handler *handlers.Handler) *http.ServeMux {
	router := http.NewServeMux()

	protectedLax := middleware.Stack(handler.ValidateSession(false))
	// protectedStrict := middleware.Stack(handler.ValidateSession(true))

	router.HandleFunc("GET /", handler.IndexPage)
	router.HandleFunc("POST /cuenta", handler.LoginForm)
	router.HandleFunc("GET /proveedores", handler.SuppliersPage)
	router.HandleFunc("GET /fse", handler.FSEPage)

	router.Handle(
		"GET /cuenta",
		middleware.With(protectedLax, handler.Dashboard),
	)

	router.Handle(
		"GET /panel",
		middleware.With(protectedLax, handler.Panel),
	)

	return router
}
