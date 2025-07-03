package handlers

import (
	"context"
	"net/http"
	"fmt"

	"git.tavo.one/tavo/axiom/views"
	"github.com/tavo-wasd-gh/sifeucr/config"
	"github.com/tavo-wasd-gh/sifeucr/database"
)

type dashboard struct {
	User      database.User
	Account   database.Account
	CSRFToken string
	Requests  *[]database.Request
	// Advanced
	ReadAdvanced bool
	// mainReport   MainReport
}

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	dashboard, err := h.loadDashboard(
		getUserID(r),
		getAccountID(r),
		getCSRFToken(r),
	)

	if err != nil {
		h.Log().Error("error loading dashboard: %v", err)
		views.RenderHTML(w, r, "login-page", nil)
		return
	}

	if err = views.RenderHTML(w, r, "dashboard-page", dashboard); err != nil {
		h.Log().Error("failed to render dashboard-page: %v", err)
	}
}

func (h *Handler) loadDashboard(
	userID,
	accountID int64,
	csrfToken string,
) (*dashboard, error) {
	if userID == 0 || accountID == 0 || csrfToken == "" {
		return nil, fmt.Errorf("cannot load dashboard: invalid data")
	}

	ctx := context.Background()
	dashboard := dashboard{}
	queries := database.New(h.DB())

	user, err := queries.UserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user by ID: %v", err)
	}
	dashboard.User = user

	account, err := queries.AccountByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query account by ID: %v", err)
	}
	dashboard.Account = account

	requests, err := queries.RequestsByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query requests by accountID: %v", err)
	}
	dashboard.Requests = &requests

	perm, err := queries.GetPermission(ctx, database.GetPermissionParams{
		PermissionUser:    userID,
		PermissionAccount: accountID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %v", err)
	}
	dashboard.ReadAdvanced = false

	if !config.HasPermission(perm.PermissionID, config.ReadAdvanced) {
		dashboard.ReadAdvanced = true
	}

	dashboard.CSRFToken = csrfToken
	return &dashboard, nil
}
