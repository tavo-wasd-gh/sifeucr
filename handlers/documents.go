package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"sifeucr/config"
	"sifeucr/internal/db"

	"git.tavo.one/tavo/axiom/views"
)

func (h *Handler) PrintRequestHandler(w http.ResponseWriter, r *http.Request) {
	var err error
	var docReq int64

	docType := r.PathValue("type")
	docReqStr := r.PathValue("req")
	docReq, err = strconv.ParseInt(docReqStr, 10, 64)
	if err != nil {
		h.Log().Error("error printing request: %v", err)
		return
	}

	switch docType {
	case "p":
		err = h.renderQuotation(w, r, docReq)
	case "j":
		err = h.renderJustification(w, r, docReq)
	default:
		//
	}

	if err != nil {
		h.Log().Error("error printing request: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

type QuotationData struct {
	ReqIssued           int64
	SupplierName        string
	SupplierEmail       string
	SupplierPhone       int64
	SupplierID          int64
	SupplierLocation    string
	PurchaseRequired    int64
	ReqDescr            string
	PurchaseGrossAmount float64
	PurchaseTaxRate     float64
	Breakdowns          []db.FullPurchaseBreakdown
	// Calculated
	AccountNames string
	HasBreakdown bool
}

func (h *Handler) renderQuotation(w http.ResponseWriter, r *http.Request, docReq int64) error {
	ctx := r.Context()
	queries := db.New(h.DB())

	purchase, err := queries.FullPurchaseByReqID(ctx, docReq)
	if err != nil {
		return fmt.Errorf("error printing request: %v", err)
	}

	breakdowns, err := queries.BreakdownsByPurchaseID(ctx, purchase.PurchaseID)
	if err != nil {
		return fmt.Errorf("error printing request: %v", err)
	}

	hasBreakdown := false
	if len(breakdowns) != 0 {
		hasBreakdown = true
	}

	subs, err := queries.PurchaseSubscriptionsByRequestID(ctx, docReq)
	if err != nil {
		return fmt.Errorf("error printing request: %v", err)
	}

	accountNames := ""
	for _, s := range subs {
		accountNames += fmt.Sprintf("%s, ", s.AccountName)
	}
	accountNames = strings.TrimSuffix(accountNames, ", ")

	data := QuotationData{
		ReqIssued:           purchase.ReqIssued,
		SupplierName:        purchase.SupplierName,
		SupplierEmail:       purchase.SupplierEmail,
		SupplierPhone:       purchase.SupplierPhone,
		SupplierID:          purchase.SupplierID,
		SupplierLocation:    purchase.SupplierLocation,
		PurchaseRequired:    purchase.PurchaseRequired,
		ReqDescr:            purchase.ReqDescr,
		PurchaseGrossAmount: purchase.PurchaseGrossAmount,
		PurchaseTaxRate:     purchase.PurchaseTaxRate,
		AccountNames:        accountNames,
		Breakdowns:          breakdowns,
		HasBreakdown:        hasBreakdown,
	}

	err = views.RenderHTML(w, r, "doc-quotation", data)
	if err != nil {
		return fmt.Errorf("error printing request: %v", err)
	}

	return nil
}

type JustifData struct {
	ReqIssued           int64
	UserName            string
	ReqDescr            string
	PurchaseRequired    int64
	PurchaseGrossAmount float64
	PurchaseTaxRate     float64
	SupplierID          int64
	SupplierName        string
	ReqJustif           string
	Subscriptions       []db.FullPurchaseSubscription
	// Calculated
	AccountNames string
	TotalAmount  float64
	Distribution string
}

func (h *Handler) renderJustification(w http.ResponseWriter, r *http.Request, docReq int64) error {
	ctx := r.Context()
	queries := db.New(h.DB())

	purchase, err := queries.FullPurchaseByReqID(ctx, docReq)
	if err != nil {
		return fmt.Errorf("error printing request: %v", err)
	}

	subs, err := queries.PurchaseSubscriptionsByRequestID(ctx, docReq)
	if err != nil {
		return fmt.Errorf("error printing request: %v", err)
	}

	distribution := ""
	accountNames := ""
	for _, s := range subs {
		percentage := 100 * s.SubscriptionGrossAmount / purchase.PurchaseGrossAmount
		subAmount := config.FormatAsCurrency(s.SubscriptionGrossAmount)
		taxedAmount := config.FormatAsCurrency(s.SubscriptionGrossAmount * purchase.PurchaseTaxRate)
		totalAmount := config.FormatAsCurrency(s.SubscriptionGrossAmount + s.SubscriptionGrossAmount*purchase.PurchaseTaxRate)
		distribution += fmt.Sprintf(
			"%s: %s + IVA %s, para un aporte total de %s (%.2f%% del servicio), ",
			s.AccountName, subAmount, taxedAmount, totalAmount, percentage,
		)
		accountNames += fmt.Sprintf("%s, ", s.AccountName)
	}
	distribution = strings.TrimSuffix(distribution, ", ")
	accountNames = strings.TrimSuffix(accountNames, ", ")

	data := JustifData{
		AccountNames:        accountNames,
		ReqIssued:           purchase.ReqIssued,
		UserName:            purchase.UserName,
		ReqDescr:            purchase.ReqDescr,
		PurchaseRequired:    purchase.PurchaseRequired,
		PurchaseGrossAmount: purchase.PurchaseGrossAmount,
		PurchaseTaxRate:     purchase.PurchaseTaxRate,
		SupplierID:          purchase.SupplierID,
		SupplierName:        purchase.SupplierName,
		ReqJustif:           purchase.ReqJustif,
		Subscriptions:       subs,
		// Calculated
		TotalAmount: purchase.PurchaseGrossAmount +
			purchase.PurchaseGrossAmount*purchase.PurchaseTaxRate,
		Distribution: distribution,
	}

	err = views.RenderHTML(w, r, "doc-justification", data)
	if err != nil {
		return fmt.Errorf("error printing request: %v", err)
	}

	return nil
}
