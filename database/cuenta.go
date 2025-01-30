package database

import (
	"database/sql"
	"fmt"
)

type Cuenta struct {
	ID string
	Privilegio uint64
	Nombre string
	Presidencia string
	Tesoreria string
	TEEU string
	COES string
	// Runtime
	Presupuestos []Presupuesto
	Servicios []Servicio
	// Suministros []Suministros
	// Bienes []Bien
}

func cuentaInit(db *sql.DB, cuenta string) (Cuenta, error) {
	query := `SELECT id, privilegio, nombre, presidencia, tesoreria, teeu, coes
	FROM cuentas WHERE id = ?`

	row := db.QueryRow(query, cuenta)

	var c Cuenta

	if err := row.Scan(
		&c.ID,
		&c.Privilegio,
		&c.Nombre,
		&c.Presidencia,
		&c.Tesoreria,
		&c.TEEU,
		&c.COES,
	); err != nil {
		return Cuenta{}, fmt.Errorf("cuentaInit: error scanning row: %w", err)
	}

	return c, nil
}
