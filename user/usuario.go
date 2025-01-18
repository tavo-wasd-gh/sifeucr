package user

import (
	"database/sql"
	"fmt"

	"github.com/tavo-wasd-gh/gosql"
)

type Usuario struct {
	ID     string `db:"id"`
	Nombre string `db:"nombre"`
	Cuenta Cuenta
}

func Login(db *sql.DB, usuarioInit, cuentaInit string) (*Usuario, error) {
	var u Usuario

	u.ID = usuarioInit
	u.Cuenta.ID = cuentaInit

	usuario := db.QueryRow(`SELECT * FROM usuarios WHERE id = ?`, u.ID)
	if err := gosql.ScanRow(usuario, u); err != nil {
		return nil, err
	}

	cuenta := db.QueryRow(`SELECT * FROM cuentas WHERE id = ?`, u.Cuenta.ID)
	if err := gosql.ScanRow(cuenta, u.Cuenta); err != nil {
		return nil, err
	}

	if u.Cuenta.Presidencia != u.ID || u.Cuenta.Tesoreria != u.ID {
		return nil, fmt.Errorf("user does not have permissions on: %v", u.Cuenta.Nombre)
	}

	found := false

	pg := db.QueryRow(`SELECT * FROM presupuestos WHERE id = ?`, u.Cuenta.PGID)
	if err := gosql.ScanRow(pg, &u.Cuenta.PG); err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to fetch PG: %v", err)
		}
	} else {
		found = true
	}

	p1 := db.QueryRow(`SELECT * FROM presupuestos WHERE id = ?`, u.Cuenta.P1ID)
	if err := gosql.ScanRow(p1, &u.Cuenta.P1); err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to fetch P1: %v", err)
		}
	} else {
		found = true
	}

	p2 := db.QueryRow(`SELECT * FROM presupuestos WHERE id = ?`, u.Cuenta.P2ID)
	if err := gosql.ScanRow(p2, &u.Cuenta.P2); err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to fetch P2: %v", err)
		}
	} else {
		found = true
	}

	if !found {
		return nil, fmt.Errorf("missing either PG, P1, or P2 from ID: %s", u.Cuenta.ID)
	}

	// Falta calcular los structs de solicitudes

	return &u, nil
}
