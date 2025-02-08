package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func Init(connDvr, connStr string) (*sql.DB, error) {
	db, err := sql.Open(connDvr, connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func CuentasPorUsuario(db *sql.DB, correo string) ([]string, error) {
	query := `SELECT id FROM cuentas WHERE presidencia = $1 OR tesoreria = $1`

	rows, err := db.Query(query, correo)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	var cuentas []string
	for rows.Next() {
		var cuenta string
		if err := rows.Scan(&cuenta); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		cuentas = append(cuentas, cuenta)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	if len(cuentas) == 0 {
		return nil, fmt.Errorf("no matching accounts found for email: %s", correo)
	}

	return cuentas, nil
}

func ListaCuentas(db *sql.DB) ([]Cuenta, error) {
	query := `SELECT id, nombre FROM cuentas`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error querying cuentas: %w", err)
	}
	defer rows.Close()

	var cuentas []Cuenta

	for rows.Next() {
		var c Cuenta
		if err := rows.Scan(&c.ID, &c.Nombre); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		cuentas = append(cuentas, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return cuentas, nil
}

func firmasCompletas(db *sql.DB, table string, column string, id int) (bool, error) {
	query := fmt.Sprintf("SELECT firma FROM %s WHERE %s = ?", table, column)

	rows, err := db.Query(query, id)
	if err != nil {
		return false, fmt.Errorf("firmasCompletas: error querying firmas from %s: %w", table, err)
	}
	defer rows.Close()

	for rows.Next() {
		var f sql.NullString
		if err := rows.Scan(&f); err != nil {
			return false, fmt.Errorf("firmasCompletas: error scanning firma: %w", err)
		}

		if !f.Valid || f.String == "" {
			return false, nil
		}
	}

	return true, nil
}
