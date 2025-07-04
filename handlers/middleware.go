package handlers

import (
	"context"
	"net/http"

	"sifeucr/config"
)

func (h *Handler) ValidateSession(strict bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			csrfTokenKey := string(config.CSRFTokenKey)

			c, err := r.Cookie(config.SessionTokenKey)
			if err != nil {
				h.Log().Error("failed to get session cookie: %v", err)
				next.ServeHTTP(w, r)
				return
			}

			oldst := c.Value
			if oldst == "" {
				h.Log().Error("failed to get session token")
				next.ServeHTTP(w, r)
				return
			}

			var newst, newct string
			var session config.Session

			oldct := r.Header.Get(csrfTokenKey)
			if oldct == "" {
				// Unset CSRF token
				// if strict, fail, if relaxed, rotate session
				// (Still needs valid session token)
				h.Log().Error("failed to get csrf token")
				if strict {
					http.Error(w, "", http.StatusUnauthorized)
					return
				}
				newst, newct, session, err = h.Sessions().RotateTokens(oldst)

			} else {
				// Present CSRF token
				// Validate session
				// if fail && strict, fail
				// if fail && relaxed, continue without tokens
				newst, newct, session, err = h.Sessions().Validate(oldst, oldct)
				if err != nil {
					h.Log().Error("failed to validate session: %v", err)
					if strict {
						http.Error(w, "", http.StatusUnauthorized)
						return
					}
					next.ServeHTTP(w, r)
					return
				}
			}

			ctx = context.WithValue(ctx, config.UserIDKey, session.UserID)
			ctx = context.WithValue(ctx, config.AccountIDKey, session.AccountID)
			ctx = context.WithValue(ctx, config.CSRFTokenKey, newct)

			http.SetCookie(w, &http.Cookie{
				Name:     config.SessionTokenKey,
				Value:    newst,
				Path:     "/",
				HttpOnly: true,
				Secure:   h.Production(),
				SameSite: http.SameSiteLaxMode,
				MaxAge:   config.CookieMaxAge,
			})

			w.Header().Set(csrfTokenKey, newct)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
