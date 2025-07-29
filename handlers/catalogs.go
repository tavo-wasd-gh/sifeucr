package handlers

import (
	"net/http"
	"strconv"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/views"

	"sifeucr/internal/db"
)

func (h *Handler) AddCatalog(w http.ResponseWriter, r *http.Request) {
	type addCatalogForm struct {
		Supplier int64 `form:"supplier" validate:"nonzero"`
		Grouping int64 `form:"grouping" validate:"nonzero"`
	}

	form, err := forms.FormToStruct[addCatalogForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	queries := db.New(h.DB())

	i, err := queries.AddCatalog(ctx, db.AddCatalogParams{
		CatalogSupplier: form.Supplier,
		CatalogGrouping: form.Grouping,
	})
	if err != nil {
		h.Log().Error("error adding catalog: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	insertedCatalog, err := queries.CatalogByID(ctx, i.CatalogID)
	if err != nil {
		h.Log().Error("error querying new catalog: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "catalog", insertedCatalog); err != nil {
		h.Log().Error("failed to render new catalog: %v", err)
	}
}

func (h *Handler) AddItem(w http.ResponseWriter, r *http.Request) {
	type addItemForm struct {
		Catalog int64   `req:"1" form:"catalog" validate:"nonzero"`
		Number  int64   `req:"1" form:"number" validate:"nonzero"`
		Summary string  `req:"1" form:"summary" fmt:"trim"`
		Desc    string  `req:"1" form:"desc" fmt:"trim"`
		Amount  float64 `req:"1" form:"amount"`
	}

	form, err := forms.FormToStruct[addItemForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	queries := db.New(h.DB())

	i, err := queries.AddItem(ctx, db.AddItemParams{
		ItemCatalog:     form.Catalog,
		ItemNumber:      form.Number,
		ItemSummary:     form.Summary,
		ItemDescription: form.Desc,
		ItemAmount:      form.Amount,
	})
	if err != nil {
		h.Log().Error("error adding item: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	insertedItem, err := queries.CatalogItemByID(ctx, i.ItemID)
	if err != nil {
		h.Log().Error("error querying new item: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "item", insertedItem); err != nil {
		h.Log().Error("failed to render new item: %v", err)
	}
}

func (h *Handler) UpdateItem(w http.ResponseWriter, r *http.Request) {
	itemIDStr := r.PathValue("id")
	itemID, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error toggling account: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	type updateItemForm struct {
		Number  int64   `req:"1" form:"number" validate:"nonzero"`
		Summary string  `req:"1" form:"summary" fmt:"trim"`
		Desc    string  `req:"1" form:"desc" fmt:"trim"`
		Amount  float64 `req:"1" form:"amount"`
	}

	form, err := forms.FormToStruct[updateItemForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	queries := db.New(h.DB())

	i, err := queries.UpdateItem(ctx, db.UpdateItemParams{
		ItemID:          itemID,
		ItemNumber:      form.Number,
		ItemSummary:     form.Summary,
		ItemDescription: form.Desc,
		ItemAmount:      form.Amount,
	})
	if err != nil {
		h.Log().Error("error updating item: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	insertedItem, err := queries.CatalogItemByID(ctx, i.ItemID)
	if err != nil {
		h.Log().Error("error querying updated item: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "item-update-form", insertedItem); err != nil {
		h.Log().Error("failed to render new item: %v", err)
	}
}
