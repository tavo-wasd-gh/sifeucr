package database

import (
	"time"
	// "database/sql"
)

type Servicio struct {
	ID int
	// Solicitud
	Emitido *time.Time
	Emisor string
	Detalle string
	PorEjecutar *time.Time
	Justif string
	// COES
	COES bool
	// OSUM
	ProvNom string
	ProvCed string
	ProvDirec string
	ProvEmail string
	ProvTel string
	ProvBanco string
	ProvIBAN string
	ProvJustif string
	MontoBruto float64
	MontoIVA float64
	MontoDesc float64
	GecoSol string
	GecoOCS string
	// ViVE
	OCSFirma string
	OCSFirmaVive string
	// Final
	Ejecutado *time.Time
	Pagado *time.Time
	Notas string
}

// func serviciosInit(db *sql.DB, cuenta string, periodo int) ([]Servicio, error) {

// }
