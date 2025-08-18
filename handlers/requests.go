package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"git.tavo.one/tavo/axiom/forms"

	"sifeucr/internal/db"
)

// Patch common request parameters, Description and Justification.
func (h *Handler) PatchRequestCommon(w http.ResponseWriter, r *http.Request) {
	type RequestMeta struct {
		Descr  string `form:"req_patch_geco_sol"`
		Justif string `form:"req_patch_geco_ord"`
	}

	requestIDStr := r.PathValue("req")
	if requestIDStr == "" {
		h.Log().Error("error patching common request params: empty requestID pathvalue")
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	requestID, err := strconv.ParseInt(requestIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error patching common request params: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	form, err := forms.FormToStruct[RequestMeta](r)
	if err != nil {
		h.Log().Error("error patching common request params: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	queries := db.New(h.DB())

	request, err := queries.RequestByID(ctx, requestID)
	if err != nil {
		h.Log().Error("error patching common request params: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if form.Descr != "" {
		request.ReqDescr = form.Descr
	}

	if form.Justif != "" {
		request.ReqJustif = form.Justif
	}

	_, err = queries.PatchRequestCommon(ctx, db.PatchRequestCommonParams{
		ReqID:     request.ReqID,
		ReqDescr:  request.ReqDescr,
		ReqJustif: request.ReqJustif,
	})
	if err != nil {
		h.Log().Error("error patching common request params: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "✅")
}

// Patch common purchase parameters, Required date and SupplierID
func (h *Handler) PatchPurchaseCommon(w http.ResponseWriter, r *http.Request) {
	type PurchaseCommon struct {
		Required int64 `form:"req_patch_id"`
		Supplier int64 `form:"req_patch_id"`
	}

	requestIDStr := r.PathValue("req")
	if requestIDStr == "" {
		h.Log().Error("error patching common purchase params: empty requestID pathvalue")
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	requestID, err := strconv.ParseInt(requestIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error patching common purchase params: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	form, err := forms.FormToStruct[PurchaseCommon](r)
	if err != nil {
		h.Log().Error("error patching common purchase params: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	tx, err := h.DB().Begin()
	if err != nil {
		h.Log().Error("error patching common purchase params: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	defer tx.Rollback()

	ctx := r.Context()
	queries := db.New(h.DB())

	purchase, err := queries.FullPurchaseByReqID(ctx, requestID)
	if err != nil {
		h.Log().Error("error patching common purchase params: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	min := purchase.PurchaseRequired - 60*60*12
	if form.Required < min {
		h.Log().Error("error patching common purchase params: cannot modify date more than 12 hours sooner than previous")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if form.Required != 0 {
		purchase.PurchaseRequired = form.Required
	}
	if form.Supplier != 0 {
		purchase.PurchaseSupplier = form.Supplier
	}

	qtx := queries.WithTx(tx)

	_, err = qtx.PatchPurchaseCommon(ctx, db.PatchPurchaseCommonParams{
		PurchaseID:       purchase.PurchaseID,
		PurchaseRequired: purchase.PurchaseRequired,
		PurchaseSupplier: purchase.PurchaseSupplier,
	})
	if err != nil {
		h.Log().Error("error patching common purchase params: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		h.Log().Error("error patching common purchase params: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

// Patch Purchase Subscriptions.
// Total Gross Ampunt must match the sum of gross contributions to the purchase.
// If an requested involved account is not currently subscribing:
//   - Due to deactivated subscription: it will be reactivated.
//   - Due to not listed subscription: it will be added to the table.
//
// If a currently involved account is not requested, its subscription will be deactivated.
func (h *Handler) PatchPurchaseSubscriptions(w http.ResponseWriter, r *http.Request) {
	type PurchaseSubs struct {
		GrossAmount                float64   `form:"purchase_patch_gross_amount"`
		InvolvedAccounts           []int64   `form:"purchase_patch_accounts[]"`
		InvolvedAccountsAmounts    []float64 `form:"purchase_patch_accounts_amounts[]"`
		InvolvedAccountsSignatures []string  `form:"purchase_patch_accounts_signatures[]"`
		InvolvedAccountsSigned     []bool    `form:"purchase_patch_accounts_signed[]"`
	}

	requestIDStr := r.PathValue("req")
	if requestIDStr == "" {
		h.Log().Error("error patching purchase subs: empty requestID pathvalue")
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	requestID, err := strconv.ParseInt(requestIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error patching purchase subs: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	form, err := forms.FormToStruct[PurchaseSubs](r)
	if err != nil {
		h.Log().Error("error patching purchase subs: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	il := len(form.InvolvedAccounts)
	if il == 0 ||
		il != len(form.InvolvedAccountsAmounts) ||
		il != len(form.InvolvedAccountsSignatures) ||
		il != len(form.InvolvedAccountsSigned) {
		h.Log().Error("error patching purchase subs: subscriptions arrays do not match length")
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	tx, err := h.DB().Begin()
	if err != nil {
		h.Log().Error("error patching purchase subs: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	defer tx.Rollback()

	ctx := r.Context()
	queries := db.New(h.DB())

	purchase, err := queries.FullPurchaseByReqID(ctx, requestID)
	if err != nil {
		h.Log().Error("error patching purchase subs: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if form.GrossAmount > 0 {
		purchase.PurchaseGrossAmount = form.GrossAmount
	}

	// Check total sum of involved accounts' gross amounts
	// is equal to purchase gross amount.
	var totalSubSum float64 = 0
	for _, a := range form.InvolvedAccountsAmounts {
		totalSubSum += a
	}
	if purchase.PurchaseGrossAmount != totalSubSum {
		h.Log().Error("error patching purchase subs: total sum of involved accounts' gross amounts does not match purchase gross amount")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	qtx := queries.WithTx(tx)

	// --- for each involved account: ---

	_, err = qtx.PatchPurchaseSub(ctx, db.PatchPurchaseSubParams{
		// SubscriptionID          int64
		// SubscriptionGrossAmount float64
		// SubscriptionSignature   string
		// SubscriptionSigned      bool
	})
	if err != nil {
		h.Log().Error("error patching purchase sub: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// --- end ---

	// TODO: Also update Purchase Gross Amount

	if err := tx.Commit(); err != nil {
		h.Log().Error("error patching purchase subs: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) PatchPurchaseMeta(w http.ResponseWriter, r *http.Request) {
	type PurchaseMeta struct {
		JustifApproved bool   `form:"purchase_patch_justif_approved"`
		GecoSol        string `form:"purchase_patch_geco_sol"`
		GecoOrd        string `form:"purchase_patch_geco_ord"`
		Bill           string `form:"purchase_patch_bill"`
		Transfer       string `form:"purchase_patch_transfer"`
		Status         string `form:"purchase_patch_status"`
	}

	// purchaseID := r.PathValue("id")
}
