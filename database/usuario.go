package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Usuario struct {
	ID     string
	Nombre string
	Cuenta Cuenta
}

func Login(db *sql.DB, u, c string) (*Usuario, error) {
	usuario, err := UsuarioAcreditado(db, u, c)

	presupuestos, err := presupuestosInit(db, c, time.Now().Year())
	if err != nil {
		return nil, fmt.Errorf("Login: failed to init presupuestos: %v", err)
	}
	usuario.Cuenta.Presupuestos = presupuestos

	servicios, err := serviciosInit(db, c, time.Now().Year())
	if err != nil {
		return nil, fmt.Errorf("Login: failed to init servicios: %v", err)
	}
	usuario.Cuenta.Servicios = servicios

	suministros, err := suministrosInit(db, c)
	if err != nil {
		return nil, fmt.Errorf("Login: failed to init servicios: %v", err)
	}
	usuario.Cuenta.Suministros = suministros

	return usuario, nil
}

func usuarioInit(db *sql.DB, usuario string) (Usuario, error) {
	var u Usuario

	queryUsuario := `SELECT id, nombre FROM usuarios WHERE id = ?`
	rowUsuario := db.QueryRow(queryUsuario, usuario)

	if err := rowUsuario.Scan(
		&u.ID,
		&u.Nombre,
	); err != nil {
		return Usuario{}, fmt.Errorf("cuenta: error scanning row: %w", err)
	}

	return u, nil
}

func UsuarioAcreditado(db *sql.DB, u, c string) (*Usuario, error) {
	usuario, err := usuarioInit(db, u)
	if err != nil {
		return nil, fmt.Errorf("Login: failed to init user: %v", err)
	}

	cuenta, err := cuentaInit(db, c)
	if err != nil {
		return nil, fmt.Errorf("Login: failed to init cuenta: %v", err)
	}
	usuario.Cuenta = cuenta

	if usuario.ID != usuario.Cuenta.Presidencia &&
		usuario.ID != usuario.Cuenta.Tesoreria {
		return nil, fmt.Errorf("Login: unauthorized account")
	}

	return &usuario, nil
}
