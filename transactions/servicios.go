package transactions

import (
	"database/sql"
)

type ServicioMovimiento struct {
	Movimiento int
	Cuenta     string
	Monto      float64
	Firma      string
}

type Servicio struct {
	ID int
	// Solicitud
	Emitido     sql.NullTime
	Emisor      string
	Detalle     string
	PorEjecutar sql.NullTime
	Justif      string
	// COES
	COES bool
	// OSUM
	ProvNom    string
	ProvCed    string
	ProvDirec  string
	ProvEmail  string
	ProvTel    string
	ProvBanco  string
	ProvIBAN   string
	ProvJustif string
	MontoBruto float64
	MontoIVA   float64
	MontoDesc  float64
	GecoSol    string
	GecoOCS    string
	// Final
	Ejecutado sql.NullTime
	Pagado    sql.NullTime
	Notas     string
	// ProvBanco   string
	// ProvIBAN    string
}
