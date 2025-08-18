package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"sifeucr/config"
	"sifeucr/internal/db"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/views"
)

type registerPurchaseParams struct {
	UserID             int64
	AccountID          int64
	Desc               string
	Justif             string
	Required           int64
	SupplierID         int64
	GrossAmount        float64
	Signature          string
	ItemsIDAndQuantity []itemIDAndQuantity
}

type itemIDAndQuantity struct {
	ID       int64
	Quantity float64
}

func (h *Handler) registerPurchase(ctx context.Context, params registerPurchaseParams) (int64, error) {
	if len(params.Desc) < 30 || len(params.Justif) < 150 {
		return 0, fmt.Errorf("description or justification length too small")
	}

	min := time.Now().Add(24 * time.Hour * 6).Unix()
	if params.Required < min {
		return 0, fmt.Errorf("minimum allowed date: %s, asked for: %s", config.UnixDateLong(min), config.UnixDateLong(params.Required))
	}

	dist, err := h.getCurrentActiveDist(ctx, params.AccountID)
	if err != nil {
		return 0, fmt.Errorf("failed to register new generic purchase: %v", err)
	}

	tx, err := h.DB().Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	queries := db.New(h.DB())
	qtx := queries.WithTx(tx)

	// Step 1: Register request
	request, err := qtx.AddRequest(ctx, db.AddRequestParams{
		ReqUser:    params.UserID,
		ReqAccount: params.AccountID,
		ReqIssued:  time.Now().Unix(),
		ReqDescr:   params.Desc,
		ReqJustif:  params.Justif,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to add request: %v", err)
	}

	// Step 2: Register purchase
	purchase, err := qtx.AddPurchase(ctx, db.AddPurchaseParams{
		PurchaseRequest:     request.ReqID,
		PurchaseRequired:    params.Required,
		PurchaseSupplier:    params.SupplierID,
		PurchaseGrossAmount: params.GrossAmount,
		// New purchase empy
		PurchaseGecoSol:  "",
		PurchaseGecoOrd:  "",
		PurchaseBill:     "",
		PurchaseTransfer: "",
		PurchaseStatus:   "",
		// Typical defaults
		PurchaseCurrency:      "CRC",
		PurchaseExRateColones: 1.00,
		PurchaseDiscount:      0.00,
		PurchaseTaxRate:       0.02,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to add purchase: %v", err)
	}

	// Step 3: Register initial subscription
	_, err = qtx.AddPurchaseSubscription(ctx, db.AddPurchaseSubscriptionParams{
		SubscriptionPurchase:    purchase.PurchaseID,
		SubscriptionUser:        params.UserID,
		SubscriptionDist:        dist.DistID,
		SubscriptionIssued:      time.Now().Unix(),
		SubscriptionGrossAmount: params.GrossAmount,
		SubscriptionSignature:   params.Signature,
		SubscriptionSigned:      true,
	})

	// Step 4: Register breakdown
	for _, item := range params.ItemsIDAndQuantity {
		_, err = qtx.AddPurchaseBreakdown(ctx, db.AddPurchaseBreakdownParams{
			BreakdownPurchase: purchase.PurchaseID,
			BreakdownItem:     item.ID,
			BreakdownQuantity: item.Quantity,
		})
		if err != nil {
			return 0, fmt.Errorf("failed to add item to purchase breakdown: %v", err)
		}
	}

	return request.ReqID, tx.Commit()
}

// TODO: This is almost the same as Catering purchase, optimize later.
func (h *Handler) newSuppliesPurchase(r *http.Request) error {
	type suppliesPurchaseForm struct {
		Required   int64     `form:"purchase_required"         req:"1"`
		Desc       string    `form:"purchase_desc"             req:"1" fmt:"trim"`
		Justif     string    `form:"purchase_justif"           req:"1" fmt:"trim"`
		Catalogs   []int64   `form:"purchase_items_catalog[]"  req:"1"`
		Articles   []int64   `form:"purchase_items_article[]"  req:"1"`
		Quantities []float64 `form:"purchase_items_quantity[]" req:"1"`
		Signature  string    `form:"purchase_signature"        req:"1"`
	}

	form, err := forms.FormToStruct[suppliesPurchaseForm](r)
	if err != nil {
		return fmt.Errorf("failed to cast form to struct: %v", err)
	}

	il := len(form.Catalogs)
	if il == 0 ||
		il != len(form.Articles) ||
		il != len(form.Quantities) {
		return fmt.Errorf("failed to register supplies purchase: slices different sizes or empty")
	}

	items := make([]itemIDAndQuantity, il)
	for i := range il {
		items[i] = itemIDAndQuantity{
			ID:       form.Articles[i],
			Quantity: form.Quantities[i],
		}
	}

	ctx := r.Context()
	queries := db.New(h.DB())

	supplier, err := queries.SupplierByName(ctx, "OSUM")
	if err != nil {
		return fmt.Errorf("failed to query supplier by Name: %v", err)
	}

	// iterate over catalogs, articles and quantities.
	// 1. Sums total purchase amount from items prices.
	var totalAmount float64 = 0
	for i, c := range form.Catalogs {
		requestedSupplier, err := queries.SupplierByCatalogGrouping(ctx, c)
		if err != nil {
			return fmt.Errorf("failed to query supplier by CatalogGrouping: %v", err)
		}
		if requestedSupplier.SupplierID != supplier.SupplierID {
			return fmt.Errorf("failed to register catering purchase: supplies purchases can only have catalogs corresponding to 'OSUM' account of ID: %d", supplier.SupplierID)
		}

		itemAmount, err := queries.ItemAmountByID(ctx, form.Articles[i])
		if err != nil {
			return fmt.Errorf("failed to query item price: %v", err)
		}

		totalAmount += itemAmount * form.Quantities[i]
	}

	userID := getUserIDFromContext(ctx)
	accountID := getAccountIDFromContext(ctx)

	params := registerPurchaseParams{
		UserID:             userID,
		AccountID:          accountID,
		Desc:               form.Desc,
		Justif:             form.Justif,
		Required:           form.Required,
		SupplierID:         supplier.SupplierID,
		GrossAmount:        totalAmount,
		Signature:          form.Signature,
		ItemsIDAndQuantity: items,
	}

	_, err = h.registerPurchase(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to register supplies purchase: %v", err)
	}

	return nil
}

func (h *Handler) newCateringPurchase(r *http.Request) error {
	type cateringPurchaseForm struct {
		Required   int64     `form:"purchase_required"         req:"1"`
		Desc       string    `form:"purchase_desc"             req:"1" fmt:"trim"`
		Justif     string    `form:"purchase_justif"           req:"1" fmt:"trim"`
		Catalogs   []int64   `form:"purchase_items_catalog[]"  req:"1"`
		Articles   []int64   `form:"purchase_items_article[]"  req:"1"`
		Quantities []float64 `form:"purchase_items_quantity[]" req:"1"`
		Signature  string    `form:"purchase_signature"        req:"1"`
	}

	form, err := forms.FormToStruct[cateringPurchaseForm](r)
	if err != nil {
		return fmt.Errorf("failed to cast form to struct: %v", err)
	}

	il := len(form.Catalogs)
	if il == 0 ||
		il != len(form.Articles) ||
		il != len(form.Quantities) {
		return fmt.Errorf("failed to register catering purchase: slices different sizes or empty")
	}

	items := make([]itemIDAndQuantity, il)
	for i := range il {
		items[i] = itemIDAndQuantity{
			ID:       form.Articles[i],
			Quantity: form.Quantities[i],
		}
	}

	ctx := r.Context()
	queries := db.New(h.DB())

	// iterate over catalogs, articles and quantities.
	// 1. Checks the purchase comes only from one catalog, since
	//    multiple catalogs in a catering purchase are not allowed.
	// 2. Sums total purchase amount from items prices.
	var totalAmount float64 = 0
	var catalog int64 = 0
	for i, c := range form.Catalogs {
		if catalog == 0 {
			catalog = c
		} else if catalog != c {
			return fmt.Errorf("failed to register catering purchase: multiple catalogs in a catering purchase are not allowed")
		}

		itemAmount, err := queries.ItemAmountByID(ctx, form.Articles[i])
		if err != nil {
			return fmt.Errorf("failed to query item price: %v", err)
		}

		totalAmount += itemAmount * form.Quantities[i]
	}

	supplier, err := queries.SupplierByCatalogGrouping(ctx, catalog)
	if err != nil {
		return fmt.Errorf("failed to query supplier by catalogID: %v", err)
	}

	userID := getUserIDFromContext(ctx)
	accountID := getAccountIDFromContext(ctx)

	params := registerPurchaseParams{
		UserID:             userID,
		AccountID:          accountID,
		Desc:               form.Desc,
		Justif:             form.Justif,
		Required:           form.Required,
		SupplierID:         supplier.SupplierID,
		GrossAmount:        totalAmount,
		Signature:          form.Signature,
		ItemsIDAndQuantity: items,
	}

	_, err = h.registerPurchase(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to register catering purchase: %v", err)
	}

	return nil
}

func (h *Handler) newGenericPurchase(r *http.Request) error {
	type genericPurchaseForm struct {
		Required    int64   `form:"purchase_required" req:"1"`
		Desc        string  `form:"purchase_desc" req:"1" fmt:"trim"`
		Justif      string  `form:"purchase_justif" req:"1" fmt:"trim"`
		SupplierID  int64   `form:"purchase_supplier" req:"1"`
		GrossAmount float64 `form:"purchase_gross_amount" req:"1"`
		Signature   string  `form:"purchase_signature" req:"1"`
	}
	form, err := forms.FormToStruct[genericPurchaseForm](r)
	if err != nil {
		return fmt.Errorf("failed to cast form to struct: %v", err)
	}

	ctx := r.Context()
	userID := getUserIDFromContext(ctx)
	accountID := getAccountIDFromContext(ctx)

	params := registerPurchaseParams{
		UserID:      userID,
		AccountID:   accountID,
		Desc:        form.Desc,
		Justif:      form.Justif,
		Required:    form.Required,
		SupplierID:  form.SupplierID,
		GrossAmount: form.GrossAmount,
		Signature:   form.Signature,
	}

	_, err = h.registerPurchase(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to register generic purchase: %v", err)
	}

	return nil
}

func (h *Handler) NewPurchase(w http.ResponseWriter, r *http.Request) {
	purchaseType := r.FormValue("purchase_type")
	var err error

	switch purchaseType {
	case "catering":
		err = h.newCateringPurchase(r)
	case "generic":
		err = h.newGenericPurchase(r)
	case "supplies":
		err = h.newSuppliesPurchase(r)
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
