package database

import (
	"database/sql"
	"fmt"
)

type Presupuesto struct {
	ID string
	Index int
	Cuenta string
	Validez sql.NullTime
	General float64
	Servicios float64
	Suministros float64
	Bienes float64
	// Runtime
	Periodo int
	Total float64
	Restante float64
	//   Emitido
	ServiciosEmitido float64
	SuministrosEmitido float64
	BienesEmitido float64
	//   Restante
	ServiciosRestante float64
	SuministrosRestante float64
	BienesRestante float64
}

func presupuestosInit(db *sql.DB, cuenta string, periodo int) ([]Presupuesto, error) {
	query := `SELECT id, cuenta, validez, general, servicios, suministros, bienes
	FROM presupuestos WHERE cuenta = ?
	ORDER BY validez`
	
	rows, err := db.Query(query, cuenta, periodo)
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

		validez := p.Validez.Time.Year()
		if validez == periodo {
			// Runtime
			index++
			p.Index = index
			p.Periodo = periodo
			p.Total = p.Servicios + p.Suministros + p.Bienes
			p.Restante = p.Total - (p.Servicios + p.Suministros + p.Bienes)

			presupuestos = append(presupuestos, p)
		}
	}

	return presupuestos, nil
}

func presupuestosEmitido(db *sql.DB, p *Presupuesto) error {
	queryServicios := `SELECT COALESCE(SUM(monto), 0) FROM servicios_movimientos WHERE presupuesto = ?`
	if err := db.QueryRow(queryServicios, p.ID).Scan(&p.ServiciosEmitido); err != nil {
		return fmt.Errorf("presupuestosEmitido: error calculating ServiciosEmitido for presupuesto %s: %w", p.ID, err)
	}

	querySuministros := `SELECT COALESCE(SUM(monto_bruto_total), 0) FROM suministros WHERE presupuesto = ?`
	if err := db.QueryRow(querySuministros, p.ID).Scan(&p.SuministrosEmitido); err != nil {
		return fmt.Errorf("presupuestosEmitido: error calculating SuministrosEmitido for presupuesto %s: %w", p.ID, err)
	}

	queryBienes := `SELECT COALESCE(SUM(monto), 0) FROM bienes_movimientos WHERE presupuesto = ?`
	if err := db.QueryRow(queryBienes, p.ID).Scan(&p.BienesEmitido); err != nil {
		return fmt.Errorf("presupuestosEmitido: error calculating BienesEmitido for presupuesto %s: %w", p.ID, err)
	}

	return nil
}
