package handlers

import (
	"net/http"
	"strconv"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/views"

	"sifeucr/config"
	"sifeucr/internal/db"
)

func (h *Handler) AddPermission(w http.ResponseWriter, r *http.Request) {
	type addPermForm struct {
		User    int64 `form:"user" req:"1"`
		Account int64 `form:"account" req:"1"`
		Integer int64 `form:"integer" req:"1"`
	}

	permForm, err := forms.FormToStruct[addPermForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	newPerm := db.AddPermissionParams{
		PermissionUser:    permForm.User,
		PermissionAccount: permForm.Account,
		PermissionInteger: permForm.Integer,
		PermissionActive:  true,
	}

	queries := db.New(h.DB())
	ctx := r.Context()

	insertedPermission, err := queries.AddPermission(ctx, newPerm)
	if err != nil {
		h.Log().Error("error adding permission: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	insertedPermissionFull, err := queries.PermissionByUserIDAndAccountID(ctx, db.PermissionByUserIDAndAccountIDParams{
		UserID: insertedPermission.PermissionUser,
		AccountID: insertedPermission.PermissionAccount,
	})
	if err != nil {
		h.Log().Error("error querying new permission: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	type permissionRow struct{
		Permission      db.PermissionByUserIDAndAccountIDRow
		PermissionTypes []config.PermissionType
	}

	perm := permissionRow{
		Permission: insertedPermissionFull,
		PermissionTypes: config.PermissionTypes,
	}

	if err = views.RenderHTML(w, r, "permission", perm); err != nil {
		h.Log().Error("failed to render new permission: %v", err)
	}
}

func (h *Handler) TogglePermission(w http.ResponseWriter, r *http.Request) {
	permIDStr := r.PathValue("id")
	permName := r.PathValue("permName")

	permID, err := strconv.ParseInt(permIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error toggling permission: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	queries := db.New(h.DB())
	ctx := r.Context()

	perm, err := queries.PermissionByUserIDAndAccountID(ctx, db.PermissionByUserIDAndAccountIDParams{
		UserID:    getUserIDFromContext(ctx),
		AccountID: getAccountIDFromContext(ctx),
	})
	if err != nil {
		h.Log().Error("error fetching permission: %v", err)
		http.Error(w, "Permission not found", http.StatusInternalServerError)
		return
	}

	var bitToToggle int64
	found := false
	for _, ptype := range config.PermissionTypes {
		if ptype.Name == permName {
			bitToToggle = ptype.Bit
			found = true
			break
		}
	}
	if !found {
		http.Error(w, "Unknown permission type", http.StatusBadRequest)
		return
	}

	newPermInteger := perm.PermissionInteger ^ bitToToggle

	newPerm := db.TogglePermissionByPermissionIDParams{
		PermissionInteger: newPermInteger,
		PermissionID: permID,
	}

	err = queries.TogglePermissionByPermissionID(r.Context(), newPerm)
	if err != nil {
		h.Log().Error("error toggling permission: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
