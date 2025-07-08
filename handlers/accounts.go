package handlers

import (
	"net/http"
	"strconv"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/views"

	"sifeucr/internal/db"
)

func (h *Handler) AddAccount(w http.ResponseWriter, r *http.Request) {
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

	newAccount := db.AddAccountParams{
		AccountAbbr:   accountForm.Abbr,
		AccountName:   accountForm.Name,
		AccountActive: true,
	}

	queries := db.New(h.DB())
	insertedAccount, err := queries.AddAccount(r.Context(), newAccount)
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
	accountIDStr := r.PathValue("id")
	accountID, err := strconv.ParseInt(accountIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error toggling account: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	queries := db.New(h.DB())
	err = queries.ToggleAccountActiveByAccountID(r.Context(), accountID)
	if err != nil {
		h.Log().Error("error toggling account: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
