package handlers

import (
	"net/http"
)

func (h *Handler) FirstTimeSetup(w http.ResponseWriter, r *http.Request) {
	// TODO: First time setup, check for h.IsFirstTimeSetup()
}
