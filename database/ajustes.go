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
	CuentaEmisora string
	Cuenta    string
	Partida   string
	Presupuesto string
	Detalle   string
	MontoBruto float64
	Notas     string
	// Runtime
	UsuarioLoggeado string
	CuentaLoggeada string
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

func NuevoAjuste(db *sql.DB, ajuste Ajuste) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("NuevoAjuste: failed to begin transaction: %w", err)
	}

	presupuestoID, err := presupuestoActual(db, ajuste.Cuenta)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("NuevoAjuste: failed to fetch presupuesto: %w", err)
	}

	_, err = tx.Exec(`
		INSERT INTO ajustes (emitido, emisor, cuenta_emisora, cuenta, presupuesto, partida, detalle, monto_bruto) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		ajuste.Emitido, ajuste.Emisor, ajuste.CuentaEmisora, ajuste.Cuenta, presupuestoID, ajuste.Partida, ajuste.Detalle, ajuste.MontoBruto,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("NuevoAjuste: failed to insert ajuste: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("NuevoAjuste: failed to commit transaction: %w", err)
	}

	return nil
}

func LeerAjuste(db *sql.DB, id, cuenta string) (Ajuste, error) {
	var a Ajuste
	var notas sql.NullString

	err := db.QueryRow(`
		SELECT id, emitido, emisor, cuenta_emisora, cuenta, presupuesto, partida, 
		detalle, monto_bruto, notas
		FROM ajustes WHERE id = ?`, id).
		Scan(
			&a.ID, &a.Emitido, &a.Emisor, &a.CuentaEmisora, &a.Cuenta, &a.Presupuesto,
			&a.Partida, &a.Detalle, &a.MontoBruto, &notas,
		)

	if err != nil {
		if err == sql.ErrNoRows {
			return Ajuste{}, fmt.Errorf("LeerAjuste: ajuste con ID '%s' no encontrado", id)
		}
		return Ajuste{}, fmt.Errorf("LeerAjuste: error al obtener ajuste: %w", err)
	}

	a.Notas = notas.String

	if a.CuentaEmisora != cuenta && a.Cuenta != cuenta {
		return Ajuste{}, fmt.Errorf("LeerAjuste: cuenta '%s' no encontrada en el ajuste", cuenta)
	}

	return a, nil
}
