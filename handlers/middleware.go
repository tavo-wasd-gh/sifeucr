package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"sifeucr/config"
	"sifeucr/internal/db"
)

func (h *Handler) AuthenticationMiddleware(enforceCSRFProtection bool, requiredPermission int64, redirect string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, accountID, newst, newct, err := h.authenticate(r, enforceCSRFProtection, requiredPermission)
			if err != nil {
				h.Log().Error("failed to authenticate user: %v", err)

				if redirect == "" {
					http.Error(w, "", http.StatusUnauthorized)
					return
				}

				http.Redirect(w, r, redirect, http.StatusFound)
				return
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, config.UserIDKey,    userID)
			ctx = context.WithValue(ctx, config.AccountIDKey, accountID)
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

			w.Header().Set("X-CSRF-Token", newct)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (h *Handler) authenticate(
	r *http.Request,
	enforceCSRFProtection bool,
	requiredPermission int64,
) (int64, int64, string, string, error) {
	c, err := r.Cookie(config.SessionTokenKey)
	oldst := c.Value
	if err != nil || oldst == "" {
		return 0, 0, "", "", fmt.Errorf("invalid or missing session cookie: %v", err)
	}

	var (
		newst, newct string
		session      config.Session
	)

	oldct := r.Header.Get("X-CSRF-Token")
	if oldct == "" {
		if enforceCSRFProtection {
			return 0, 0, "", "", fmt.Errorf("missing CSRF token")
		}

		newst, newct, session, err = h.Sessions().RotateTokens(oldst)
		if err != nil {
			return 0, 0, "", "", fmt.Errorf("error rotating tokens")
		}
	} else {
		newst, newct, session, err = h.Sessions().Validate(oldst, oldct)
		// Having a defined CSRF token but it being invalid is a
		// possible expired session or forged request. CSRF protection
		// is meant for unsafe methods, return immediately.
		if err != nil {
			return 0, 0, "", "", fmt.Errorf("session/csrf validation error: %v", err)
		}
	}

	ctx := r.Context()
	queries := db.New(h.DB())
	perm, err := queries.PermissionByUserIDAndAccountID(ctx, db.PermissionByUserIDAndAccountIDParams{
		UserID:    session.UserID,
		AccountID: session.AccountID,
	})
	if err != nil {
		return 0, 0, "", "", fmt.Errorf("permission lookup failed: %v", err)
	}
	if !config.HasPermission(perm.PermissionInteger, requiredPermission) {
		return 0, 0, "", "", fmt.Errorf("insufficient permissions: got:%d want:%d", perm.PermissionInteger, requiredPermission)
	}

	// DEBUG: Check loading-state indicators
	if !h.Production() {
		time.Sleep(300 * time.Millisecond)
	}

	return session.UserID, session.AccountID, newst, newct, nil
}
