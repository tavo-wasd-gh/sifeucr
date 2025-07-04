package handlers

import (
	"context"
	"net/http"
	"fmt"

	"git.tavo.one/tavo/axiom/views"
	"github.com/tavo-wasd-gh/sifeucr/config"
	"github.com/tavo-wasd-gh/sifeucr/database"
)

type panel struct {
	Users     []database.User
	CSRFToken string
}

func (h *Handler) Panel(w http.ResponseWriter, r *http.Request) {
	panel, err := h.loadPanel(r.Context())

	if err != nil {
		h.Log().Error("error loading panel: %v", err)
		http.Redirect(w, r, "/cuenta", http.StatusFound)
		return
	}

	if err = views.RenderHTML(w, r, "panel-page", panel); err != nil {
		h.Log().Error("failed to render panel-page: %v", err)
	}
}

func (h *Handler) loadPanel(ctx context.Context) (*panel, error) {
	queries := database.New(h.DB())

	userID := getUserIDFromContext(ctx)
	accountID := getAccountIDFromContext(ctx)

	perm, err := queries.GetPermission(ctx, database.GetPermissionParams{
		PermissionUser:    userID,
		PermissionAccount: accountID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %v", err)
	}

	if !config.HasPermission(perm.PermissionInteger, config.ReadAdvanced) {
		return nil, fmt.Errorf("incorrect permissions, got:%d want:%d", perm.PermissionInteger, config.ReadAdvanced)
	}

	csrfToken := getCSRFTokenFromContext(ctx)

	if userID == 0 || accountID == 0 || csrfToken == "" {
		return nil, fmt.Errorf("cannot load dashboard: invalid data")
	}

	panel := panel{}

	panel.Users, err = queries.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query all users: %v", err)
	}

	panel.CSRFToken = csrfToken
	return &panel, nil
}
