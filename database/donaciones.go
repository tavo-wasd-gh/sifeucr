package database

import (
	"time"
	"fmt"
	"database/sql"
)

type Donacion struct {
	ID                 int
	Emitido            time.Time
	Emisor             string
	Cuenta             string
	CuentaSalida       string
	PresupuestoSalida  string
	PartidaSalida      string
	CuentaEntrada      string
	PresupuestoEntrada string
	PartidaEntrada     string
	Detalle            string
	MontoBruto         float64
	CartaCOES          string
	Notas              string
	// Runtime
	UsuarioLoggeado string
	CuentaLoggeada  string
}

func donacionesInit(db *sql.DB, cuenta string, periodo int) ([]Donacion, error) {
	var donaciones []Donacion

	query := `
		SELECT id, emitido, emisor, cuenta, cuenta_salida, partida_salida, cuenta_entrada, partida_entrada, 
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
		var cartaCOES, notas sql.NullString

		err := rows.Scan(
			&d.ID, &d.Emitido, &d.Emisor, &d.Cuenta, &d.CuentaSalida, &d.PartidaSalida, &d.CuentaEntrada, &d.PartidaEntrada,
			&d.Detalle, &d.MontoBruto, &cartaCOES, &notas,
		)
		if err != nil {
			return nil, fmt.Errorf("donacionesInit: error scanning row: %w", err)
		}

		if d.Emitido.Year() == periodo {
			d.Notas = notas.String
			d.CartaCOES = cartaCOES.String
			donaciones = append(donaciones, d)
		}
	}

	return donaciones, nil
}

func NuevoDonacion(db *sql.DB, donacion Donacion) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("NuevoDonacion: failed to begin transaction: %w", err)
	}

	ps, err := presupuestoActual(db, donacion.CuentaSalida)
	if err != nil {
		return fmt.Errorf("NuevoDonacion: failed to get presupuestoActual: %w", err)
	}

	pe, err := presupuestoActual(db, donacion.CuentaEntrada)
	if err != nil {
		return fmt.Errorf("NuevoDonacion: failed to get presupuestoActual: %w", err)
	}

	donacion.PresupuestoSalida = ps
	donacion.PresupuestoEntrada = pe

	_, err = tx.Exec(
		`INSERT INTO donaciones (emitido, emisor, cuenta, cuenta_salida, presupuesto_salida, partida_salida, cuenta_entrada, presupuesto_entrada, partida_entrada, detalle, monto_bruto) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		donacion.Emitido, donacion.Emisor, donacion.Cuenta, donacion.CuentaSalida, donacion.PresupuestoSalida, donacion.PartidaSalida, donacion.CuentaEntrada, donacion.PresupuestoEntrada, donacion.PartidaEntrada, donacion.Detalle, donacion.MontoBruto,
	)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("NuevoDonacion: failed to insert donacion: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("NuevoDonacion: failed to commit transaction: %w", err)
	}

	return nil
}

func LeerDonacion(db *sql.DB, id, cuenta string) (Donacion, error) {
	var d Donacion
	var cartaCOES, notas sql.NullString

	err := db.QueryRow(`
		SELECT id, emitido, emisor, cuenta, cuenta_salida, presupuesto_salida, partida_salida, cuenta_entrada, presupuesto_entrada, partida_entrada, 
		detalle, monto_bruto, carta_coes, notas 
		FROM donaciones WHERE id = ?`, id).
		Scan(
			&d.ID, &d.Emitido, &d.Emisor, &d.Cuenta,
			&d.CuentaSalida, &d.PresupuestoSalida, &d.PartidaSalida,
			&d.CuentaEntrada, &d.PresupuestoEntrada, &d.PartidaEntrada,
			&d.Detalle, &d.MontoBruto, &cartaCOES, &notas,
		)

	if err != nil {
		if err == sql.ErrNoRows {
			return Donacion{}, fmt.Errorf("LeerDonacion: donación con ID '%s' no encontrada", id)
		}
		return Donacion{}, fmt.Errorf("LeerDonacion: error al obtener donación: %w", err)
	}

	d.Notas = notas.String
	d.CartaCOES = cartaCOES.String

	if d.CuentaSalida != cuenta && d.CuentaEntrada != cuenta && cuenta != "COES" && cuenta != "SF" {
		return Donacion{}, fmt.Errorf("LeerDonacion: cuenta '%s' no encontrada en la donación", cuenta)
	}

	return d, nil
}

func DonacionesPendientesCOES(db *sql.DB, periodo int) ([]Donacion, error) {
	query := `
		SELECT id, emitido, emisor, cuenta, cuenta_salida, presupuesto_salida, partida_salida, 
		       cuenta_entrada, presupuesto_entrada, partida_entrada, detalle, monto_bruto, 
		       carta_coes, notas
		FROM donaciones
		WHERE carta_coes IS NULL OR carta_coes = ''
		ORDER BY emitido DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("DonacionesPendientesCOES: error fetching donaciones: %w", err)
	}
	defer rows.Close()

	var donaciones []Donacion

	for rows.Next() {
		var d Donacion
		var emitido time.Time
		var cartaCOES, notas sql.NullString

		if err := rows.Scan(
			&d.ID, &emitido, &d.Emisor, &d.Cuenta, &d.CuentaSalida, &d.PresupuestoSalida, &d.PartidaSalida,
			&d.CuentaEntrada, &d.PresupuestoEntrada, &d.PartidaEntrada, &d.Detalle, &d.MontoBruto,
			&cartaCOES, &notas,
		); err != nil {
			return nil, fmt.Errorf("DonacionesPendientesCOES: error scanning row: %w", err)
		}

		if emitido.Year() == periodo {
			d.Emitido = emitido
			d.CartaCOES = cartaCOES.String
			d.Notas = notas.String

			donaciones = append(donaciones, d)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("DonacionesPendientesCOES: error iterating rows: %w", err)
	}

	return donaciones, nil
}

func AprobarDonacionCOES(db *sql.DB, id string) error {
	_, err := db.Exec(`UPDATE donaciones SET carta_coes = TRUE WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("AprobarServicioCOES: failed to update service: %w", err)
	}
	return nil
}
