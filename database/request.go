package database

import (
	"time"
)

type Request struct {
	ID            int        `db:"id"`
	Type          string     `db:"type"`
	Issued        *time.Time `db:"issued"`
	Wanted        *time.Time `db:"wanted"`
	Issuer        string     `db:"issuer"`
	Description   string     `db:"description"`
	Justification string     `db:"justification"`
	// COES
	Coes       bool    `db:"coes"`
	Correction *string `db:"correction"`
	// OSUM
	GecoID                *string  `db:"geco_id"`
	SupplierName          *string  `db:"supplier_name"`
	SupplierID            *string  `db:"supplier_id"`
	SupplierAddress       *string  `db:"supplier_address"`
	SupplierEmail         *string  `db:"supplier_email"`
	SupplierPhone         *string  `db:"supplier_phone"`
	SupplierBank          *string  `db:"supplier_bank"`
	SupplierIBAN          *string  `db:"supplier_iban"`
	SupplierJustification *string  `db:"supplier_justification"`
	GrossAmount           *float64 `db:"gross_amount"`
	TaxPercentage         *float64 `db:"tax_percentage"`
	DiscountAmount        *float64 `db:"discount_amount"`
	// ViVE
	OrderDocument   *string `db:"order_document"`
	OrderSignedViVE *string `db:"order_signed_vive"`
	// Executed
	Received                 *time.Time `db:"received"`
	Acknowledgement          *string    `db:"acknowledgement"`
	AcknowledgedBy           *string    `db:"acknowledged_by"`
	AcknowledgementSignature *string    `db:"acknowledgement_signature"`
	// Final
	Payed *time.Time `db:"payed"`
	Notes *string    `db:"notes"`
	// Conditions
	Void    bool `db:"void"`
	Deleted bool `db:"deleted"`
}
