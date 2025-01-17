package transactions

import (
	"database/sql"
)

type Suministros struct {
	ID int
	// Solicitud
	Emitido       sql.NullTime
	Cuenta      string
	Justif     string
	Desglose	[]Suministro
	// COES
	COES          bool
	// OSUM
	Geco          string
	// Final
	Recibido         string
	Notas         string
	// Agregado
	MontoBrutoTotal float64
}

type Suministro struct {
	ID int
	Articulo string
	Agrupacion string
	Cantidad float64
	MontoUnitario float64
	MontoTotal float64
}
