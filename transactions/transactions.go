package transactions

import (
	"database/sql"
)

type Usuario struct {
	PeriodoActual int
	Cuenta        Cuenta
	Servicios     []Servicio
	Suministros   []Suministros
	Bienes        []Bien
	Ajustes       []Ajuste
	Donaciones    []Donacion
}

type Cuenta struct {
	ID          string
	Nombre      string
	Presidencia string
	Tesoreria   string
	P1          Presupuesto
	P2          Presupuesto
	TEEU        bool
	COES        bool
}

type Presupuesto struct {
	ID      int
	Cuenta  string
	Validez sql.NullTime
	// Asignado
	General     float64
	Servicios   float64
	Suministros float64
	Bienes      float64
	// Runtime:
	//  -> Emitido
	GeneralEmitido     float64
	ServiciosEmitido   float64
	SuministrosEmitido float64
	BienesEmitido      float64
	//  -> Restante
	GeneralRestante     float64
	ServiciosRestante   float64
	SuministrosRestante float64
	BienesRestante      float64
}
