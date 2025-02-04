package database

import (
	"time"
	"fmt"
	"database/sql"
)

type Ajuste struct {
	ID        int
	Emitido   time.Time
	Emisor    string
	Cuenta    string
	Partida   string
	Detalle   string
	MontoBruto float64
	Notas     string
}

func ajustesInit(db *sql.DB, cuenta string, periodo int) ([]Ajuste, error) {
	var ajustes []Ajuste

	query := `
		SELECT id, emitido, emisor, cuenta, partida, detalle, monto_bruto, notas
		FROM ajustes
		WHERE cuenta = ?
		ORDER BY emitido DESC
	`

	rows, err := db.Query(query, cuenta)
	if err != nil {
		return nil, fmt.Errorf("ajustesInit: error querying ajustes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var a Ajuste
		var notas sql.NullString

		err := rows.Scan(
			&a.ID, &a.Emitido, &a.Emisor, &a.Cuenta, &a.Partida, &a.Detalle, &a.MontoBruto, &notas,
		)
		if err != nil {
			return nil, fmt.Errorf("ajustesInit: error scanning row: %w", err)
		}

		if a.Emitido.Year() == periodo {
			a.Notas = notas.String
			ajustes = append(ajustes, a)
		}
	}

	return ajustes, nil
}
