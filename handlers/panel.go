package handlers

import (
	"context"
	"fmt"
	"net/http"

	"git.tavo.one/tavo/axiom/views"

	"sifeucr/config"
	"sifeucr/internal/db"
)

type panel struct {
	Users         []db.User
	BudgetEntries []db.BudgetEntry
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
	err := h.checkPermissionFromContext(ctx, config.ReadAdvanced)
	if err != nil {
		return nil, fmt.Errorf("error checking permissions: %v", err)
	}

	queries := db.New(h.DB())
	panel := panel{}

	panel.Users, err = queries.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query all users: %v", err)
	}

	panel.BudgetEntries, err = queries.GetAllBudgetEntries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query all budget entries: %v", err)
	}

	csrfToken := getCSRFTokenFromContext(ctx)
	panel.CSRFToken = csrfToken

	return &panel, nil
}
