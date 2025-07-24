package handlers

import (
	"net/http"

	"git.tavo.one/tavo/axiom/views"

	"sifeucr/config"
)

func (h *Handler) Static(key string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := views.RenderHTML(w, r, key, nil); err != nil {
			h.Log().Error("error rendering static page: %v", err)
		}
	}
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
