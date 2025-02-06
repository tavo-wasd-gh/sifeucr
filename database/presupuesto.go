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
		return nil, fmt.Errorf("presupuestosInit: error querying database: %w", err)
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
			return nil, fmt.Errorf("presupuestosInit: error scanning row: %w", err)
		}

		validez := p.Validez.Year()
		if validez == periodo {
			// Runtime
			index++
			p.Index = index
			p.Periodo = periodo
			p.Total = p.Servicios + p.Suministros + p.Bienes

			if err := p.calcularPresupuesto(db); err != nil {
				return nil, fmt.Errorf("presupuestosInit: error calculating budget: %w", err)
			}

			presupuestos = append(presupuestos, p)
		}
	}

	return presupuestos, nil
}

func (p *Presupuesto) calcularPresupuesto(db *sql.DB) error {
	queryGeneral := `SELECT general FROM presupuestos WHERE id = ?`
	if err := db.QueryRow(queryGeneral, p.ID).Scan(&p.General); err != nil {
		return fmt.Errorf("calcularPresupuesto: error calculating presupuesto general %s: %w", p.ID, err)
	}

	var donacionesServiciosSalida, donacionesSuministrosSalida, donacionesBienesSalida float64
	var donacionesServiciosEntrada, donacionesSuministrosEntrada, donacionesBienesEntrada float64

	queryDonacionesSalida := `SELECT COALESCE(SUM(monto_bruto), 0) 
	FROM donaciones 
	WHERE presupuesto_salida = ? AND partida_salida = ?`

	if err := db.QueryRow(queryDonacionesSalida, p.ID, "servicios").Scan(&donacionesServiciosSalida); err != nil {
		return fmt.Errorf("error calculating DonacionesServicios: %w", err)
	}

	if err := db.QueryRow(queryDonacionesSalida, p.ID, "suministros").Scan(&donacionesSuministrosSalida); err != nil {
		return fmt.Errorf("error calculating DonacionesSuministros: %w", err)
	}

	if err := db.QueryRow(queryDonacionesSalida, p.ID, "bienes").Scan(&donacionesBienesSalida); err != nil {
		return fmt.Errorf("error calculating DonacionesBienes: %w", err)
	}

	queryDonacionesEntrada := `SELECT COALESCE(SUM(monto_bruto), 0) 
	FROM donaciones 
	WHERE presupuesto_entrada = ? AND partida_entrada = ?`

	if err := db.QueryRow(queryDonacionesEntrada, p.ID, "servicios").Scan(&donacionesServiciosEntrada); err != nil {
		return fmt.Errorf("error calculating DonacionesServicios: %w", err)
	}

	if err := db.QueryRow(queryDonacionesEntrada, p.ID, "suministros").Scan(&donacionesSuministrosEntrada); err != nil {
		return fmt.Errorf("error calculating DonacionesSuministros: %w", err)
	}

	if err := db.QueryRow(queryDonacionesEntrada, p.ID, "bienes").Scan(&donacionesBienesEntrada); err != nil {
		return fmt.Errorf("error calculating DonacionesBienes: %w", err)
	}

	queryAjustes := `SELECT COALESCE(SUM(monto_bruto), 0) 
	FROM ajustes 
	WHERE presupuesto = ? AND partida = ?`

	var ajustesServicios, ajustesSuministros, ajustesBienes float64

	if err := db.QueryRow(queryAjustes, p.ID, "servicios").Scan(&ajustesServicios); err != nil {
		return fmt.Errorf("error calculating AjustesServicios: %w", err)
	}

	if err := db.QueryRow(queryAjustes, p.ID, "suministros").Scan(&ajustesSuministros); err != nil {
		return fmt.Errorf("error calculating AjustesSuministros: %w", err)
	}

	if err := db.QueryRow(queryAjustes, p.ID, "bienes").Scan(&ajustesBienes); err != nil {
		return fmt.Errorf("error calculating AjustesBienes: %w", err)
	}

	queryServicios := `SELECT COALESCE(SUM(monto), 0) FROM servicios_movimientos WHERE presupuesto = ?`
	if err := db.QueryRow(queryServicios, p.ID).Scan(&p.ServiciosEmitido); err != nil {
		return fmt.Errorf("calcularPresupuesto: error calculating ServiciosEmitido for presupuesto %s: %w", p.ID, err)
	}

	querySuministros := `SELECT COALESCE(SUM(monto_bruto_total), 0) FROM suministros WHERE presupuesto = ?`
	if err := db.QueryRow(querySuministros, p.ID).Scan(&p.SuministrosEmitido); err != nil {
		return fmt.Errorf("calcularPresupuesto: error calculating SuministrosEmitido for presupuesto %s: %w", p.ID, err)
	}

	queryBienes := `SELECT COALESCE(SUM(monto), 0) FROM bienes_movimientos WHERE presupuesto = ?`
	if err := db.QueryRow(queryBienes, p.ID).Scan(&p.BienesEmitido); err != nil {
		return fmt.Errorf("calcularPresupuesto: error calculating BienesEmitido for presupuesto %s: %w", p.ID, err)
	}

	p.ServiciosRestante = p.Servicios - p.ServiciosEmitido + ajustesServicios + donacionesServiciosEntrada - donacionesServiciosSalida
	p.SuministrosRestante = p.Suministros - p.SuministrosEmitido + ajustesSuministros + donacionesSuministrosEntrada - donacionesSuministrosSalida
	p.BienesRestante = p.Bienes - p.BienesEmitido + ajustesBienes + donacionesBienesEntrada - donacionesBienesSalida

	p.TotalEmitido = p.ServiciosEmitido + p.SuministrosEmitido + p.BienesEmitido
	p.TotalRestante = p.ServiciosRestante + p.SuministrosRestante + p.BienesRestante

	return nil
}

func presupuestoActual(db *sql.DB, cuentaID string) (string, error) {
	var presupuestoID string
	currentTime := time.Now()

	query := `
	SELECT id, validez FROM presupuestos
	WHERE cuenta = ?
	AND validez > ? 
	ORDER BY validez
	LIMIT 1;`

	rows, err := db.Query(query, cuentaID, currentTime)
	if err != nil {
		return "", fmt.Errorf("presupuestoActual: failed to fetch valid presupuestos for cuenta %s: %w", cuentaID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var validez time.Time

		if err := rows.Scan(&id, &validez); err != nil {
			return "", fmt.Errorf("presupuestoActual: failed to scan row: %w", err)
		}

		if validez.Year() == currentTime.Year() {
			presupuestoID = id
			break
		}
	}

	if presupuestoID == "" {
		return "", fmt.Errorf("presupuestoActual: no valid presupuesto found for cuenta %s", cuentaID)
	}

	return presupuestoID, nil
}
