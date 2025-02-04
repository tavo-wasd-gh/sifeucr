package database

import (
	"time"
	"fmt"
	"database/sql"
)

type Donacion struct {
	ID             int
	Emitido        time.Time
	CuentaSalida   string
	PartidaSalida  string
	CuentaEntrada  string
	PartidaEntrada string
	Detalle        string
	MontoBruto     float64
	CartaCOES      string
	Notas          string
}

func donacionesInit(db *sql.DB, cuenta string, periodo int) ([]Donacion, error) {
	var donaciones []Donacion

	query := `
		SELECT id, emitido, cuenta_salida, partida_salida, cuenta_entrada, partida_entrada, 
		       detalle, monto_bruto, carta_coes, notas
		FROM donaciones
		WHERE cuenta_salida = ? OR cuenta_entrada = ?
		ORDER BY emitido DESC
	`

	rows, err := db.Query(query, cuenta, cuenta)
	if err != nil {
		return nil, fmt.Errorf("donacionesInit: error querying donaciones: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var d Donacion
		var notas sql.NullString

		err := rows.Scan(
			&d.ID, &d.Emitido, &d.CuentaSalida, &d.PartidaSalida, &d.CuentaEntrada, &d.PartidaEntrada,
			&d.Detalle, &d.MontoBruto, &d.CartaCOES, &notas,
		)
		if err != nil {
			return nil, fmt.Errorf("donacionesInit: error scanning row: %w", err)
		}

		if d.Emitido.Year() == periodo {
			d.Notas = notas.String
			donaciones = append(donaciones, d)
		}
	}

	return donaciones, nil
}
