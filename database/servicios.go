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
	ProvNom    string
	ProvCed    string
	ProvDirec  string
	ProvEmail  string
	ProvTel    string
	ProvBanco  string
	ProvIBAN   string
	ProvJustif string
	MontoBruto float64
	MontoIVA   float64
	MontoDesc  float64
	GecoSol    string
	GecoOCS    string
	// ViVE
	OCSFirma     string
	OCSFirmaVive string
	// Ejecutado
	AcuseUsuario string
	AcuseFecha   time.Time
	Acuse        string
	AcuseFirma   string
	// Final
	Pagado time.Time
	Notas  string
	// Runtime
	Movimientos     []ServicioMovimiento
	FirmasCompletas bool
	UsuarioLoggeado string
	CuentaLoggeada  string
}

type ServicioMovimiento struct {
	ID          int
	Servicio    int
	Usuario     string
	Cuenta      string
	Presupuesto string
	Monto       float64
	Firma       string
}

func serviciosInit(db *sql.DB, cuenta string, periodo int) ([]Servicio, error) {
	query := `
		SELECT s.id, s.emitido, s.emisor, s.detalle, s.por_ejecutar, s.justif, s.coes,
		       s.prov_nom, s.prov_ced, s.prov_direc, s.prov_email, s.prov_tel,
		       s.prov_banco, s.prov_iban, s.prov_justif, s.monto_bruto, s.monto_iva, s.monto_desc, 
		       s.geco_sol, s.geco_ocs, s.ocs_firma, s.ocs_firma_vive,
		       s.acuse_usuario, s.acuse_fecha, s.acuse, s.acuse_firma,
		       s.pagado, s.notas
		FROM servicios s
		JOIN servicios_movimientos sm ON s.id = sm.servicio
		JOIN presupuestos p ON sm.presupuesto = p.id
		JOIN cuentas c ON p.cuenta = c.id
		WHERE c.id = ?
		ORDER BY s.emitido DESC;`

	rows, err := db.Query(query, cuenta)
	if err != nil {
		return nil, fmt.Errorf("serviciosInit: error querying database: %w", err)
	}
	defer rows.Close()

	var servicios []Servicio

	for rows.Next() {
		var s Servicio
		var acuseFecha, pagado sql.NullTime
		var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOCS sql.NullString
		var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
		var montoBruto, montoIVA, montoDesc sql.NullFloat64
		var ocsFirma, ocsFirmaVive sql.NullString
		var justif sql.NullString
		var notas sql.NullString

		if err := rows.Scan(
			&s.ID, &s.Emitido, &s.Emisor, &s.Detalle, &s.PorEjecutar, &justif, &s.COES,
			&provNom, &provCed, &provDirec, &provEmail, &provTel,
			&provBanco, &provIBAN, &provJustif, &montoBruto, &montoIVA, &montoDesc,
			&gecoSol, &gecoOCS, &ocsFirma, &ocsFirmaVive,
			&acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		); err != nil {
			return nil, fmt.Errorf("serviciosInit: error scanning row: %w", err)
		}

		s.Pagado = pagado.Time
		s.AcuseFecha = acuseFecha.Time
		s.AcuseUsuario = acuseUsuario.String
		s.Acuse = acuse.String
		s.AcuseFirma = acuseFirma.String
		s.GecoSol = gecoSol.String
		s.GecoOCS = gecoOCS.String
		s.OCSFirma = ocsFirma.String
		s.OCSFirmaVive = ocsFirmaVive.String
		s.ProvNom = provNom.String
		s.ProvCed = provCed.String
		s.ProvDirec = provDirec.String
		s.ProvEmail = provEmail.String
		s.ProvTel = provTel.String
		s.ProvBanco = provBanco.String
		s.ProvIBAN = provIBAN.String
		s.ProvJustif = provJustif.String
		s.MontoBruto = montoBruto.Float64
		s.MontoIVA = montoIVA.Float64
		s.MontoDesc = montoDesc.Float64
		s.Justif = justif.String
		s.Notas = notas.String

		if s.Emitido.Year() == periodo {
			s.Movimientos, err = servicioMovimientosInit(db, s.ID)
			if err != nil {
				return nil, fmt.Errorf("serviciosInit: error fetching movimientos for servicio %d: %w", s.ID, err)
			}

			s.FirmasCompletas, err = firmasCompletas(db, "servicios_movimientos", "servicio", s.ID)
			if err != nil {
				return nil, err
			}

			servicios = append(servicios, s)
		}
	}

	return servicios, nil
}

func servicioMovimientosInit(db *sql.DB, servicioID int) ([]ServicioMovimiento, error) {
	var movimientos []ServicioMovimiento

	query := `
		SELECT id, servicio, usuario, cuenta, presupuesto, monto, firma
		FROM servicios_movimientos
		WHERE servicio = ?
	`

	rows, err := db.Query(query, servicioID)
	if err != nil {
		return nil, fmt.Errorf("servicioMovimientosInit: error querying movimientos: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m ServicioMovimiento
		var firma sql.NullString
		var usuario sql.NullString
		var monto sql.NullFloat64

		err := rows.Scan(
			&m.ID, &m.Servicio, &usuario, &m.Cuenta, &m.Presupuesto, &monto, &firma,
		)
		if err != nil {
			return nil, fmt.Errorf("servicioMovimientosInit: error scanning row: %w", err)
		}

		m.Usuario = usuario.String
		m.Monto = monto.Float64
		m.Firma = firma.String

		movimientos = append(movimientos, m)
	}

	return movimientos, nil
}

func NuevoServicio(db *sql.DB, servicio Servicio) error {
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

	for i, mov := range servicio.Movimientos {
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
	var acuseFecha, pagado sql.NullTime
	var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOCS sql.NullString
	var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
	var montoBruto, montoIVA, montoDesc sql.NullFloat64
	var ocsFirma, ocsFirmaVive sql.NullString
	var justif sql.NullString
	var notas sql.NullString

	err := db.QueryRow(`
		SELECT id, emitido, emisor, detalle, por_ejecutar, justif, coes,
		prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, prov_iban, prov_justif,
		monto_bruto, monto_iva, monto_desc, geco_sol, geco_ocs, 
		ocs_firma, ocs_firma_vive, acuse_usuario, acuse_fecha, acuse, acuse_firma,
		pagado, notas
		FROM servicios WHERE id = ?`, id).
		Scan(
			&s.ID, &s.Emitido, &s.Emisor, &s.Detalle, &s.PorEjecutar, &justif, &s.COES,
			&provNom, &provCed, &provDirec, &provEmail, &provTel, &provBanco, &provIBAN, &provJustif,
			&montoBruto, &montoIVA, &montoDesc, &gecoSol, &gecoOCS,
			&ocsFirma, &ocsFirmaVive, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		)

	if err != nil {
		if err == sql.ErrNoRows {
			return Servicio{}, fmt.Errorf("LeerServicio: servicio con ID '%s' no encontrado", id)
		}
		return Servicio{}, fmt.Errorf("LeerServicio: error al obtener servicio: %w", err)
	}

	s.Pagado = pagado.Time
	s.AcuseFecha = acuseFecha.Time
	s.AcuseUsuario = acuseUsuario.String
	s.Acuse = acuse.String
	s.AcuseFirma = acuseFirma.String
	s.GecoSol = gecoSol.String
	s.GecoOCS = gecoOCS.String
	s.OCSFirma = ocsFirma.String
	s.OCSFirmaVive = ocsFirmaVive.String
	s.ProvNom = provNom.String
	s.ProvCed = provCed.String
	s.ProvDirec = provDirec.String
	s.ProvEmail = provEmail.String
	s.ProvTel = provTel.String
	s.ProvBanco = provBanco.String
	s.ProvIBAN = provIBAN.String
	s.ProvJustif = provJustif.String
	s.MontoBruto = montoBruto.Float64
	s.MontoIVA = montoIVA.Float64
	s.MontoDesc = montoDesc.Float64
	s.Justif = justif.String
	s.Notas = notas.String

	rows, err := db.Query(`
		SELECT id, servicio, usuario, cuenta, presupuesto, monto, firma 
		FROM servicios_movimientos 
		WHERE servicio = ?`, id)
	if err != nil {
		return Servicio{}, fmt.Errorf("LeerServicio: error al obtener movimientos: %w", err)
	}
	defer rows.Close()

	var movimientos []ServicioMovimiento
	found := false
	firmasCompletas := true

	for rows.Next() {
		var m ServicioMovimiento
		var firma sql.NullString
		var usuario sql.NullString
		var monto sql.NullFloat64

		if err := rows.Scan(&m.ID, &m.Servicio, &usuario, &m.Cuenta, &m.Presupuesto, &monto, &firma); err != nil {
			return Servicio{}, fmt.Errorf("LeerServicio: error al escanear movimientos: %w", err)
		}

		m.Usuario = usuario.String
		m.Monto = monto.Float64
		m.Firma = firma.String

		movimientos = append(movimientos, m)

		if m.Firma == "" {
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

	if !found && cuenta != "COES" && cuenta != "SF" {
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

	if _, err = db.Exec(query, usuario, firma, id); err != nil {
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

func ServiciosPendientesCOES(db *sql.DB, periodo int) ([]Servicio, error) {
	query := `
		SELECT id, emitido, emisor, detalle, por_ejecutar, justif, 
		       prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, 
		       prov_iban, prov_justif, monto_bruto, monto_iva, monto_desc, 
		       geco_sol, geco_ocs, ocs_firma, ocs_firma_vive, 
		       acuse_usuario, acuse_fecha, acuse, acuse_firma, 
		       pagado, notas
		FROM servicios
		WHERE coes = FALSE
		ORDER BY emitido
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ServiciosPendientesCOES: error fetching servicios: %w", err)
	}
	defer rows.Close()

	var servicios []Servicio

	for rows.Next() {
		var s Servicio
		var emitido time.Time
		var acuseFecha, pagado sql.NullTime
		var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOCS sql.NullString
		var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
		var montoBruto, montoIVA, montoDesc sql.NullFloat64
		var ocsFirma, ocsFirmaVive sql.NullString
		var justif, notas sql.NullString

		if err := rows.Scan(
			&s.ID, &emitido, &s.Emisor, &s.Detalle, &s.PorEjecutar, &justif,
			&provNom, &provCed, &provDirec, &provEmail, &provTel, &provBanco, &provIBAN, &provJustif,
			&montoBruto, &montoIVA, &montoDesc, &gecoSol, &gecoOCS,
			&ocsFirma, &ocsFirmaVive, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		); err != nil {
			return nil, fmt.Errorf("ServiciosPendientesCOES: error scanning row: %w", err)
		}

		s.FirmasCompletas, err = firmasCompletas(db, "servicios_movimientos", "servicio", s.ID)
		if err != nil {
			return nil, err
		}

		if emitido.Year() == periodo && s.FirmasCompletas {
			s.Emitido = emitido
			s.Justif = justif.String
			s.ProvNom = provNom.String
			s.ProvCed = provCed.String
			s.ProvDirec = provDirec.String
			s.ProvEmail = provEmail.String
			s.ProvTel = provTel.String
			s.ProvBanco = provBanco.String
			s.ProvIBAN = provIBAN.String
			s.ProvJustif = provJustif.String
			s.MontoBruto = montoBruto.Float64
			s.MontoIVA = montoIVA.Float64
			s.MontoDesc = montoDesc.Float64
			s.GecoSol = gecoSol.String
			s.GecoOCS = gecoOCS.String
			s.OCSFirma = ocsFirma.String
			s.OCSFirmaVive = ocsFirmaVive.String
			s.AcuseUsuario = acuseUsuario.String
			s.AcuseFecha = acuseFecha.Time
			s.Acuse = acuse.String
			s.AcuseFirma = acuseFirma.String
			s.Pagado = pagado.Time
			s.Notas = notas.String

			servicios = append(servicios, s)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ServiciosPendientesCOES: error iterating rows: %w", err)
	}

	return servicios, nil
}

func AprobarServicioCOES(db *sql.DB, id string) error {
	_, err := db.Exec(`UPDATE servicios SET coes = TRUE WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("AprobarServicioCOES: failed to update service: %w", err)
	}
	return nil
}

func ServicioPorID(db *sql.DB, usuarioLoggeado, cuentaLoggeada, id string) (Servicio, error) {
	var s Servicio

	s.UsuarioLoggeado = usuarioLoggeado
	s.CuentaLoggeada = cuentaLoggeada

	var acuseFecha, pagado sql.NullTime
	var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOCS sql.NullString
	var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
	var montoBruto, montoIVA, montoDesc sql.NullFloat64
	var ocsFirma, ocsFirmaVive sql.NullString
	var justif sql.NullString
	var notas sql.NullString

	err := db.QueryRow(`
		SELECT id, emitido, emisor, detalle, por_ejecutar, justif, coes,
		prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, prov_iban, prov_justif,
		monto_bruto, monto_iva, monto_desc, geco_sol, geco_ocs, 
		ocs_firma, ocs_firma_vive, acuse_usuario, acuse_fecha, acuse, acuse_firma,
		pagado, notas
		FROM servicios WHERE id = ?`, id).
		Scan(
			&s.ID, &s.Emitido, &s.Emisor, &s.Detalle, &s.PorEjecutar, &justif, &s.COES,
			&provNom, &provCed, &provDirec, &provEmail, &provTel, &provBanco, &provIBAN, &provJustif,
			&montoBruto, &montoIVA, &montoDesc, &gecoSol, &gecoOCS,
			&ocsFirma, &ocsFirmaVive, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		)

	if err != nil {
		if err == sql.ErrNoRows {
			return Servicio{}, fmt.Errorf("LeerServicio: servicio con ID '%s' no encontrado", id)
		}
		return Servicio{}, fmt.Errorf("LeerServicio: error al obtener servicio: %w", err)
	}

	s.Pagado = pagado.Time
	s.AcuseFecha = acuseFecha.Time
	s.AcuseUsuario = acuseUsuario.String
	s.Acuse = acuse.String
	s.AcuseFirma = acuseFirma.String
	s.GecoSol = gecoSol.String
	s.GecoOCS = gecoOCS.String
	s.OCSFirma = ocsFirma.String
	s.OCSFirmaVive = ocsFirmaVive.String
	s.ProvNom = provNom.String
	s.ProvCed = provCed.String
	s.ProvDirec = provDirec.String
	s.ProvEmail = provEmail.String
	s.ProvTel = provTel.String
	s.ProvBanco = provBanco.String
	s.ProvIBAN = provIBAN.String
	s.ProvJustif = provJustif.String
	s.MontoBruto = montoBruto.Float64
	s.MontoIVA = montoIVA.Float64
	s.MontoDesc = montoDesc.Float64
	s.Justif = justif.String
	s.Notas = notas.String

	rows, err := db.Query(`
		SELECT id, servicio, usuario, cuenta, presupuesto, monto, firma 
		FROM servicios_movimientos 
		WHERE servicio = ?`, id)
	if err != nil {
		return Servicio{}, fmt.Errorf("LeerServicio: error al obtener movimientos: %w", err)
	}
	defer rows.Close()

	var movimientos []ServicioMovimiento
	found := false
	firmasCompletas := true

	for rows.Next() {
		var m ServicioMovimiento
		var firma sql.NullString
		var usuario sql.NullString
		var monto sql.NullFloat64

		if err := rows.Scan(&m.ID, &m.Servicio, &usuario, &m.Cuenta, &m.Presupuesto, &monto, &firma); err != nil {
			return Servicio{}, fmt.Errorf("LeerServicio: error al escanear movimientos: %w", err)
		}

		m.Usuario = usuario.String
		m.Monto = monto.Float64
		m.Firma = firma.String

		movimientos = append(movimientos, m)

		if m.Firma == "" {
			firmasCompletas = false
		}

		if m.Cuenta == cuentaLoggeada {
			found = true
		}
	}

	s.FirmasCompletas = firmasCompletas

	if err := rows.Err(); err != nil {
		return Servicio{}, fmt.Errorf("LeerServicio: error al recorrer movimientos: %w", err)
	}

	s.Movimientos = movimientos

	if !found && cuentaLoggeada != "COES" && cuentaLoggeada != "SF" {
		return Servicio{}, fmt.Errorf("LeerServicio: cuenta '%s' no encontrada en participantes", cuentaLoggeada)
	}

	return s, nil
}

// Registrar el número de solicitud de GECO en la base de datos
// servicio.RegistrarGECO(db, solicitud)
func (s *Servicio) RegistrarSolicitudGECO(db *sql.DB, sol string) error {
	if s.CuentaLoggeada != "SF" {
		return fmt.Errorf("RegistrarSolicitudGECO: failed to update service: unauthorized account")
	}

	_, err := db.Exec(`UPDATE servicios SET geco_sol = ? WHERE id = ?`, sol, s.ID)
	if err != nil {
		return fmt.Errorf("RegistrarSolicitudGECO: failed to update service: %w", err)
	}

	return nil
}

// Establecer la distribución de montos de un servicio segun cuentas participantes
func (s *Servicio) EstablecerMontos(db *sql.DB, montos map[string]float64) error {
	if s.MontoBruto <= 0 {
		return fmt.Errorf("EstablecerMontos: monto bruto is not yet set")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("EstablecerMontos: failed to begin transaction: %w", err)
	}

	var totalSum float64
	for _, mov := range s.Movimientos {
		if mov.Monto > 0 {
			tx.Rollback()
			return fmt.Errorf("EstablecerMontos: monto already set for movimiento ID %d", mov.ID)
		}

		monto, exists := montos[mov.Cuenta]
		if !exists {
			tx.Rollback()
			return fmt.Errorf("EstablecerMontos: cuenta %s not found in request", mov.Cuenta)
		}

		totalSum += monto

		_, err := tx.Exec(`UPDATE servicios_movimientos SET monto = ? WHERE id = ?`, monto, mov.ID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("EstablecerMontos: failed to update movimiento ID %d: %w", mov.ID, err)
		}
	}

	if totalSum != s.MontoBruto {
		tx.Rollback()
		return fmt.Errorf("EstablecerMontos: total montos (%.2f) do not match MontoBruto (%.2f)", totalSum, s.MontoBruto)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("EstablecerMontos: failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Servicio) RegistrarOCS(
	db *sql.DB, gecoOCS, provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif string,
	montoBruto, montoIVA, montoDesc float64,
) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("RegistrarOCS: failed to begin transaction: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE servicios 
		SET geco_ocs = ?, prov_nom = ?, prov_ced = ?, prov_direc = ?, prov_email = ?, prov_tel = ?, 
		    prov_banco = ?, prov_iban = ?, prov_justif = ?, monto_bruto = ?, monto_iva = ?, monto_desc = ?
		WHERE id = ?
	`, gecoOCS, provNom, provCed, provDirec, provEmail, provTel,
		provBanco, provIBAN, provJustif, montoBruto, montoIVA, montoDesc, s.ID)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("RegistrarOCS: failed to update service ID %d: %w", s.ID, err)
	}

	if len(s.Movimientos) == 1 {
		_, err = tx.Exec(`
			UPDATE servicios_movimientos
			SET monto = ? 
			WHERE servicio = ?
			`, montoBruto, s.ID)

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("RegistrarOCS: failed to update movimiento for bienes ID %d: %w", s.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("RegistrarOCS: failed to commit transaction: %w", err)
	}

	return nil
}

func ServiciosPendientesGECO(db *sql.DB, periodo int) ([]Servicio, error) {
	query := `
		SELECT id, emitido, emisor, detalle, por_ejecutar, justif, coes,
		       prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, 
		       prov_iban, prov_justif, monto_bruto, monto_iva, monto_desc, 
		       geco_sol, geco_ocs, ocs_firma, ocs_firma_vive, 
		       acuse_usuario, acuse_fecha, acuse, acuse_firma, 
		       pagado, notas
		FROM servicios
		WHERE geco_sol IS NULL
		AND coes = TRUE
		ORDER BY emitido
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ServiciosPendientesGECO: error fetching servicios: %w", err)
	}
	defer rows.Close()

	var servicios []Servicio

	for rows.Next() {
		var s Servicio
		var emitido time.Time
		var acuseFecha, pagado sql.NullTime
		var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOCS sql.NullString
		var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
		var montoBruto, montoIVA, montoDesc sql.NullFloat64
		var ocsFirma, ocsFirmaVive sql.NullString
		var justif, notas sql.NullString

		if err := rows.Scan(
			&s.ID, &emitido, &s.Emisor, &s.Detalle, &s.PorEjecutar, &justif, &s.COES,
			&provNom, &provCed, &provDirec, &provEmail, &provTel, &provBanco, &provIBAN, &provJustif,
			&montoBruto, &montoIVA, &montoDesc, &gecoSol, &gecoOCS,
			&ocsFirma, &ocsFirmaVive, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		); err != nil {
			return nil, fmt.Errorf("ServiciosPendientesGECO: error scanning row: %w", err)
		}

		s.FirmasCompletas, err = firmasCompletas(db, "servicios_movimientos", "servicio", s.ID)
		if err != nil {
			return nil, err
		}

		if emitido.Year() == periodo && s.FirmasCompletas {
			s.Emitido = emitido
			s.Justif = justif.String
			s.ProvNom = provNom.String
			s.ProvCed = provCed.String
			s.ProvDirec = provDirec.String
			s.ProvEmail = provEmail.String
			s.ProvTel = provTel.String
			s.ProvBanco = provBanco.String
			s.ProvIBAN = provIBAN.String
			s.ProvJustif = provJustif.String
			s.MontoBruto = montoBruto.Float64
			s.MontoIVA = montoIVA.Float64
			s.MontoDesc = montoDesc.Float64
			s.GecoSol = gecoSol.String
			s.GecoOCS = gecoOCS.String
			s.OCSFirma = ocsFirma.String
			s.OCSFirmaVive = ocsFirmaVive.String
			s.AcuseUsuario = acuseUsuario.String
			s.AcuseFecha = acuseFecha.Time
			s.Acuse = acuse.String
			s.AcuseFirma = acuseFirma.String
			s.Pagado = pagado.Time
			s.Notas = notas.String

			servicios = append(servicios, s)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ServiciosPendientesGECO: error iterating rows: %w", err)
	}

	return servicios, nil
}

func ServiciosPendientesOCS(db *sql.DB, periodo int) ([]Servicio, error) {
	query := `
		SELECT id, emitido, emisor, detalle, por_ejecutar, justif, coes,
		       prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, 
		       prov_iban, prov_justif, monto_bruto, monto_iva, monto_desc, 
		       geco_sol, geco_ocs, ocs_firma, ocs_firma_vive, 
		       acuse_usuario, acuse_fecha, acuse, acuse_firma, 
		       pagado, notas
		FROM servicios
		WHERE geco_ocs IS NULL
		AND geco_sol IS NOT NULL
		AND coes = TRUE
		ORDER BY emitido
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ServiciosPendientesOCS: error fetching servicios: %w", err)
	}
	defer rows.Close()

	var servicios []Servicio

	for rows.Next() {
		var s Servicio
		var emitido time.Time
		var acuseFecha, pagado sql.NullTime
		var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOCS sql.NullString
		var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
		var montoBruto, montoIVA, montoDesc sql.NullFloat64
		var ocsFirma, ocsFirmaVive sql.NullString
		var justif, notas sql.NullString

		if err := rows.Scan(
			&s.ID, &emitido, &s.Emisor, &s.Detalle, &s.PorEjecutar, &justif, &s.COES,
			&provNom, &provCed, &provDirec, &provEmail, &provTel, &provBanco, &provIBAN, &provJustif,
			&montoBruto, &montoIVA, &montoDesc, &gecoSol, &gecoOCS,
			&ocsFirma, &ocsFirmaVive, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		); err != nil {
			return nil, fmt.Errorf("ServiciosPendientesOCS: error scanning row: %w", err)
		}

		s.FirmasCompletas, err = firmasCompletas(db, "servicios_movimientos", "servicio", s.ID)
		if err != nil {
			return nil, err
		}

		if emitido.Year() == periodo && s.FirmasCompletas {
			s.Emitido = emitido
			s.Justif = justif.String
			s.ProvNom = provNom.String
			s.ProvCed = provCed.String
			s.ProvDirec = provDirec.String
			s.ProvEmail = provEmail.String
			s.ProvTel = provTel.String
			s.ProvBanco = provBanco.String
			s.ProvIBAN = provIBAN.String
			s.ProvJustif = provJustif.String
			s.MontoBruto = montoBruto.Float64
			s.MontoIVA = montoIVA.Float64
			s.MontoDesc = montoDesc.Float64
			s.GecoSol = gecoSol.String
			s.GecoOCS = gecoOCS.String
			s.OCSFirma = ocsFirma.String
			s.OCSFirmaVive = ocsFirmaVive.String
			s.AcuseUsuario = acuseUsuario.String
			s.AcuseFecha = acuseFecha.Time
			s.Acuse = acuse.String
			s.AcuseFirma = acuseFirma.String
			s.Pagado = pagado.Time
			s.Notas = notas.String

			servicios = append(servicios, s)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ServiciosPendientesOCS: error iterating rows: %w", err)
	}

	return servicios, nil
}

func ServiciosPendientesDist(db *sql.DB, periodo int) ([]Servicio, error) {
	query := `
		SELECT id, emitido, emisor, detalle, por_ejecutar, justif, coes,
		       prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, 
		       prov_iban, prov_justif, monto_bruto, monto_iva, monto_desc, 
		       geco_sol, geco_ocs, ocs_firma, ocs_firma_vive, 
		       acuse_usuario, acuse_fecha, acuse, acuse_firma, 
		       pagado, notas
		FROM servicios
		WHERE id IN (
		    SELECT servicio 
		    FROM servicios_movimientos
		    GROUP BY servicio
		    HAVING COUNT(*) > 1
		    AND SUM(CASE WHEN monto IS NULL THEN 1 ELSE 0 END) = COUNT(*)
		)
		AND geco_ocs IS NOT NULL
		ORDER BY emitido
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ServiciosPendientesDist: error fetching servicios: %w", err)
	}
	defer rows.Close()

	var servicios []Servicio

	for rows.Next() {
		var s Servicio
		var emitido time.Time
		var acuseFecha, pagado sql.NullTime
		var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOCS sql.NullString
		var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
		var montoBruto, montoIVA, montoDesc sql.NullFloat64
		var ocsFirma, ocsFirmaVive sql.NullString
		var justif, notas sql.NullString

		if err := rows.Scan(
			&s.ID, &emitido, &s.Emisor, &s.Detalle, &s.PorEjecutar, &justif, &s.COES,
			&provNom, &provCed, &provDirec, &provEmail, &provTel, &provBanco, &provIBAN, &provJustif,
			&montoBruto, &montoIVA, &montoDesc, &gecoSol, &gecoOCS,
			&ocsFirma, &ocsFirmaVive, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		); err != nil {
			return nil, fmt.Errorf("ServiciosPendientesDist: error scanning row: %w", err)
		}

		s.FirmasCompletas, err = firmasCompletas(db, "servicios_movimientos", "servicio", s.ID)
		if err != nil {
			return nil, err
		}

		if emitido.Year() == periodo && s.FirmasCompletas {
			s.Emitido = emitido
			s.Justif = justif.String
			s.ProvNom = provNom.String
			s.ProvCed = provCed.String
			s.ProvDirec = provDirec.String
			s.ProvEmail = provEmail.String
			s.ProvTel = provTel.String
			s.ProvBanco = provBanco.String
			s.ProvIBAN = provIBAN.String
			s.ProvJustif = provJustif.String
			s.MontoBruto = montoBruto.Float64
			s.MontoIVA = montoIVA.Float64
			s.MontoDesc = montoDesc.Float64
			s.GecoSol = gecoSol.String
			s.GecoOCS = gecoOCS.String
			s.OCSFirma = ocsFirma.String
			s.OCSFirmaVive = ocsFirmaVive.String
			s.AcuseUsuario = acuseUsuario.String
			s.AcuseFecha = acuseFecha.Time
			s.Acuse = acuse.String
			s.AcuseFirma = acuseFirma.String
			s.Pagado = pagado.Time
			s.Notas = notas.String

			servicios = append(servicios, s)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ServiciosPendientesDist: error iterating rows: %w", err)
	}

	return servicios, nil
}

func ServiciosPendientesEjecucion(db *sql.DB, periodo int) ([]Servicio, error) {
	query := `
		SELECT id, emitido, emisor, detalle, por_ejecutar, justif, coes,
		       prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, 
		       prov_iban, prov_justif, monto_bruto, monto_iva, monto_desc, 
		       geco_sol, geco_ocs, ocs_firma, ocs_firma_vive, 
		       acuse_usuario, acuse_fecha, acuse, acuse_firma, 
		       pagado, notas
		FROM servicios
		WHERE geco_ocs IS NOT NULL
		AND acuse IS NULL
		ORDER BY emitido
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("ServiciosPendientesDist: error fetching servicios: %w", err)
	}
	defer rows.Close()

	var servicios []Servicio

	for rows.Next() {
		var s Servicio
		var emitido time.Time
		var acuseFecha, pagado sql.NullTime
		var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOCS sql.NullString
		var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
		var montoBruto, montoIVA, montoDesc sql.NullFloat64
		var ocsFirma, ocsFirmaVive sql.NullString
		var justif, notas sql.NullString

		if err := rows.Scan(
			&s.ID, &emitido, &s.Emisor, &s.Detalle, &s.PorEjecutar, &justif, &s.COES,
			&provNom, &provCed, &provDirec, &provEmail, &provTel, &provBanco, &provIBAN, &provJustif,
			&montoBruto, &montoIVA, &montoDesc, &gecoSol, &gecoOCS,
			&ocsFirma, &ocsFirmaVive, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		); err != nil {
			return nil, fmt.Errorf("ServiciosPendientesDist: error scanning row: %w", err)
		}

		s.FirmasCompletas, err = firmasCompletas(db, "servicios_movimientos", "servicio", s.ID)
		if err != nil {
			return nil, err
		}

		if emitido.Year() == periodo && s.FirmasCompletas {
			s.Emitido = emitido
			s.Justif = justif.String
			s.ProvNom = provNom.String
			s.ProvCed = provCed.String
			s.ProvDirec = provDirec.String
			s.ProvEmail = provEmail.String
			s.ProvTel = provTel.String
			s.ProvBanco = provBanco.String
			s.ProvIBAN = provIBAN.String
			s.ProvJustif = provJustif.String
			s.MontoBruto = montoBruto.Float64
			s.MontoIVA = montoIVA.Float64
			s.MontoDesc = montoDesc.Float64
			s.GecoSol = gecoSol.String
			s.GecoOCS = gecoOCS.String
			s.OCSFirma = ocsFirma.String
			s.OCSFirmaVive = ocsFirmaVive.String
			s.AcuseUsuario = acuseUsuario.String
			s.AcuseFecha = acuseFecha.Time
			s.Acuse = acuse.String
			s.AcuseFirma = acuseFirma.String
			s.Pagado = pagado.Time
			s.Notas = notas.String

			servicios = append(servicios, s)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ServiciosPendientesDist: error iterating rows: %w", err)
	}

	return servicios, nil
}
