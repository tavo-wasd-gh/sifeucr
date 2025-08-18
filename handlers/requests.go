package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"git.tavo.one/tavo/axiom/forms"

	"sifeucr/config"
	"sifeucr/internal/db"
)

// Patch common request parameters, Description and Justification.
func (h *Handler) PatchRequestCommon(w http.ResponseWriter, r *http.Request) {
	type RequestMeta struct {
		Descr  string `form:"req_patch_descr"`
		Justif string `form:"req_patch_justif"`
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

	if form.Descr == "" && form.Justif == "" {
		h.Log().Error("error patching common request params: empty request")
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

	if request.ReqAccount != getAccountIDFromContext(ctx) {
		h.Log().Error("error patching common request params: account must be the owner of the request")
		http.Error(w, "", http.StatusUnauthorized)
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

	if form.Descr != "" && form.Justif == "" {
		fmt.Fprintln(w, config.Summary(request.ReqDescr, 40))
	} else if form.Descr == "" && form.Justif != "" {
		fmt.Fprintln(w, config.Summary(request.ReqJustif, 40))
	} else if form.Descr != "" && form.Justif != "" {
		fmt.Fprintln(w, "✅")
	}
}

// Patch common purchase parameters, like Required (date)
func (h *Handler) PatchPurchaseCommon(w http.ResponseWriter, r *http.Request) {
	type PurchaseCommon struct {
		Required int64 `form:"purchase_patch_required"`
		Supplier int64 `form:"purchase_patch_supplier"`
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

	if form.Required != 0 {
		min := purchase.PurchaseRequired - 60*60*12
		if form.Required < min {
			h.Log().Error("error patching common purchase params: cannot modify date more than 12 hours sooner than previous, minimum: %s, requested: %v", config.UnixDateLong(min), config.UnixDateLong(form.Required))
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
		purchase.PurchaseRequired = form.Required
	}
	if form.Supplier != 0 {
		purchase.PurchaseSupplier = form.Supplier
	}

	qtx := queries.WithTx(tx)

	_, err = qtx.PatchPurchaseCommon(ctx, db.PatchPurchaseCommonParams{
		PurchaseID:          purchase.PurchaseID,
		PurchaseRequired:    purchase.PurchaseRequired,
		PurchaseSupplier:    purchase.PurchaseSupplier,
		PurchaseGrossAmount: purchase.PurchaseGrossAmount,
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

	if form.Required != 0 && form.Supplier == 0 {
		fmt.Fprintln(w, config.UnixDateLong(purchase.PurchaseRequired))
	} else if form.Required == 0 && form.Supplier != 0 {
		fmt.Fprintln(w, config.Summary(purchase.SupplierName, 40))
	} else if form.Required != 0 && form.Supplier != 0 {
		fmt.Fprintln(w, "✅")
	}
}

// Patch Purchase Subscriptions. For a patch to be made:
// 1. Total Gross Amount must match the sum of gross contributions to the purchase.
// 2. If an requested involved account is not currently subscribing:
//   - Due to deactivated subscription: it will be reactivated.
//   - Due to not listed subscription: it will be added to the table.
//
// 3. If a currently involved account is not requested, its subscription will be deactivated.
// 4. Only the "owner" of the request can patch it (the account from which it was made)
// 5. The "owner" of the request must be subscribed to it
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

	ctx := r.Context()
	queries := db.New(h.DB())

	requestingUserID := getUserIDFromContext(ctx)
	if requestingUserID == 0 {
		h.Log().Error("error patching purchase subs: needs a requqesting user")
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	purchase, err := queries.FullPurchaseByReqID(ctx, requestID)
	if err != nil {
		h.Log().Error("error patching purchase subs: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	requestingAccountID := getAccountIDFromContext(ctx)

	currentSubsSlice, err := queries.PurchaseSubscriptionsByRequestID(ctx, requestID)
	if err != nil {
		h.Log().Error("error patching purchase subs: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	currentSubs := make(map[int64]db.FullPurchaseSubscription)
	for i := range currentSubsSlice {
		accountID := currentSubsSlice[i].AccountID
		currentSubs[accountID] = currentSubsSlice[i]
	}

	requestedSubs := make(map[int64]db.PurchaseSubscription)
	for i := range il {
		accountID := form.InvolvedAccounts[i]
		if err != nil {
			h.Log().Error("error patching purchase subs: %v", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}

		// These will be the defaults used when patching or adding subscriptions.
		requestedSubs[accountID] = db.PurchaseSubscription{
			// SubscriptionID - Can't yet know, defined in step 2.
			SubscriptionPurchase:    purchase.PurchaseID,
			SubscriptionUser:        requestingUserID, // Can't yet know who will sign for now use the requesting userID since it can't be null.
			SubscriptionDist:        requestedSubs[int64(i)].SubscriptionDist,
			SubscriptionIssued:      time.Now().Unix(),
			SubscriptionGrossAmount: form.InvolvedAccountsAmounts[i],
			SubscriptionSignature:   "",    // Remove requested signatures
			SubscriptionSigned:      false, // ... by default.
			SubscriptionActive:      true,  // Reactivate requested subs by default.
		}
	}

	// 4. Only the "owner" of the request can patch it (the account from which it was made)
	if purchase.ReqAccount != requestingAccountID {
		h.Log().Error("error patching purchase subs: only request owner can manage subscriptions")
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	// 5. The "owner" of the request must be subscribed to it
	if _, ok := requestedSubs[purchase.ReqAccount]; !ok {
		h.Log().Error("error patching purchase subs: the owner of the request must be subscribed to it")
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	// 1. Total Gross Amount must match the sum of gross contributions to the purchase.

	if form.GrossAmount > 0 {
		purchase.PurchaseGrossAmount = form.GrossAmount
	}
	var totalSubSum float64 = 0
	for _, a := range form.InvolvedAccountsAmounts {
		totalSubSum += a
	}
	if purchase.PurchaseGrossAmount != totalSubSum {
		h.Log().Error("error patching purchase subs: total sum of involved accounts' gross amounts does not match purchase gross amount")
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// 2. If an requested involved account is not currently subscribing:

	tx, err := h.DB().Begin()
	if err != nil {
		h.Log().Error("error patching purchase subs: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	defer tx.Rollback()
	qtx := queries.WithTx(tx)

	for reqAcc, reqSub := range requestedSubs {
		currSub, ok := currentSubs[reqAcc]
		var reqSubID int64 = 0

		if !ok {
			// 2.1. Due to not listed subscription: it will be added to the table.
			newSub, err := qtx.AddPurchaseSubscription(ctx, db.AddPurchaseSubscriptionParams{
				SubscriptionPurchase:    reqSub.SubscriptionPurchase,
				SubscriptionUser:        reqSub.SubscriptionUser,
				SubscriptionDist:        reqSub.SubscriptionDist,
				SubscriptionIssued:      reqSub.SubscriptionIssued,
				SubscriptionGrossAmount: reqSub.SubscriptionGrossAmount,
				SubscriptionSignature:   reqSub.SubscriptionSignature,
				SubscriptionSigned:      reqSub.SubscriptionSigned,
				SubscriptionActive:      reqSub.SubscriptionActive,
			})
			if err != nil {
				h.Log().Error("error patching purchase subs: %v", err)
				http.Error(w, "", http.StatusBadRequest)
				return
			}

			reqSubID = newSub.SubscriptionID
		} else {
			// 2.2. Due to deactivated subscription: it will be reactivated.
			// ... Already the default when defining requestedSubs

			reqSubID = currSub.SubscriptionID
		}

		reqSub.SubscriptionID = reqSubID

		_, err := qtx.PatchPurchaseSub(ctx, db.PatchPurchaseSubParams{
			SubscriptionID:          reqSub.SubscriptionID,
			SubscriptionPurchase:    reqSub.SubscriptionPurchase,
			SubscriptionUser:        reqSub.SubscriptionUser,
			SubscriptionDist:        reqSub.SubscriptionDist,
			SubscriptionIssued:      reqSub.SubscriptionIssued,
			SubscriptionGrossAmount: reqSub.SubscriptionGrossAmount,
			SubscriptionSignature:   reqSub.SubscriptionSignature,
			SubscriptionSigned:      reqSub.SubscriptionSigned,
			SubscriptionActive:      reqSub.SubscriptionActive,
		})
		if err != nil {
			h.Log().Error("error patching purchase subs: %v", err)
			http.Error(w, "", http.StatusBadRequest)
			return
		}
	}

	for currAcc, currSub := range currentSubs {
		if _, ok := requestedSubs[currAcc]; !ok && currSub.SubscriptionActive {
			// 3. If a currently subscribed account is not requested, its subscription will be deactivated.
			err = qtx.ToggleSubscriptionActiveByID(ctx, currSub.SubscriptionID)
			if err != nil {
				h.Log().Error("error patching purchase sub: %v", err)
				http.Error(w, "", http.StatusInternalServerError)
				return
			}
		}
	}

	_, err = qtx.PatchPurchaseCommon(ctx, db.PatchPurchaseCommonParams{
		PurchaseID:          purchase.PurchaseID,
		PurchaseRequired:    purchase.PurchaseRequired,
		PurchaseSupplier:    purchase.PurchaseSupplier,
		PurchaseGrossAmount: purchase.PurchaseGrossAmount,
	})
	if err != nil {
		h.Log().Error("error patching purchase subs: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		h.Log().Error("error patching purchase subs: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

// Patch advanced meta information of the purchase, requires WriteAdvanced permission
func (h *Handler) PatchPurchaseMeta(w http.ResponseWriter, r *http.Request) {
	type PurchaseMeta struct {
		GecoSol  string `form:"purchase_patch_geco_sol" fmt:"trim"`
		GecoOrd  string `form:"purchase_patch_geco_ord" fmt:"trim"`
		Bill     string `form:"purchase_patch_bill"     fmt:"trim"`
		Transfer string `form:"purchase_patch_transfer" fmt:"trim"`
		Status   string `form:"purchase_patch_status"   fmt:"trim"`
	}

	reqIDStr := r.PathValue("req")
	reqID, err := strconv.ParseInt(reqIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error patching purchase meta: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	form, err := forms.FormToStruct[PurchaseMeta](r)
	if err != nil {
		h.Log().Error("error patching purchase meta: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	queries := db.New(h.DB())
	ctx := r.Context()

	purchase, err := queries.FullPurchaseByReqID(ctx, reqID)
	if err != nil {
		h.Log().Error("error patching purchase meta: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	var updatedFields []string

	if form.GecoSol != "" {
		purchase.PurchaseGecoSol = form.GecoSol
		updatedFields = append(updatedFields, purchase.PurchaseGecoSol)
	}
	if form.GecoOrd != "" {
		purchase.PurchaseGecoOrd = form.GecoOrd
		updatedFields = append(updatedFields, purchase.PurchaseGecoOrd)
	}
	if form.Bill != "" {
		purchase.PurchaseBill = form.Bill
		updatedFields = append(updatedFields, purchase.PurchaseBill)
	}
	if form.Transfer != "" {
		purchase.PurchaseTransfer = form.Transfer
		updatedFields = append(updatedFields, purchase.PurchaseTransfer)
	}
	if form.Status != "" {
		purchase.PurchaseStatus = form.Status
		updatedFields = append(updatedFields, purchase.PurchaseStatus)
	}

	_, err = queries.PatchPurchaseMeta(ctx, db.PatchPurchaseMetaParams{
		PurchaseID:       purchase.PurchaseID,
		PurchaseGecoSol:  purchase.PurchaseGecoSol,
		PurchaseGecoOrd:  purchase.PurchaseGecoOrd,
		PurchaseBill:     purchase.PurchaseBill,
		PurchaseTransfer: purchase.PurchaseTransfer,
		PurchaseStatus:   purchase.PurchaseStatus,
	})
	if err != nil {
		h.Log().Error("error patching purchase meta: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	switch len(updatedFields) {
	case 0:
		h.Log().Error("error patching purchase meta: no fields updated")
		http.Error(w, "", http.StatusInternalServerError)
		return
	case 1:
		fmt.Fprintln(w, updatedFields[0])
	default:
		fmt.Fprintln(w, "✅")
	}
}
