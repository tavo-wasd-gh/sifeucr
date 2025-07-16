package handlers

import (
	"net/http"

	"git.tavo.one/tavo/axiom/views"

	"sifeucr/config"
)

func (h *Handler) IndexPage(w http.ResponseWriter, r *http.Request) {
	views.RenderHTML(w, r, "index-page", nil)
}

func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	views.RenderHTML(w, r, "login-page", nil)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(config.SessionTokenKey)
	oldst := c.Value
	if err == nil && oldst != "" {
		h.Sessions().Delete(oldst)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     config.SessionTokenKey,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.Production(),
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	http.Redirect(w, r, "/cuenta", http.StatusFound)
}

func (h *Handler) SuppliersPage(w http.ResponseWriter, r *http.Request) {
	views.RenderHTML(w, r, "suppliers-page", nil)
}

func (h *Handler) FSEPage(w http.ResponseWriter, r *http.Request) {
	views.RenderHTML(w, r, "fse-page", nil)
}
