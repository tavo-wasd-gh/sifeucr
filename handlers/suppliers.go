package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"time"

	"git.tavo.one/tavo/axiom/forms"
	"git.tavo.one/tavo/axiom/mail"
	"git.tavo.one/tavo/axiom/views"

	"sifeucr/internal/db"
)

const (
	tokenResponse = `<p>Para ver los trámites pendientes, ingrese <a href="https://si.feucr.org/proveedores/%s/%s">aquí</a></p>`
)

var SuppliersLastRequest = map[string]int64{}

func (h *Handler) AddSupplier(w http.ResponseWriter, r *http.Request) {
	type addSupplierForm struct {
		ID    int64  `form:"id" validate:"nonzero"`
		Name  string `form:"name" fmt:"trim"`
		Email string `form:"email" fmt:"trim,lower" validate:"email"`
		Cntry int64  `form:"country"`
		Phone int64  `form:"phone"`
		Loctn string `form:"location"`
	}

	form, err := forms.FormToStruct[addSupplierForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	queries := db.New(h.DB())

	insertedSupplier, err := queries.AddSupplier(ctx, db.AddSupplierParams{
		SupplierID:               form.ID,
		SupplierName:             form.Name,
		SupplierEmail:            form.Email,
		SupplierPhoneCountryCode: form.Cntry,
		SupplierPhone:            form.Phone,
		SupplierLocation:         form.Loctn,
	})
	if err != nil {
		h.Log().Error("error adding supplier: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "supplier", insertedSupplier); err != nil {
		h.Log().Error("failed to render new supplier: %v", err)
	}
}

func (h *Handler) UpdateSupplier(w http.ResponseWriter, r *http.Request) {
	supplierIDStr := r.PathValue("id")
	supplierID, err := strconv.ParseInt(supplierIDStr, 10, 64)
	if err != nil {
		h.Log().Error("error toggling account: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	type updateSupplierForm struct {
		Name  string `form:"name" fmt:"trim"`
		Email string `form:"email" fmt:"trim,lower" validate:"email"`
		Cntry int64  `form:"country"`
		Phone int64  `form:"phone"`
		Loctn string `form:"location"`
	}

	form, err := forms.FormToStruct[updateSupplierForm](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	queries := db.New(h.DB())

	updatedSupplier, err := queries.UpdateSupplier(ctx, db.UpdateSupplierParams{
		SupplierID:               supplierID,
		SupplierName:             form.Name,
		SupplierEmail:            form.Email,
		SupplierPhoneCountryCode: form.Cntry,
		SupplierPhone:            form.Phone,
		SupplierLocation:         form.Loctn,
	})
	if err != nil {
		h.Log().Error("error updating supplier: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if err = views.RenderHTML(w, r, "supplier-update-form", updatedSupplier); err != nil {
		h.Log().Error("failed to render updated supplier: %v", err)
	}
}

func SupplierToken(key, input string) (string, error) {
	keyBytes := []byte(key)
	message := []byte(input)

	h := hmac.New(sha256.New, keyBytes)

	_, err := h.Write(message)
	if err != nil {
		return "", err
	}

	hash := h.Sum(nil)

	return base64.URLEncoding.EncodeToString(hash), nil
}

func (h *Handler) SendSupplierSummaryToken(w http.ResponseWriter, r *http.Request) {
	type summaryReq struct {
		Email string `form:"email" fmt:"trim,lower" validate:"email"`
	}

	form, err := forms.FormToStruct[summaryReq](r)
	if err != nil {
		h.Log().Error("error casting form to struct: %v", err)
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	queries := db.New(h.DB())

	allEmails, err := queries.SupplierEmails(r.Context())
	if err != nil {
		h.Log().Error("error querting all supplier emails: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if !slices.Contains(allEmails, form.Email) {
		h.Log().Error("the requested supplier does not exist: %s", form.Email)
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	if lastRequest, ok := SuppliersLastRequest[form.Email]; ok {
		timeoutEnd := lastRequest + 60*60*1
		if time.Now().Unix() < timeoutEnd {
			h.Log().Error("error generating supplier token: timeout has not been reached")
			http.Error(w, "", http.StatusInternalServerError)
			return
		}
	} else {
		SuppliersLastRequest[form.Email] = time.Now().Unix()
	}

	token, err := SupplierToken(h.ServerSecret(), form.Email)
	if err != nil {
		h.Log().Error("error generating supplier token: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	s := smtp.Client("smtp.ucr.ac.cr", "587", h.SmtpPass())

	err = s.SendHTML(
		h.SmtpUser(),
		[]string{form.Email},
		"[SIFEUCR] Ver Procesos",
		fmt.Sprintf(tokenResponse, url.QueryEscape(form.Email), token),
		nil,
	)
	if err != nil {
		h.Log().Error("error sending supplier token: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, "Listo! Enviamos el link para visualizar los trámites a su correo.")
}

func (h *Handler) LoadSupplierSummary(w http.ResponseWriter, r *http.Request) {
	email := r.PathValue("email")
	claimedToken := r.PathValue("token")

	token, err := SupplierToken(h.ServerSecret(), email)
	if err != nil {
		h.Log().Error("error generating supplier token: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if token != claimedToken {
		h.Log().Error("error fetching supplier summary: bad token")
		http.Error(w, "", http.StatusUnauthorized)
		return
	}

	ctx := r.Context()
	queries := db.New(h.DB())
	allPurchasesBySupplierEmail, err := queries.AllPurchasesBySupplierEmail(ctx, email)
	if err != nil {
		h.Log().Error("error loading supplier summary: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	type supplierPurchase struct {
		ReqID               int64
		ReqDescr            string
		PurchaseRequired    int64
		PurchaseTaxRate     float64
		PurchaseGrossAmount float64
	}

	type supplierSummary struct {
		Purchases     []supplierPurchase
		CSRFToken     string
		SupplierToken string
	}

	summary := supplierSummary{
		Purchases:     make([]supplierPurchase, len(allPurchasesBySupplierEmail)),
		SupplierToken: token,
	}

	for i, p := range allPurchasesBySupplierEmail {
		summary.Purchases[i] = supplierPurchase{
			ReqID:               p.ReqID,
			ReqDescr:            p.ReqDescr,
			PurchaseRequired:    p.PurchaseRequired,
			PurchaseTaxRate:     p.PurchaseTaxRate,
			PurchaseGrossAmount: p.PurchaseGrossAmount,
		}
	}

	err = views.RenderHTML(w, r, "supplier-summary-page", summary)
	if err != nil {
		h.Log().Error("error loading supplier summary: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
