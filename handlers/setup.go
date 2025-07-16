package handlers

import (
	"net/http"
	"strings"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/views"

	"sifeucr/config"
	"sifeucr/internal/db"
)

func (h *Handler) FirstTimeSetupPage(w http.ResponseWriter, r *http.Request) {
	if !h.IsFirstTimeSetup() {
		http.Error(w, "", http.StatusUnauthorized)
		h.Log().Error("invalid try to do first time setup")
		return
	}

	views.RenderHTML(w, r, "setup-page", nil)
}

func (h *Handler) FirstTimeSetup(w http.ResponseWriter, r *http.Request) {
	if !h.IsFirstTimeSetup() {
		http.Error(w, "", http.StatusUnauthorized)
		h.Log().Error("invalid try to do first time setup")
		return
	}

	type setupForm struct {
		UserEmail   string `form:"userEmail" fmt:"trim,lower" validate:"email" req:"1"`
		UserName    string `form:"userName" fmt:"trim" req:"1"`
		AccountName string `form:"accountName" fmt:"trim" req:"1"`
		AccountAbbr string `form:"accountAbbr" fmt:"trim,upper" req:"1"`
	}

	setup, err := forms.FormToStruct[setupForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	trimmedUser := setup.UserEmail
	if strings.Contains(strings.ToLower(trimmedUser), "@ucr.ac.cr") {
		pos := strings.IndexRune(trimmedUser, '@')
		if pos != -1 {
			trimmedUser = trimmedUser[:pos]
		}
	} else {
		h.Log().Error("must be institutional email")
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	firstUser := db.NewUserParams{
		UserEmail:  trimmedUser,
		UserName:   setup.UserName,
		UserActive: true,
	}

	queries := db.New(h.DB())
	ctx := r.Context()

	insertedUser, err := queries.NewUser(ctx, firstUser)
	if err != nil {
		h.Log().Error("error adding user: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	firstAccount := db.AddAccountParams{
		AccountAbbr:   setup.AccountAbbr,
		AccountName:   setup.AccountName,
		AccountActive: true,
	}

	insertedAccount, err := queries.AddAccount(ctx, firstAccount)
	if err != nil {
		h.Log().Error("error adding account: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	firstPerm := db.AddPermissionParams{
		PermissionUser:    insertedUser.UserID,
		PermissionAccount: insertedAccount.AccountID,
		PermissionInteger:
		config.Read |
			config.Write |
			config.ReadOther |
			config.WriteOther |
			config.ReadAdvanced |
			config.WriteAdvanced,
		PermissionActive:  true,
	}

	_, err = queries.AddPermission(ctx, firstPerm)
	if err != nil {
		h.Log().Error("error adding permission: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/cuenta", http.StatusFound)
}
