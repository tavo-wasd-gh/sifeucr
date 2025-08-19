package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/views"

	"sifeucr/internal/db"
)

func (h *Handler) AddDistribution(w http.ResponseWriter, r *http.Request) {
	type addDistributionForm struct {
		Period    int64   `form:"period"  req:"1"`
		EntryCode int64   `form:"entry"   req:"1"`
		Account   int64   `form:"account" req:"1"`
		Amount    float64 `form:"amount"  req:"1"`
	}

	distributionForm, err := forms.FormToStruct[addDistributionForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	newDistribution := db.AddDistributionParams{
		DistPeriod:    distributionForm.Period,
		DistEntryCode: distributionForm.EntryCode,
		DistAccount:   distributionForm.Account,
		DistAmount:    distributionForm.Amount,
		DistActive:    true,
	}

	queries := db.New(h.DB())
	ctx := r.Context()

	i, err := queries.AddDistribution(ctx, newDistribution)
	if err != nil {
		h.Log().Error("error adding distribution: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	insertedDistribution, err := queries.DistributionByID(ctx, i.DistID)
	if err != nil {
		h.Log().Error("error querying new distribution: %v", err)
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
		DistID:     distID,
		DistAmount: distributionForm.Amount,
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

func (h *Handler) budgetEntryByObject(ctx context.Context, obj string) (db.BudgetEntry, error) {
	queries := db.New(h.DB())
	allEntries, err := queries.GetAllBudgetEntries(ctx)
	if err != nil {
		return db.BudgetEntry{}, fmt.Errorf("error looking for budget entry by object: %v", err)
	}

	entry := db.BudgetEntry{}

	found := false
	ob := strings.ToLower(obj)
	for _, e := range allEntries {
		if strings.ToLower(e.EntryObject) == ob {
			entry = e
			found = true
		}
	}

	if !found {
		return db.BudgetEntry{}, fmt.Errorf("error looking for budget entry by object: %v", err)
	}

	return entry, nil
}

func (h *Handler) validDistributionsByAccountID(ctx context.Context, accountID int64) ([]db.FullDistribution, error) {
	queries := db.New(h.DB())

	dd, err := queries.ActiveDistributionsByAccountID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("failed to query account active dists: %v", err)
	}

	now := time.Now().Unix()
	validDists := []db.FullDistribution{}

	for i := range dd {
		if now >= dd[i].PeriodStart && now <= dd[i].PeriodEnd {
			validDists = append(validDists, dd[i])
		}
	}

	if len(validDists) == 0 {
		return nil, fmt.Errorf("could not find current active dist")
	}

	return validDists, nil
}

func (h *Handler) validDistributionByAccountIDAndEntryObject(ctx context.Context, accountID int64, entryObject string) (db.FullDistribution, error) {
	dd, err := h.validDistributionsByAccountID(ctx, accountID)
	if err != nil {
		return db.FullDistribution{}, fmt.Errorf("failed to find dist by accountID and EntryObject: %v", err)
	}

	dist := db.FullDistribution{}

	obj := strings.ToLower(entryObject)
	found := false
	for _, d := range dd {
		if obj == strings.ToLower(d.EntryObject) {
			dist = d
			found = true
		}
	}

	if !found {
		return db.FullDistribution{}, fmt.Errorf("failed to find dist by accountID and EntryObject: not found")
	}

	return dist, nil
}

func (h *Handler) currentServicesDistByAccountID(ctx context.Context, accountID int64) (db.FullDistribution, error) {
	return h.validDistributionByAccountIDAndEntryObject(ctx, accountID, "servicios")
}
