package session

import (
	"database/sql"
)

type Donacion struct {
	ID             int
	Emitido        sql.NullTime
	CuentaSalida   string
	PartidaSalida  string
	CuentaEntrada  string
	PartidaEntrada string
	Detalle        string
	MontoBruto     float64
	CartaCOES      string
	Notas          string
}
