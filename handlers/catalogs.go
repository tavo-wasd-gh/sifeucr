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
		Provider int64 `form:"provider" validate:"nonzero"`
		Grouping int64 `form:"grouping" validate:"nonzero"`
		Article  int64 `form:"article" validate:"nonzero"`
		Desc     string `form:"description" fmt:"trim"`
		Amount   float64 `form:"amount"`
	}

	form, err := forms.FormToStruct[addCatalogForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	queries := db.New(h.DB())

	insertedCatalog, err := queries.AddCatalog(ctx, db.AddCatalogParams{
		CatalogProvider: form.Provider,
		CatalogGrouping: form.Grouping,
		CatalogArticle: form.Article,
		CatalogDescription: form.Desc,
		CatalogAmount: form.Amount,
	})
	if err != nil {
		h.Log().Error("error adding catalog: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "catalog", insertedCatalog); err != nil {
		h.Log().Error("failed to render new catalog: %v", err)
	}
}

func (h *Handler) UpdateCatalog(w http.ResponseWriter, r *http.Request) {
	catalogIDStr := r.PathValue("id")
	catalogID, err := strconv.ParseInt(catalogIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error toggling account: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	type updateCatalogForm struct {
		Provider int64 `form:"provider" validate:"nonzero"`
		Grouping int64 `form:"grouping" validate:"nonzero"`
		Article  int64 `form:"article" validate:"nonzero"`
		Desc     string `form:"description" fmt:"trim"`
		Amount   float64 `form:"amount"`
	}

	form, err := forms.FormToStruct[updateCatalogForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	queries := db.New(h.DB())

	updatedCatalog, err := queries.UpdateCatalog(ctx, db.UpdateCatalogParams{
		CatalogID: catalogID,
		CatalogProvider: form.Provider,
		CatalogGrouping: form.Grouping,
		CatalogArticle: form.Article,
		CatalogDescription: form.Desc,
		CatalogAmount: form.Amount,
	})
	if err != nil {
		h.Log().Error("error updating catalog: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "catalog-update-form", updatedCatalog); err != nil {
		h.Log().Error("failed to render updated catalog: %v", err)
	}
}
