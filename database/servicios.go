package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Servicio struct {
	ID int
	// Solicitud
	Emitido     time.Time
	Emisor      string
	Detalle     string
	PorEjecutar time.Time
	Justif      string
	// COES
	COES bool
	// OSUM
	ProvNom    sql.NullString
	ProvCed    sql.NullString
	ProvDirec  sql.NullString
	ProvEmail  sql.NullString
	ProvTel    sql.NullString
	ProvBanco  sql.NullString
	ProvIBAN   sql.NullString
	ProvJustif sql.NullString
	MontoBruto sql.NullFloat64
	MontoIVA   sql.NullFloat64
	MontoDesc  sql.NullFloat64
	GecoSol    sql.NullString
	GecoOCS    sql.NullString
	// ViVE
	OCSFirma     sql.NullString
	OCSFirmaVive sql.NullString
	// Ejecutado
	AcuseUsuario sql.NullString
	AcuseFecha   sql.NullTime
	Acuse        sql.NullString
	AcuseFirma   sql.NullString
	// Final
	Pagado sql.NullTime
	Notas  sql.NullString
	// Runtime
	Movimientos     []ServicioMovimiento
	FirmasCompletas bool
	UsuarioLoggeado string
	CuentaLoggeada  string
}

type ServicioMovimiento struct {
	ID          int
	Servicio    int
	Usuario     sql.NullString
	Cuenta      string
	Presupuesto string
	Monto       sql.NullFloat64
	Firma       sql.NullString
}

func serviciosInit(db *sql.DB, cuenta string, periodo int) ([]Servicio, error) {
	query := `SELECT
	s.id,
	s.emitido,s.emisor,s.detalle,s.por_ejecutar,s.justif,
	s.coes,
	s.prov_nom,s.prov_ced,s.prov_direc,s.prov_email,s.prov_tel,s.prov_banco,s.prov_iban,s.prov_justif,s.monto_bruto,s.monto_iva,s.monto_desc,s.geco_sol,s.geco_ocs,
	s.ocs_firma,s.ocs_firma_vive,
	s.acuse_usuario, s.acuse_fecha, s.acuse, s.acuse_firma,
	s.pagado, s.notas
	FROM servicios s
	JOIN servicios_movimientos sm
	ON s.id = sm.servicio
	JOIN presupuestos p
	ON sm.presupuesto = p.id
	JOIN cuentas c
	ON p.cuenta = c.id
	WHERE c.id = ?
	ORDER BY s.emitido;`

	rows, err := db.Query(query, cuenta)
	if err != nil {
		return nil, fmt.Errorf("serviciosInit: error querying database: %w", err)
	}
	defer rows.Close()

	var servicios []Servicio

	for rows.Next() {
		var s Servicio
		if err := rows.Scan(
			&s.ID,
			&s.Emitido, &s.Emisor, &s.Detalle, &s.PorEjecutar, &s.Justif,
			&s.COES,
			&s.ProvNom, &s.ProvCed, &s.ProvDirec, &s.ProvEmail, &s.ProvTel, &s.ProvBanco, &s.ProvIBAN, &s.ProvJustif, &s.MontoBruto, &s.MontoIVA, &s.MontoDesc, &s.GecoSol, &s.GecoOCS,
			&s.OCSFirma, &s.OCSFirmaVive,
			&s.AcuseUsuario, &s.AcuseFecha, &s.Acuse, &s.AcuseFirma,
			&s.Pagado, &s.Notas,
		); err != nil {
			return nil, fmt.Errorf("serviciosInit: error scanning row: %w", err)
		}

		validez := s.Emitido.Year()
		if validez == periodo {
			s.FirmasCompletas, err = firmasCompletas(db, "servicios_movimientos", "servicio", s.ID)
			if err != nil {
				return nil, err
			}

			servicios = append(servicios, s)
		}
	}

	return servicios, nil
}

func NuevoServicio(db *sql.DB, servicio Servicio, movimientos []ServicioMovimiento) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("NuevoServicio: failed to begin transaction: %w", err)
	}

	var servicioID int
	err = tx.QueryRow(`
		INSERT INTO servicios (emitido, emisor, detalle, por_ejecutar, justif, coes) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		time.Now(), servicio.Emisor, servicio.Detalle, servicio.PorEjecutar, servicio.Justif, false,
		).Scan(&servicioID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("NuevoServicio: failed to insert servicio: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO servicios_movimientos (servicio, usuario, cuenta, presupuesto, firma) 
		VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("NuevoServicio: failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, mov := range movimientos {
		presupuestoID, err := presupuestoActual(db, mov.Cuenta)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("NuevoServicio: failed to fetch presupuesto for movimiento %d: %w", i+1, err)
		}

		_, err = stmt.Exec(servicioID, mov.Usuario, mov.Cuenta, presupuestoID, mov.Firma)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("NuevoServicio: failed to insert servicio_movimiento: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("NuevoServicio: failed to commit transaction: %w", err)
	}

	return nil
}

func LeerServicio(db *sql.DB, id, cuenta string) (Servicio, error) {
	var s Servicio
	err := db.QueryRow(`
		SELECT id, emitido, emisor, detalle, por_ejecutar, justif, coes,
		prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, prov_iban, prov_justif,
		monto_bruto, monto_iva, monto_desc, geco_sol, geco_ocs, 
		ocs_firma, ocs_firma_vive, acuse_usuario, acuse_fecha, acuse, acuse_firma,
		pagado, notas
		FROM servicios WHERE id = ?`, id).
		Scan(
			&s.ID, &s.Emitido, &s.Emisor, &s.Detalle, &s.PorEjecutar, &s.Justif, &s.COES,
			&s.ProvNom, &s.ProvCed, &s.ProvDirec, &s.ProvEmail, &s.ProvTel, &s.ProvBanco, &s.ProvIBAN, &s.ProvJustif,
			&s.MontoBruto, &s.MontoIVA, &s.MontoDesc, &s.GecoSol, &s.GecoOCS,
			&s.OCSFirma, &s.OCSFirmaVive, &s.AcuseUsuario, &s.AcuseFecha, &s.Acuse, &s.AcuseFirma,
			&s.Pagado, &s.Notas,
		)
	if err != nil {
		if err == sql.ErrNoRows {
			return Servicio{}, fmt.Errorf("LeerServicio: servicio con ID '%s' no encontrado", id)
		}
		return Servicio{}, fmt.Errorf("LeerServicio: error al obtener servicio: %w", err)
	}

	rows, err := db.Query("SELECT id, servicio, usuario, cuenta, presupuesto, monto, firma FROM servicios_movimientos WHERE servicio = ?", id)
	if err != nil {
		return Servicio{}, fmt.Errorf("LeerServicio: error al obtener movimientos: %w", err)
	}
	defer rows.Close()

	var movimientos []ServicioMovimiento
	found := false
	firmasCompletas := true

	for rows.Next() {
		var m ServicioMovimiento
		if err := rows.Scan(&m.ID, &m.Servicio, &m.Usuario, &m.Cuenta, &m.Presupuesto, &m.Monto, &m.Firma); err != nil {
			return Servicio{}, fmt.Errorf("LeerServicio: error al escanear movimientos: %w", err)
		}
		movimientos = append(movimientos, m)

		if m.Firma.String == "" {
			firmasCompletas = false
		}

		if m.Cuenta == cuenta {
			found = true
		}
	}

	s.FirmasCompletas = firmasCompletas

	if err := rows.Err(); err != nil {
		return Servicio{}, fmt.Errorf("LeerServicio: error al recorrer movimientos: %w", err)
	}

	s.Movimientos = movimientos

	if !found {
		return Servicio{}, fmt.Errorf("LeerServicio: cuenta '%s' no encontrada en participantes", cuenta)
	}

	return s, nil
}

func FirmarMovimientoServicios(db *sql.DB, id, usuario, cuenta, firma string) error {
	_, err := UsuarioAcreditado(db, usuario, cuenta)
	if err != nil {
		return fmt.Errorf("FirmarMovimientoServicios: error al iniciar usuario: %w", err)
	}

	var existingCuenta string
	err = db.QueryRow("SELECT cuenta FROM servicios_movimientos WHERE id = ?", id).Scan(&existingCuenta)
	if err != nil {
		return fmt.Errorf("FirmarMovimientoServicios: error retrieving cuenta for id %s: %w", id, err)
	}
	if existingCuenta != cuenta {
		return fmt.Errorf("FirmarMovimientoServicios: cuenta mismatch for id %s (expected: %s, got: %s)", id, existingCuenta, cuenta)
	}

	query := `UPDATE servicios_movimientos
	SET usuario = ?, firma = ?
	WHERE id = ?;`

	if _, err = db.Exec(query, usuario, firma, id) ; err != nil {
		return fmt.Errorf("FirmarMovimientoServicios: failed to update servicio_movimiento with id %s: %w", id, err)
	}

	return nil
}

func ConfirmarEjecutadoServicios(db *sql.DB, id, usuario, cuenta string, fecha time.Time, acuse, firma string) error {
	now := time.Now()
	oneMonthAgo := now.AddDate(0, -1, 0)

	if fecha.After(now) || fecha.Before(oneMonthAgo) {
		return fmt.Errorf("ConfirmarEjecutadoServicios: invalid date")
	}

	_, err := UsuarioAcreditado(db, usuario, cuenta)
	if err != nil {
		return fmt.Errorf("ConfirmarEjecutadoServicios: usuario %s no acreditado para cuenta %s: %w", usuario, cuenta, err)
	}

	var servicioID int
	err = db.QueryRow("SELECT servicio FROM servicios_movimientos WHERE servicio = ? AND cuenta = ?", id, cuenta).Scan(&servicioID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("ConfirmarEjecutadoServicios: no matching servicio_movimiento found for id %s and cuenta %s", id, cuenta)
		}
		return fmt.Errorf("ConfirmarEjecutadoServicios: error retrieving servicio for id %s: %w", id, err)
	}

	query := `UPDATE servicios
		SET acuse_usuario = ?, acuse_fecha = ?, acuse = ?, acuse_firma = ?
		WHERE id = ?;`

	_, err = db.Exec(query, usuario, fecha, acuse, firma, servicioID)
	if err != nil {
		return fmt.Errorf("ConfirmarEjecutadoServicios: failed to update servicio with id %d: %w", servicioID, err)
	}

	return nil
}
