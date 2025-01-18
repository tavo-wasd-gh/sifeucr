package session

import (
	"database/sql"
)

type Presupuesto struct {
	ID      string       `db:"id"`
	Validez sql.NullTime `db:"validez"`
	// Asignado
	Total       float64 `db:"general"`
	Servicios   float64 `db:"servicios"`
	Suministros float64 `db:"suministros"`
	Bienes      float64 `db:"bienes"`
	// Runtime:
	//  -> Emitido
	TotalEmitido       float64
	ServiciosEmitido   float64
	SuministrosEmitido float64
	BienesEmitido      float64
	//  -> Restante
	TotalRestante       float64
	ServiciosRestante   float64
	SuministrosRestante float64
	BienesRestante      float64
}
