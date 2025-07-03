package handlers

import (
	"context"
	"net/http"
	"fmt"

	"git.tavo.one/tavo/axiom/views"
	"github.com/tavo-wasd-gh/sifeucr/database"
)

type panel struct {
	Users []database.User
}

func (h *Handler) Panel(w http.ResponseWriter, r *http.Request) {
	panel, err := h.loadPanel()

	if err != nil {
		h.Log().Error("error loading panel: %v", err)
		views.RenderHTML(w, r, "login-page", nil)
		return
	}

	if err = views.RenderHTML(w, r, "panel-page", panel); err != nil {
		h.Log().Error("failed to render panel-page: %v", err)
	}
}

func (h *Handler) loadPanel() (*panel, error) {
	ctx := context.Background()
	panel := panel{}
	queries := database.New(h.DB())
	var err error

	panel.Users, err = queries.GetAllUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to query all users: %v", err)
	}

	return &panel, nil
}
