package handlers

import (
	"net/http"

	"git.tavo.one/tavo/axiom/views"
)

func (h *Handler) IndexPage(w http.ResponseWriter, r *http.Request) {
	views.RenderHTML(w, r, "index-page", nil)
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	views.RenderHTML(w, r, "login-page", nil)
}

func (h *Handler) SuppliersPage(w http.ResponseWriter, r *http.Request) {
	views.RenderHTML(w, r, "suppliers-page", nil)
}

func (h *Handler) FSEPage(w http.ResponseWriter, r *http.Request) {
	views.RenderHTML(w, r, "fse-page", nil)
}
