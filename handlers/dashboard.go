package handlers

import (
	"context"
	"fmt"
	"net/http"

	"git.tavo.one/tavo/axiom/views"

	"sifeucr/config"
	"sifeucr/internal/db"
)

type dashboard struct {
	User      db.User
	Account   db.Account
	CSRFToken string
	Purchases []db.Purchase
	// Advanced
	// AllRequests  []db.AllRequestsRow
	ReadAdvanced bool
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	dashboard, err := h.loadDashboard(r.Context())

	if err != nil {
		h.Log().Error("error loading dashboard: %v", err)
		views.RenderHTML(w, r, "login-page", nil)
		return
	}

	if err = views.RenderHTML(w, r, "dashboard-page", dashboard); err != nil {
		h.Log().Error("failed to render dashboard-page: %v", err)
	}
}

func (h *Handler) loadDashboard(ctx context.Context) (*dashboard, error) {
	err := h.checkPermissionFromContext(ctx, config.Read)
	if err != nil {
		return nil, fmt.Errorf("cannot load dashboard: invalid data")
	}

	userID := getUserIDFromContext(ctx)
	accountID := getAccountIDFromContext(ctx)

	queries := db.New(h.DB())
	dashboard := dashboard{}

	perm, err := queries.PermissionByUserIDAndAccountID(ctx, db.PermissionByUserIDAndAccountIDParams{
		UserID:    userID,
		AccountID: accountID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %v", err)
	}
	dashboard.ReadAdvanced = config.HasPermission(perm.PermissionInteger, config.ReadAdvanced)

	dashboard.User, err = queries.UserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user by ID: %v", err)
	}

	dashboard.Account, err = queries.AccountByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query account by ID: %v", err)
	}

	csrfToken := getCSRFTokenFromContext(ctx)
	dashboard.CSRFToken = csrfToken

	return &dashboard, nil
}

func (h *Handler) DashboardMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, accountID, newst, newct, err := h.authenticate(r, false, config.Read)
		if err != nil {
			h.Log().Error("failed to authenticate user: %v", err)
			views.RenderHTML(w, r, "login-page", nil)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, config.UserIDKey, userID)
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
