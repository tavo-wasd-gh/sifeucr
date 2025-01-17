package transactions

import (
	"database/sql"
)

type Servicio struct {
	ID int
	// Solicitud
	Emitido     sql.NullTime
	Cuenta    string
	Detalle     string
	PorEjecutar sql.NullTime
	Justif  string
	Firma string	
	// COES
	COES        bool
	// OSUM
	ProvNom     string
	ProvCed     string
	ProvDirec   string
	ProvEmail   string
	ProvTel     string
	ProvJustif  string
	MontoBruto  float64
	MontoIVA    float64
	MontoDesc   float64
	GecoSol     string
	GecoOCS     string
	// Final
	Ejecutado   sql.NullTime
	Pagado      sql.NullTime
	Notas       string
	// ProvBanco   string
	// ProvIBAN    string
}
