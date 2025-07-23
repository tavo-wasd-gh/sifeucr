package handlers

import (
	"net/http"
	"strconv"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/views"

	"sifeucr/internal/db"
)

func (h *Handler) AddSupplier(w http.ResponseWriter, r *http.Request) {
	type addSupplierForm struct {
		ID    int64  `form:"id" validate:"nonzero"`
		Name  string `form:"name" fmt:"trim"`
		Email string `form:"email" fmt:"trim,lower" validate:"email"`
		Cntry int64  `form:"country"`
		Phone int64  `form:"phone"`
		Loctn string `form:"location"`
	}

	form, err := forms.FormToStruct[addSupplierForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	queries := db.New(h.DB())

	insertedSupplier, err := queries.AddSupplier(ctx, db.AddSupplierParams{
		SupplierID:               form.ID,
		SupplierName:             form.Name,
		SupplierEmail:            form.Email,
		SupplierPhoneCountryCode: form.Cntry,
		SupplierPhone:            form.Phone,
		SupplierLocation:         form.Loctn,
	})
	if err != nil {
		h.Log().Error("error adding supplier: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "supplier", insertedSupplier); err != nil {
		h.Log().Error("failed to render new supplier: %v", err)
	}
}

func (h *Handler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	supplierIDStr := r.PathValue("id")
	supplierID, err := strconv.ParseInt(supplierIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error toggling account: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	type updateSupplierForm struct {
		Name  string `form:"name" fmt:"trim"`
		Email string `form:"email" fmt:"trim,lower" validate:"email"`
		Cntry int64  `form:"country"`
		Phone int64  `form:"phone"`
		Loctn string `form:"location"`
	}

	form, err := forms.FormToStruct[updateSupplierForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	queries := db.New(h.DB())

	updatedSupplier, err := queries.UpdateSupplier(ctx, db.UpdateSupplierParams{
		SupplierID:               supplierID,
		SupplierName:             form.Name,
		SupplierEmail:            form.Email,
		SupplierPhoneCountryCode: form.Cntry,
		SupplierPhone:            form.Phone,
		SupplierLocation:         form.Loctn,
	})
	if err != nil {
		h.Log().Error("error updating supplier: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "supplier-update-form", updatedSupplier); err != nil {
		h.Log().Error("failed to render updated supplier: %v", err)
	}
}
