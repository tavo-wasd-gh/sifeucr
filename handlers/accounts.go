package handlers

import (
	"net/http"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/views"

	"sifeucr/config"
	"sifeucr/internal/db"
)

func (h *Handler) AddAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.checkPermissionFromContext(ctx, config.WriteAdvanced)
	if err != nil {
		h.Log().Error("error checking permissions: %v", err)
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	type addAccountForm struct {
		Abbr string `form:"abbr" fmt:"trim,upper" req:"1"`
		Name string `form:"name" req:"1"`
	}

	accountForm, err := forms.FormToStruct[addAccountForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	newAccount := db.NewAccountParams{
		AccountAbbr:   accountForm.Abbr,
		AccountName:   accountForm.Name,
		AccountActive: true,
	}

	queries := db.New(h.DB())
	insertedAccount, err := queries.NewAccount(ctx, newAccount)
	if err != nil {
		h.Log().Error("error adding account: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "account", insertedAccount); err != nil {
		h.Log().Error("failed to render new account: %v", err)
	}
}

func (h *Handler) ToggleAccount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.checkPermissionFromContext(ctx, config.WriteAdvanced)
	if err != nil {
		h.Log().Error("error checking permissions: %v", err)
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	type toggleAccountForm struct {
		ID int64 `form:"account_id" req:"1"`
	}

	toggleForm, err := forms.FormToStruct[toggleAccountForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	queries := db.New(h.DB())
	err = queries.ToggleAccountActiveByAccountID(ctx, toggleForm.ID)
	if err != nil {
		h.Log().Error("error toggling account: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
