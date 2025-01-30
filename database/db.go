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

func CuentasActivas(db *sql.DB, correo string) ([]string, error) {
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
