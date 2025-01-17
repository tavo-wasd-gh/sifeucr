package transactions

import (
	"database/sql"
)

type Bien struct {
	ID int
	// Solicitud
	Emitido    sql.NullTime
	Cuenta     string
	Detalle    string
	PorRecibir sql.NullTime
	JustifBien string
	Firma      string
	// COES
	COES bool
	// OSUM
	ProvNom    string
	ProvCed    string
	ProvDirec  string
	ProvEmail  string
	ProvTel    string
	ProvJustif string
	MontoBruto float64
	MontoIVA   float64
	MontoDesc  float64
	GecoSol    string
	GecoOC     string
	// Final
	Recibido sql.NullTime
	Notas    string
	// ProvBanco  string
	// ProvIBAN   string
}
