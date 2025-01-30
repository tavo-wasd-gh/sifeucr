package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Presupuesto struct {
	ID          string
	Index       int
	Cuenta      string
	Validez     time.Time
	General     float64
	Servicios   float64
	Suministros float64
	Bienes      float64
	// Runtime
	Periodo int
	Total   float64
	//   Emitido
	ServiciosEmitido   float64
	SuministrosEmitido float64
	BienesEmitido      float64
	TotalEmitido       float64
	//   Restante
	ServiciosRestante   float64
	SuministrosRestante float64
	BienesRestante      float64
	TotalRestante       float64
}

func presupuestosInit(db *sql.DB, cuenta string, periodo int) ([]Presupuesto, error) {
	query := `SELECT id, cuenta, validez, general, servicios, suministros, bienes
	FROM presupuestos WHERE cuenta = ?
	ORDER BY validez`

	rows, err := db.Query(query, cuenta)
	if err != nil {
		return nil, fmt.Errorf("presupuestos: error querying database: %w", err)
	}
	defer rows.Close()

	var presupuestos []Presupuesto

	index := 0
	for rows.Next() {
		var p Presupuesto
		if err := rows.Scan(
			&p.ID,
			&p.Cuenta,
			&p.Validez,
			&p.General,
			&p.Servicios,
			&p.Suministros,
			&p.Bienes,
		); err != nil {
			return nil, fmt.Errorf("presupuestosActivos: error scanning row: %w", err)
		}

		validez := p.Validez.Year()
		if validez == periodo {
			// Runtime
			index++
			p.Index = index
			p.Periodo = periodo
			p.Total = p.Servicios + p.Suministros + p.Bienes

			if err := p.calcularPresupuesto(db); err != nil {
				return nil, fmt.Errorf("presupuestosActivos: error calculating budget: %w", err)
			}

			presupuestos = append(presupuestos, p)
		}
	}

	return presupuestos, nil
}

func (p *Presupuesto) calcularPresupuesto(db *sql.DB) error {
	queryServicios := `SELECT COALESCE(SUM(monto), 0) FROM servicios_movimientos WHERE presupuesto = ?`
	if err := db.QueryRow(queryServicios, p.ID).Scan(&p.ServiciosEmitido); err != nil {
		return fmt.Errorf("presupuestosEmitido: error calculating ServiciosEmitido for presupuesto %s: %w", p.ID, err)
	}
	p.ServiciosRestante = p.Servicios + p.ServiciosEmitido

	querySuministros := `SELECT COALESCE(SUM(monto_bruto_total), 0) FROM suministros WHERE presupuesto = ?`
	if err := db.QueryRow(querySuministros, p.ID).Scan(&p.SuministrosEmitido); err != nil {
		return fmt.Errorf("presupuestosEmitido: error calculating SuministrosEmitido for presupuesto %s: %w", p.ID, err)
	}
	p.SuministrosRestante = p.Suministros + p.SuministrosEmitido

	queryBienes := `SELECT COALESCE(SUM(monto), 0) FROM bienes_movimientos WHERE presupuesto = ?`
	if err := db.QueryRow(queryBienes, p.ID).Scan(&p.BienesEmitido); err != nil {
		return fmt.Errorf("presupuestosEmitido: error calculating BienesEmitido for presupuesto %s: %w", p.ID, err)
	}
	p.BienesRestante = p.Bienes + p.BienesEmitido

	p.TotalEmitido = p.ServiciosEmitido + p.SuministrosEmitido + p.BienesEmitido
	p.TotalRestante = p.ServiciosRestante + p.SuministrosRestante + p.BienesRestante

	return nil
}
