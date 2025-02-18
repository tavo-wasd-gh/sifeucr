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
	// Runtime
	//     COES
	ServiciosPendientesCOES   []Servicio
	SuministrosPendientesCOES []Suministros
	BienesPendientesCOES      []Bien
	DonacionesPendientesCOES  []Donacion
	//     SF
	ServiciosPendientesGECO   []Servicio
	SuministrosPendientesGECO []Suministros
	BienesPendientesGECO      []Bien
	//     SF
	ServiciosPendientesOCS []Servicio
	BienesPendientesOC     []Bien
	//     SF
	ServiciosPendientesDist []Servicio
	BienesPendientesDist    []Bien
	//     SF
	ServiciosPendientesEj   []Servicio
	SuministrosPendientesRe []Suministros
	BienesPendientesRe      []Bien
	//     CC
	AjustesCC []Ajuste
}

func Login(db *sql.DB, u, c string) (*Usuario, error) {
	usuario, err := UsuarioAcreditado(db, u, c)
	if err != nil {
		return nil, fmt.Errorf("Login: unauthorized: %v", err)
	}

	periodoActual := time.Now().Year()

	presupuestos, err := presupuestosInit(db, c, periodoActual)
	if err != nil {
		return nil, fmt.Errorf("Login: failed to init presupuestos: %v", err)
	}
	usuario.Cuenta.Presupuestos = presupuestos

	servicios, err := serviciosInit(db, c, periodoActual)
	if err != nil {
		return nil, fmt.Errorf("Login: failed to init servicios: %v", err)
	}
	usuario.Cuenta.Servicios = servicios

	suministros, err := suministrosInit(db, c, periodoActual)
	if err != nil {
		return nil, fmt.Errorf("Login: failed to init suministros: %v", err)
	}
	usuario.Cuenta.Suministros = suministros

	bienes, err := bienesInit(db, c, periodoActual)
	if err != nil {
		return nil, fmt.Errorf("Login: failed to init bienes: %v", err)
	}
	usuario.Cuenta.Bienes = bienes

	ajustes, err := ajustesInit(db, c, periodoActual)
	if err != nil {
		return nil, fmt.Errorf("Login: failed to init ajustes: %v", err)
	}
	usuario.Cuenta.Ajustes = ajustes

	donaciones, err := donacionesInit(db, c, periodoActual)
	if err != nil {
		return nil, fmt.Errorf("Login: failed to init donaciones: %v", err)
	}
	usuario.Cuenta.Donaciones = donaciones

	if usuario.Cuenta.ID == "COES" {
		tServ, err := ServiciosPendientesCOES(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load COES: %w", err)
		}
		usuario.ServiciosPendientesCOES = tServ

		tSum, err := SuministrosPendientesCOES(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load COES: %w", err)
		}
		usuario.SuministrosPendientesCOES = tSum

		tBien, err := BienesPendientesCOES(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load COES: %w", err)
		}
		usuario.BienesPendientesCOES = tBien

		tDona, err := DonacionesPendientesCOES(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load COES: %w", err)
		}
		usuario.DonacionesPendientesCOES = tDona
	} else if usuario.Cuenta.ID == "SF" {
		usuario.ServiciosPendientesGECO, err = ServiciosPendientesGECO(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load SF: %w", err)
		}

		usuario.SuministrosPendientesGECO, err = SuministrosPendientesGECO(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load SF: %w", err)
		}

		usuario.BienesPendientesGECO, err = BienesPendientesGECO(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load SF: %w", err)
		}

		usuario.ServiciosPendientesOCS, err = ServiciosPendientesOCS(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load SF: %w", err)
		}

		usuario.BienesPendientesOC, err = BienesPendientesOC(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load SF: %w", err)
		}

		usuario.ServiciosPendientesDist, err = ServiciosPendientesDist(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load SF: %w", err)
		}

		usuario.BienesPendientesDist, err = BienesPendientesDist(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load SF: %w", err)
		}

		usuario.ServiciosPendientesEj, err = ServiciosPendientesEjecucion(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load SF: %w", err)
		}

		usuario.BienesPendientesRe, err = BienesPendientesRecepcion(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load SF: %w", err)
		}

		usuario.SuministrosPendientesRe, err = SuministrosPendientesRecepcion(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load SF: %w", err)
		}
	} else if usuario.Cuenta.ID == "CC" {
		usuario.AjustesCC, err = AjustesCC(db, periodoActual)
		if err != nil {
			return nil, fmt.Errorf("Login: failed to load CC: %w", err)
		}
	}

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
