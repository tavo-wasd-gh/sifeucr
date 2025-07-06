package handlers

import (
	"net/http"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/views"

	"sifeucr/config"
	"sifeucr/internal/db"
)

func (h *Handler) AddBudgetEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.checkPermissionFromContext(ctx, config.WriteAdvanced)
	if err != nil {
		h.Log().Error("error checking permissions: %v", err)
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	type budgetEntryForm struct {
		Year    int64   `form:"year"   req:"1"`
		Code    int64   `form:"code"   req:"1"`
		Object  string  `form:"object" req:"1"`
		Amount  float64 `form:"amount" req:"1"`
	}

	entryForm, err := forms.FormToStruct[budgetEntryForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	newEntry := db.NewBudgetEntryParams{
		EntryYear: entryForm.Year,
		EntryCode: entryForm.Code,
		EntryObject: entryForm.Object,
		EntryAmount: entryForm.Amount,
	}

	queries := db.New(h.DB())
	insertedEntry, err := queries.NewBudgetEntry(ctx, newEntry)
	if err != nil {
		h.Log().Error("error adding budget entry: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "budget", insertedEntry); err != nil {
		h.Log().Error("failed to render new budget: %v", err)
	}
}
