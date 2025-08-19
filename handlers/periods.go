package handlers

import (
	"net/http"
	"strconv"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/views"

	"sifeucr/internal/db"
)

func (h *Handler) AddPeriod(w http.ResponseWriter, r *http.Request) {
	type addPeriodForm struct {
		Name  string `form:"name"  req:"1"`
		Start int64  `form:"start" req:"1"`
		End   int64  `form:"end"   req:"1"`
	}

	periodForm, err := forms.FormToStruct[addPeriodForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	newPeriod := db.AddPeriodParams{
		PeriodName:   periodForm.Name,
		PeriodStart:  periodForm.Start,
		PeriodEnd:    periodForm.End,
		PeriodActive: true,
	}

	queries := db.New(h.DB())
	insertedPeriod, err := queries.AddPeriod(r.Context(), newPeriod)
	if err != nil {
		h.Log().Error("error adding period: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "period", insertedPeriod); err != nil {
		h.Log().Error("failed to render new period: %v", err)
	}
}

func (h *Handler) TogglePeriod(w http.ResponseWriter, r *http.Request) {
	periodIDStr := r.PathValue("id")
	periodID, err := strconv.ParseInt(periodIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error toggling period: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	queries := db.New(h.DB())
	err = queries.TogglePeriodActiveByPeriodID(r.Context(), periodID)
	if err != nil {
		h.Log().Error("error toggling period: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) UpdatePeriod(w http.ResponseWriter, r *http.Request) {
	periodIDStr := r.PathValue("id")
	periodID, err := strconv.ParseInt(periodIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error updating period: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	type updatePeriodForm struct {
		Name  string `form:"name"  req:"1"`
		Start int64  `form:"start" req:"1"`
		End   int64  `form:"end"   req:"1"`
	}

	periodForm, err := forms.FormToStruct[updatePeriodForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	updatedPeriod := db.UpdatePeriodParams{
		PeriodID:    periodID,
		PeriodName:  periodForm.Name,
		PeriodStart: periodForm.Start,
		PeriodEnd:   periodForm.End,
	}

	queries := db.New(h.DB())
	insertedPeriod, err := queries.UpdatePeriod(r.Context(), updatedPeriod)
	if err != nil {
		h.Log().Error("error updating distribution: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "period-update-form", insertedPeriod); err != nil {
		h.Log().Error("failed to render updated period: %v", err)
	}
}
