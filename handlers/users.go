package handlers

import (
	"net/http"
	"strconv"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/views"

	"sifeucr/config"
	"sifeucr/internal/db"
)

func (h *Handler) AddUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.checkPermissionFromContext(ctx, config.WriteAdvanced)
	if err != nil {
		h.Log().Error("error checking permissions: %v", err)
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	type addUserForm struct {
		Email string `form:"email" fmt:"trim,lower" validate:"email" req:"1"`
		Name  string `form:"name" req:"1"`
	}

	userForm, err := forms.FormToStruct[addUserForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	newUser := db.NewUserParams{
		UserEmail:  userForm.Email,
		UserName:   userForm.Name,
		UserActive: true,
	}

	queries := db.New(h.DB())
	insertedUser, err := queries.NewUser(ctx, newUser)
	if err != nil {
		h.Log().Error("error adding user: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "user", insertedUser); err != nil {
		h.Log().Error("failed to render new user: %v", err)
	}
}

func (h *Handler) ToggleUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error toggling user: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	queries := db.New(h.DB())
	err = queries.ToggleUserActiveByUserID(r.Context(), userID)
	if err != nil {
		h.Log().Error("error toggling user: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
