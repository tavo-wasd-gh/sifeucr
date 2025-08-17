package handlers

import (
	"fmt"
	"net/http"
	"time"

	"sifeucr/config"
	"sifeucr/internal/db"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/views"
)

func (h *Handler) AddPurchase() {
	// TODO: AddPurchase
}

func (h *Handler) PurchaseMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			userID := getUserIDFromContext(ctx)
			accountID := getAccountIDFromContext(ctx)

			queries := db.New(h.DB())
			perm, err := queries.PermissionByUserIDAndAccountID(ctx, db.PermissionByUserIDAndAccountIDParams{
				UserID:    userID,
				AccountID: accountID,
			})
			if err != nil {
				h.Log().Error("failed to authenticate user: %v", err)
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			type purchaseClaims struct {
				Dist int64 `form:"purchase_dist"`
			}

			form, err := forms.FormToStruct[purchaseClaims](r)
			if err != nil {
				h.Log().Error("error casting form to struct: %v", err)
				http.Error(w, "", http.StatusBadRequest)
				return
			}

			if form.Dist != 0 && !config.HasPermission(
				perm.PermissionInteger,
				config.WriteOther,
			) {
				h.Log().Error("cannot add request, insufficient permissions: %v", err)
				http.Error(w, "", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func (h *Handler) PurchaseFormPage(w http.ResponseWriter, r *http.Request) {
	queries := db.New(h.DB())
	ctx := r.Context()

	type Suppliers struct {
		Suppliers []db.Supplier
		Catalogs  []db.FullCatalog
		Items     []db.FullCatalogItem
		CSRFToken string
	}

	suppliers, err := queries.AllSuppliers(ctx)
	if err != nil {
		h.Log().Error("error querying all suppliers: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	catalogs, err := queries.AllCatalogs(ctx)
	if err != nil {
		h.Log().Error("error querying all catalogs: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	items, err := queries.AllCatalogItems(ctx)
	if err != nil {
		h.Log().Error("error querying all items: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	csrfToken := getCSRFTokenFromContext(ctx)

	data := Suppliers{
		Suppliers: suppliers,
		Catalogs:  catalogs,
		Items:     items,
		CSRFToken: csrfToken,
	}

	err = views.RenderHTML(w, r, "forms-purchase-form-page", data)
	if err != nil {
		h.Log().Error("error rendering view: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}
}

func (h *Handler) newGenericPurchase(r *http.Request) error {
	type genericPurchaseForm struct {
		Required    int64   `form:"purchase_required" req:"1"`
		Desc        string  `form:"purchase_desc" req:"1" fmt:"trim"`
		Justif      string  `form:"purchase_justif" req:"1" fmt:"trim"`
		Supplier    int64   `form:"purchase_supplier" req:"1"`
		GrossAmount float64 `form:"purchase_gross_amount" req:"1"`
		Signature   string  `form:"purchase_signature" req:"1"`
	}
	form, err := forms.FormToStruct[genericPurchaseForm](r)
	if err != nil {
		return fmt.Errorf("failed to cast form to struct: %v", err)
	}

	if len(form.Desc) < 30 || len(form.Justif) < 150 {
		return fmt.Errorf("description or justification length too small")
	}

	min := time.Now().Add(24*time.Hour*7).Unix()

	if form.Required > min {
		return fmt.Errorf("failed to register new generic purchase: minimum allowed date: %d, asked for: %d", min, form.Required)
	}

	ctx := r.Context()

	userID := getUserIDFromContext(ctx)
	accountID := getAccountIDFromContext(ctx)

	dist, err := h.getCurrentActiveDist(ctx, accountID)
	if err != nil {
		return fmt.Errorf("failed to register new generic purchase: %v", err)
	}

	tx, err := h.DB().Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queries := db.New(h.DB())
	qtx := queries.WithTx(tx)

	// Step 1: Register request
	request, err := qtx.AddRequest(ctx, db.AddRequestParams{
		ReqUser: userID,
		ReqAccount: accountID,
		ReqIssued: time.Now().Unix(),
		ReqDescr: form.Desc,
		ReqJustif: form.Justif,
	})
	if err != nil {
		return fmt.Errorf("failed to add request: %v", err)
	}

	// Step 2: Register purchase
	purchase, err := qtx.AddPurchase(ctx, db.AddPurchaseParams{
		PurchaseRequest:     request.ReqID,
		PurchaseRequired:    form.Required,
		PurchaseSupplier:    form.Supplier,
		PurchaseGrossAmount: form.GrossAmount,
		// New purchase empy
		PurchaseGecoSol:  "",
		PurchaseGecoOrd:  "",
		PurchaseBill:     "",
		PurchaseTransfer: "",
		PurchaseStatus:   "",
		// Typical defaults
		PurchaseJustifApproved: false,
		PurchaseCurrency:       "CRC",
		PurchaseExRateColones:  1.00,
		PurchaseDiscount:       0.00,
		PurchaseTaxRate:        0.02,
	})
	if err != nil {
		return fmt.Errorf("failed to add purchase: %v", err)
	}

	// Step 3: Register initial subscription
	_, err = qtx.AddPurchaseSubscription(ctx, db.AddPurchaseSubscriptionParams{
		SubscriptionPurchase:    purchase.PurchaseID,
		SubscriptionUser:        userID,
		SubscriptionDist:        dist.DistID,
		SubscriptionIssued:      time.Now().Unix(),
		SubscriptionGrossAmount: form.GrossAmount,
		SubscriptionSignature:   form.Signature,
		SubscriptionSigned:      true,
	})

	return tx.Commit()
}

func (h *Handler) NewPurchase(w http.ResponseWriter, r *http.Request) {
	purchaseType := r.FormValue("purchase_type")
	var err error

	switch purchaseType {
	case "catering":
		// err = h.newCateringPurchase(r)
	case "generic":
		err = h.newGenericPurchase(r)
	default:
		// handle anything else
	}

	if err != nil {
		h.Log().Error("error registering new purchase: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	} else {
		err = views.RenderHTML(w, r, "forms-purchase-registered", nil)
		if err != nil {
			h.Log().Error("error registering new purchase: %v", err)
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	}
}
