package handlers

import (
	"net/http"
	"strconv"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/views"

	"sifeucr/internal/db"
)

func (h *Handler) AddDistribution(w http.ResponseWriter, r *http.Request) {
	type addDistributionForm struct {
		Period     int64   `form:"period"      req:"1"`
		EntryCode  int64   `form:"entry"       req:"1"`
		Account    int64   `form:"account"     req:"1"`
		Amount     float64 `form:"amount"      req:"1"`
	}

	distributionForm, err := forms.FormToStruct[addDistributionForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	newDistribution := db.AddDistributionParams{
		DistPeriod:     distributionForm.Period,
		DistEntryCode:  distributionForm.EntryCode,
		DistAccount:    distributionForm.Account,
		DistAmount:     distributionForm.Amount,
		DistActive:     true,
	}

	queries := db.New(h.DB())
	insertedDistribution, err := queries.AddDistribution(r.Context(), newDistribution)
	if err != nil {
		h.Log().Error("error adding distribution: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "distribution", insertedDistribution); err != nil {
		h.Log().Error("failed to render new distribution: %v", err)
	}
}

func (h *Handler) ToggleDistribution(w http.ResponseWriter, r *http.Request) {
	distIDStr := r.PathValue("id")
	distID, err := strconv.ParseInt(distIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error toggling distribution: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	queries := db.New(h.DB())
	err = queries.ToggleDistributionActiveByDistributionID(r.Context(), distID)
	if err != nil {
		h.Log().Error("error toggling distribution: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdateDistribution(w http.ResponseWriter, r *http.Request) {
	distIDStr := r.PathValue("id")
	distID, err := strconv.ParseInt(distIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error updating distribution: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	type updateDistributionForm struct {
		Amount float64 `form:"amount" req:"1"`
	}

	distributionForm, err := forms.FormToStruct[updateDistributionForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	updatedDistribution := db.UpdateDistributionParams{
		DistID:         distID,
		DistAmount:     distributionForm.Amount,
	}

	queries := db.New(h.DB())
	insertedDistribution, err := queries.UpdateDistribution(r.Context(), updatedDistribution)
	if err != nil {
		h.Log().Error("error updating distribution: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "dist-update-form", insertedDistribution); err != nil {
		h.Log().Error("failed to render updated distribution: %v", err)
	}
}
