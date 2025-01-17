package transactions

import (
	"database/sql"
)

type Ajuste struct {
	ID         int
	Emitido    sql.NullTime
	Cuenta     string
	Partida    string
	Detalle    string
	MontoBruto float64
	Notas      string
}
