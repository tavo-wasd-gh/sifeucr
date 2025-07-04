package handlers

import (
	"context"
	"net/http"
	"fmt"

	"git.tavo.one/tavo/axiom/views"
	"sifeucr/config"
	"sifeucr/internal/db"
)

type dashboard struct {
	User      db.User
	Account   db.Account
	CSRFToken string
	Requests  []db.Request
	// Advanced
	ReadAdvanced bool
	// mainReport   MainReport
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
	userID := getUserIDFromContext(ctx)
	accountID := getAccountIDFromContext(ctx)
	csrfToken := getCSRFTokenFromContext(ctx)

	if userID == 0 || accountID == 0 || csrfToken == "" {
		return nil, fmt.Errorf("cannot load dashboard: invalid data")
	}

	dashboard := dashboard{}
	queries := db.New(h.DB())

	user, err := queries.UserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user by ID: %v", err)
	}
	dashboard.User = user

	account, err := queries.AccountByID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query account by ID: %v", err)
	}
	dashboard.Account = account

	requests, err := queries.RequestsByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query requests by accountID: %v", err)
	}
	dashboard.Requests = requests

	perm, err := queries.GetPermission(ctx, db.GetPermissionParams{
		PermissionUser:    userID,
		PermissionAccount: accountID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %v", err)
	}

	if !config.HasPermission(perm.PermissionID, config.Read) {
		return nil, fmt.Errorf("incorrect permissions, want:%d got:%d", config.Read, perm.PermissionID)
	}

	dashboard.ReadAdvanced = config.HasPermission(perm.PermissionInteger, config.ReadAdvanced)
	dashboard.CSRFToken = csrfToken
	return &dashboard, nil
}
