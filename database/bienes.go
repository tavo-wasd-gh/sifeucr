package database

import (
	"database/sql"
	"fmt"
	"time"
)

type Bien struct {
	ID int
	// Solicitud
	Emitido     time.Time
	Emisor      string
	Detalle     string
	PorRecibir  time.Time
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
	GecoOC     string
	// ViVE
	OCFirma     string
	OCFirmaVive string
	// Recibido
	AcuseUsuario string
	AcuseFecha   time.Time
	Acuse        string
	AcuseFirma   string
	// Final
	Pagado time.Time
	Notas  string
	// Runtime
	Movimientos     []BienMovimiento
	FirmasCompletas bool
	UsuarioLoggeado string
	CuentaLoggeada  string
}

type BienMovimiento struct {
	ID          int
	Bien        int
	Usuario     string
	Cuenta      string
	Presupuesto string
	Monto       float64
	Firma       string
}

func bienesInit(db *sql.DB, cuenta string, periodo int) ([]Bien, error) {
	query := `
		SELECT b.id, b.emitido, b.emisor, b.detalle, b.por_recibir, b.justif, b.coes,
		       b.prov_nom, b.prov_ced, b.prov_direc, b.prov_email, b.prov_tel,
		       b.prov_banco, b.prov_iban, b.prov_justif, b.monto_bruto, b.monto_iva, b.monto_desc,
		       b.geco_sol, b.geco_oc, b.oc_firma, b.oc_firma_vive,
		       b.acuse_usuario, b.acuse_fecha, b.acuse, b.acuse_firma,
		       b.pagado, b.notas
		FROM bienes b
		JOIN bienes_movimientos bm ON b.id = bm.bien
		JOIN presupuestos p ON bm.presupuesto = p.id
		JOIN cuentas c ON p.cuenta = c.id
		WHERE c.id = ?
		ORDER BY b.emitido DESC
	`

	rows, err := db.Query(query, cuenta)
	if err != nil {
		return nil, fmt.Errorf("bienesInit: error querying bienes: %w", err)
	}
	defer rows.Close()

	var bienes []Bien

	for rows.Next() {
		var b Bien
		var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
		var acuseFecha, pagado sql.NullTime
		var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOC sql.NullString
		var ocFirma, ocFirmaVive sql.NullString
		var montoBruto, montoIVA, montoDesc sql.NullFloat64
		var justif, notas sql.NullString

		err := rows.Scan(
			&b.ID, &b.Emitido, &b.Emisor, &b.Detalle, &b.PorRecibir, &justif, &b.COES,
			&provNom, &provCed, &provDirec, &provEmail, &provTel,
			&provBanco, &provIBAN, &provJustif, &montoBruto, &montoIVA, &montoDesc,
			&gecoSol, &gecoOC, &ocFirma, &ocFirmaVive,
			&acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		)
		if err != nil {
			return nil, fmt.Errorf("bienesInit: error scanning row: %w", err)
		}

		b.MontoBruto = montoBruto.Float64
		b.MontoIVA = montoIVA.Float64
		b.MontoDesc = montoDesc.Float64
		b.GecoSol = gecoSol.String
		b.ProvNom = provNom.String
		b.ProvCed = provCed.String
		b.ProvDirec = provDirec.String
		b.ProvEmail = provEmail.String
		b.ProvTel = provTel.String
		b.ProvBanco = provBanco.String
		b.ProvIBAN = provIBAN.String
		b.ProvJustif = provJustif.String
		b.GecoOC = gecoOC.String
		b.OCFirma = ocFirma.String
		b.OCFirmaVive = ocFirmaVive.String
		b.AcuseUsuario = acuseUsuario.String
		b.AcuseFecha = acuseFecha.Time
		b.Acuse = acuse.String
		b.AcuseFirma = acuseFirma.String
		b.Pagado = pagado.Time
		b.Justif = justif.String
		b.Notas = notas.String

		if b.Emitido.Year() == periodo {
			b.Movimientos, err = bienMovimientosInit(db, b.ID)
			if err != nil {
				return nil, fmt.Errorf("bienesInit: error fetching movimientos for bien %d: %w", b.ID, err)
			}

			b.FirmasCompletas, err = firmasCompletas(db, "bienes_movimientos", "bien", b.ID)
			if err != nil {
				return nil, err
			}

			bienes = append(bienes, b)
		}
	}

	return bienes, nil
}

func bienMovimientosInit(db *sql.DB, bienID int) ([]BienMovimiento, error) {
	var movimientos []BienMovimiento

	query := `
		SELECT id, bien, usuario, cuenta, presupuesto, monto, firma
		FROM bienes_movimientos
		WHERE bien = ?
	`

	rows, err := db.Query(query, bienID)
	if err != nil {
		return nil, fmt.Errorf("bienMovimientosInit: error querying movimientos: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var m BienMovimiento
		var firma sql.NullString
		var usuario sql.NullString
		var monto sql.NullFloat64

		err := rows.Scan(
			&m.ID, &m.Bien, &usuario, &m.Cuenta, &m.Presupuesto, &monto, &firma,
		)
		if err != nil {
			return nil, fmt.Errorf("bienMovimientosInit: error scanning row: %w", err)
		}

		m.Usuario = usuario.String
		m.Monto = monto.Float64
		m.Firma = firma.String

		movimientos = append(movimientos, m)
	}

	return movimientos, nil
}

func NuevoBien(db *sql.DB, bien Bien) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("NuevoBien: failed to begin transaction: %w", err)
	}

	var bienID int
	err = tx.QueryRow(`
		INSERT INTO bienes (emitido, emisor, detalle, por_recibir, justif, coes) 
		VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`,
		time.Now(), bien.Emisor, bien.Detalle, bien.PorRecibir, bien.Justif, false,
		).Scan(&bienID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("NuevoBien: failed to insert bien: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO bienes_movimientos (bien, usuario, cuenta, presupuesto, firma) 
		VALUES ($1, $2, $3, $4, $5)`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("NuevoBien: failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for i, mov := range bien.Movimientos {
		presupuestoID, err := presupuestoActual(db, mov.Cuenta)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("NuevoBien: failed to fetch presupuesto for movimiento %d: %w", i+1, err)
		}

		_, err = stmt.Exec(bienID, mov.Usuario, mov.Cuenta, presupuestoID, mov.Firma)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("NuevoBien: failed to insert bienes_movimientos: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("NuevoBien: failed to commit transaction: %w", err)
	}

	return nil
}

func LeerBien(db *sql.DB, id, cuenta string) (Bien, error) {
	var b Bien
	var acuseFecha, pagado sql.NullTime
	var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOC sql.NullString
	var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
	var montoBruto, montoIVA, montoDesc sql.NullFloat64
	var ocFirma, ocFirmaVive sql.NullString
	var justif, notas sql.NullString

	err := db.QueryRow(`
		SELECT id, emitido, emisor, detalle, por_recibir, justif, coes,
		prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, prov_iban, prov_justif,
		monto_bruto, monto_iva, monto_desc, geco_sol, geco_oc, 
		oc_firma, oc_firma_vive, acuse_usuario, acuse_fecha, acuse, acuse_firma,
		pagado, notas
		FROM bienes WHERE id = ?`, id).
		Scan(
			&b.ID, &b.Emitido, &b.Emisor, &b.Detalle, &b.PorRecibir, &justif, &b.COES,
			&provNom, &provCed, &provDirec, &provEmail, &provTel, &provBanco, &provIBAN, &provJustif,
			&montoBruto, &montoIVA, &montoDesc, &gecoSol, &gecoOC,
			&ocFirma, &ocFirmaVive, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		)

	if err != nil {
		if err == sql.ErrNoRows {
			return Bien{}, fmt.Errorf("LeerBien: bien con ID '%s' no encontrado", id)
		}
		return Bien{}, fmt.Errorf("LeerBien: error al obtener bien: %w", err)
	}

	b.Pagado = pagado.Time
	b.AcuseFecha = acuseFecha.Time
	b.AcuseUsuario = acuseUsuario.String
	b.Acuse = acuse.String
	b.AcuseFirma = acuseFirma.String
	b.GecoSol = gecoSol.String
	b.GecoOC = gecoOC.String
	b.OCFirma = ocFirma.String
	b.OCFirmaVive = ocFirmaVive.String
	b.ProvNom = provNom.String
	b.ProvCed = provCed.String
	b.ProvDirec = provDirec.String
	b.ProvEmail = provEmail.String
	b.ProvTel = provTel.String
	b.ProvBanco = provBanco.String
	b.ProvIBAN = provIBAN.String
	b.ProvJustif = provJustif.String
	b.MontoBruto = montoBruto.Float64
	b.MontoIVA = montoIVA.Float64
	b.MontoDesc = montoDesc.Float64
	b.Justif = justif.String
	b.Notas = notas.String

	rows, err := db.Query(`
		SELECT id, bien, usuario, cuenta, presupuesto, monto, firma 
		FROM bienes_movimientos 
		WHERE bien = ?`, id)
	if err != nil {
		return Bien{}, fmt.Errorf("LeerBien: error al obtener movimientos: %w", err)
	}
	defer rows.Close()

	var movimientos []BienMovimiento
	found := false
	firmasCompletas := true

	for rows.Next() {
		var m BienMovimiento
		var firma sql.NullString
		var usuario sql.NullString
		var monto sql.NullFloat64

		if err := rows.Scan(&m.ID, &m.Bien, &usuario, &m.Cuenta, &m.Presupuesto, &monto, &firma); err != nil {
			return Bien{}, fmt.Errorf("LeerBien: error al escanear movimientos: %w", err)
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

	b.FirmasCompletas = firmasCompletas

	if err := rows.Err(); err != nil {
		return Bien{}, fmt.Errorf("LeerBien: error al recorrer movimientos: %w", err)
	}

	b.Movimientos = movimientos

	if !found && cuenta != "COES" && cuenta != "SF" {
		return Bien{}, fmt.Errorf("LeerBien: cuenta '%s' no encontrada en participantes", cuenta)
	}

	return b, nil
}

func FirmarMovimientoBienes(db *sql.DB, id, usuario, cuenta, firma string) error {
	_, err := UsuarioAcreditado(db, usuario, cuenta)
	if err != nil {
		return fmt.Errorf("FirmarMovimientoBienes: error al iniciar usuario: %w", err)
	}

	var existingCuenta string
	err = db.QueryRow("SELECT cuenta FROM bienes_movimientos WHERE id = ?", id).Scan(&existingCuenta)
	if err != nil {
		return fmt.Errorf("FirmarMovimientoBienes: error retrieving cuenta for id %s: %w", id, err)
	}
	if existingCuenta != cuenta {
		return fmt.Errorf("FirmarMovimientoBienes: cuenta mismatch for id %s (expected: %s, got: %s)", id, existingCuenta, cuenta)
	}

	query := `UPDATE bienes_movimientos
	SET usuario = ?, firma = ?
	WHERE id = ?;`

	if _, err = db.Exec(query, usuario, firma, id) ; err != nil {
		return fmt.Errorf("FirmarMovimientoBienes: failed to update bien_movimiento with id %s: %w", id, err)
	}

	return nil
}

func ConfirmarRecibidoBienes(db *sql.DB, id, usuario, cuenta string, fecha time.Time, acuse, firma string) error {
	now := time.Now()
	oneMonthAgo := now.AddDate(0, -1, 0)

	if fecha.After(now) || fecha.Before(oneMonthAgo) {
		return fmt.Errorf("ConfirmarRecibidoBienes: invalid date")
	}

	_, err := UsuarioAcreditado(db, usuario, cuenta)
	if err != nil {
		return fmt.Errorf("ConfirmarRecibidoBienes: usuario %s no acreditado para cuenta %s: %w", usuario, cuenta, err)
	}

	query := `UPDATE bienes
		SET acuse_usuario = ?, acuse_fecha = ?, acuse = ?, acuse_firma = ?
		WHERE id = ?;`

	_, err = db.Exec(query, usuario, fecha, acuse, firma, id)
	if err != nil {
		return fmt.Errorf("ConfirmarRecibidoBienes: failed to update bien with id %s: %w", id, err)
	}

	return nil
}

func BienesPendientesCOES(db *sql.DB, periodo int) ([]Bien, error) {
	query := `
		SELECT id, emitido, emisor, detalle, por_recibir, justif, 
		       coes, prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, 
		       prov_iban, prov_justif, monto_bruto, monto_iva, monto_desc, 
		       geco_sol, geco_oc, oc_firma, oc_firma_vive, 
		       acuse_usuario, acuse_fecha, acuse, acuse_firma, 
		       pagado, notas
		FROM bienes
		WHERE coes = FALSE
		ORDER BY emitido
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("BienesPendientesCOES: error fetching bienes: %w", err)
	}
	defer rows.Close()

	var bienes []Bien

	for rows.Next() {
		var b Bien
		var emitido time.Time
		var acuseFecha, pagado sql.NullTime
		var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOC sql.NullString
		var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
		var montoBruto, montoIVA, montoDesc sql.NullFloat64
		var ocFirma, ocFirmaVive sql.NullString
		var justif, notas sql.NullString

		if err := rows.Scan(
			&b.ID, &emitido, &b.Emisor, &b.Detalle, &b.PorRecibir, &justif,
			&b.COES, &provNom, &provCed, &provDirec, &provEmail, &provTel, &provBanco, &provIBAN, &provJustif,
			&montoBruto, &montoIVA, &montoDesc, &gecoSol, &gecoOC,
			&ocFirma, &ocFirmaVive, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		); err != nil {
			return nil, fmt.Errorf("BienesPendientesCOES: error scanning row: %w", err)
		}

		b.FirmasCompletas, err = firmasCompletas(db, "bienes_movimientos", "bien", b.ID)
		if err != nil {
			return nil, err
		}

		if emitido.Year() == periodo && b.FirmasCompletas {
			b.Emitido = emitido
			b.Justif = justif.String
			b.ProvNom = provNom.String
			b.ProvCed = provCed.String
			b.ProvDirec = provDirec.String
			b.ProvEmail = provEmail.String
			b.ProvTel = provTel.String
			b.ProvBanco = provBanco.String
			b.ProvIBAN = provIBAN.String
			b.ProvJustif = provJustif.String
			b.MontoBruto = montoBruto.Float64
			b.MontoIVA = montoIVA.Float64
			b.MontoDesc = montoDesc.Float64
			b.GecoSol = gecoSol.String
			b.GecoOC = gecoOC.String
			b.OCFirma = ocFirma.String
			b.OCFirmaVive = ocFirmaVive.String
			b.AcuseUsuario = acuseUsuario.String
			b.AcuseFecha = acuseFecha.Time
			b.Acuse = acuse.String
			b.AcuseFirma = acuseFirma.String
			b.Pagado = pagado.Time
			b.Notas = notas.String

			bienes = append(bienes, b)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("BienesPendientesCOES: error iterating rows: %w", err)
	}

	return bienes, nil
}

func AprobarBienCOES(db *sql.DB, id string) error {
	_, err := db.Exec(`UPDATE bienes SET coes = TRUE WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("AprobarBienCOES: failed to update service: %w", err)
	}
	return nil
}

func BienPorID(db *sql.DB, usuarioLoggeado, cuentaLoggeada, id string) (Bien, error) {
	var b Bien

	b.UsuarioLoggeado = usuarioLoggeado
	b.CuentaLoggeada = cuentaLoggeada

	var acuseFecha, pagado sql.NullTime
	var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOC sql.NullString
	var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
	var montoBruto, montoIVA, montoDesc sql.NullFloat64
	var ocFirma, ocFirmaVive sql.NullString
	var justif, notas sql.NullString

	err := db.QueryRow(`
		SELECT id, emitido, emisor, detalle, por_recibir, justif, coes,
		prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, prov_iban, prov_justif,
		monto_bruto, monto_iva, monto_desc, geco_sol, geco_oc, 
		oc_firma, oc_firma_vive, acuse_usuario, acuse_fecha, acuse, acuse_firma,
		pagado, notas
		FROM bienes WHERE id = ?`, id).
		Scan(
			&b.ID, &b.Emitido, &b.Emisor, &b.Detalle, &b.PorRecibir, &justif, &b.COES,
			&provNom, &provCed, &provDirec, &provEmail, &provTel, &provBanco, &provIBAN, &provJustif,
			&montoBruto, &montoIVA, &montoDesc, &gecoSol, &gecoOC,
			&ocFirma, &ocFirmaVive, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		)

	if err != nil {
		if err == sql.ErrNoRows {
			return Bien{}, fmt.Errorf("LeerBien: bien con ID '%s' no encontrado", id)
		}
		return Bien{}, fmt.Errorf("LeerBien: error al obtener bien: %w", err)
	}

	b.Pagado = pagado.Time
	b.AcuseFecha = acuseFecha.Time
	b.AcuseUsuario = acuseUsuario.String
	b.Acuse = acuse.String
	b.AcuseFirma = acuseFirma.String
	b.GecoSol = gecoSol.String
	b.GecoOC = gecoOC.String
	b.OCFirma = ocFirma.String
	b.OCFirmaVive = ocFirmaVive.String
	b.ProvNom = provNom.String
	b.ProvCed = provCed.String
	b.ProvDirec = provDirec.String
	b.ProvEmail = provEmail.String
	b.ProvTel = provTel.String
	b.ProvBanco = provBanco.String
	b.ProvIBAN = provIBAN.String
	b.ProvJustif = provJustif.String
	b.MontoBruto = montoBruto.Float64
	b.MontoIVA = montoIVA.Float64
	b.MontoDesc = montoDesc.Float64
	b.Justif = justif.String
	b.Notas = notas.String

	rows, err := db.Query(`
		SELECT id, bien, usuario, cuenta, presupuesto, monto, firma 
		FROM bienes_movimientos 
		WHERE bien = ?`, id)
	if err != nil {
		return Bien{}, fmt.Errorf("LeerBien: error al obtener movimientos: %w", err)
	}
	defer rows.Close()

	var movimientos []BienMovimiento
	found := false
	firmasCompletas := true

	for rows.Next() {
		var m BienMovimiento
		var firma sql.NullString
		var usuario sql.NullString
		var monto sql.NullFloat64

		if err := rows.Scan(&m.ID, &m.Bien, &usuario, &m.Cuenta, &m.Presupuesto, &monto, &firma); err != nil {
			return Bien{}, fmt.Errorf("LeerBien: error al escanear movimientos: %w", err)
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

	b.FirmasCompletas = firmasCompletas

	if err := rows.Err(); err != nil {
		return Bien{}, fmt.Errorf("LeerBien: error al recorrer movimientos: %w", err)
	}

	b.Movimientos = movimientos

	if !found && cuentaLoggeada != "COES" && cuentaLoggeada != "SF" {
		return Bien{}, fmt.Errorf("LeerBien: cuenta '%s' no encontrada en participantes", cuentaLoggeada)
	}

	return b, nil
}

func (b *Bien) EstablecerMontos(db *sql.DB, montos map[string]float64) error {
	if b.MontoBruto <= 0 {
		return fmt.Errorf("EstablecerMontos: monto bruto is not yet set")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("EstablecerMontos: failed to begin transaction: %w", err)
	}

	var totalSum float64
	for _, mov := range b.Movimientos {
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

		_, err := tx.Exec(`UPDATE bienes_movimientos SET monto = ? WHERE id = ?`, monto, mov.ID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("EstablecerMontos: failed to update movimiento ID %d: %w", mov.ID, err)
		}
	}

	if totalSum != b.MontoBruto {
		tx.Rollback()
		return fmt.Errorf("EstablecerMontos: total montos (%.2f) do not match MontoBruto (%.2f)", totalSum, b.MontoBruto)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("EstablecerMontos: failed to commit transaction: %w", err)
	}

	return nil
}

// Registrar el número de solicitud de GECO en la base de datos
// bien.RegistrarGECO(db, solicitud)
func (b *Bien) RegistrarSolicitudGECO(db *sql.DB, sol string) error {
	if b.CuentaLoggeada != "SF" {
		return fmt.Errorf("RegistrarSolicitudGECO: failed to update bien: unauthorized account")
	}

	_, err := db.Exec(`UPDATE bienes SET geco_sol = ? WHERE id = ?`, sol, b.ID)
	if err != nil {
		return fmt.Errorf("RegistrarSolicitudGECO: failed to update bien: %w", err)
	}

	return nil
}

func (b *Bien) RegistrarOC(
	db *sql.DB, gecoOCS, provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif string,
	montoBruto, montoIVA, montoDesc float64,
) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("RegistrarOC: failed to begin transaction: %w", err)
	}

	_, err = tx.Exec(`
		UPDATE bienes 
		SET geco_oc = ?, prov_nom = ?, prov_ced = ?, prov_direc = ?, prov_email = ?, prov_tel = ?, 
		    prov_banco = ?, prov_iban = ?, prov_justif = ?, monto_bruto = ?, monto_iva = ?, monto_desc = ?
		WHERE id = ?
	`, gecoOCS, provNom, provCed, provDirec, provEmail, provTel,
		provBanco, provIBAN, provJustif, montoBruto, montoIVA, montoDesc, b.ID)

	if err != nil {
		tx.Rollback()
		return fmt.Errorf("RegistrarOC: failed to update service ID %d: %w", b.ID, err)
	}

	if len(b.Movimientos) == 1 {
		_, err = tx.Exec(`
			UPDATE bienes_movimientos 
			SET monto = ? 
			WHERE bien = ?
			`, montoBruto, b.ID)

		if err != nil {
			tx.Rollback()
			return fmt.Errorf("RegistrarOC: failed to update movimiento for bienes ID %d: %w", b.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("RegistrarOC: failed to commit transaction: %w", err)
	}

	return nil
}

func BienesPendientesGECO(db *sql.DB, periodo int) ([]Bien, error) {
	query := `
		SELECT id, emitido, emisor, detalle, por_recibir, justif, 
		       coes, prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, 
		       prov_iban, prov_justif, monto_bruto, monto_iva, monto_desc, 
		       geco_sol, geco_oc, oc_firma, oc_firma_vive, 
		       acuse_usuario, acuse_fecha, acuse, acuse_firma, 
		       pagado, notas
		FROM bienes
		WHERE geco_sol IS NULL OR geco_sol = ''
		ORDER BY emitido
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("BienesPendientesGECO: error fetching bienes: %w", err)
	}
	defer rows.Close()

	var bienes []Bien

	for rows.Next() {
		var b Bien
		var emitido time.Time
		var acuseFecha, pagado sql.NullTime
		var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOC sql.NullString
		var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
		var montoBruto, montoIVA, montoDesc sql.NullFloat64
		var ocFirma, ocFirmaVive sql.NullString
		var justif, notas sql.NullString

		if err := rows.Scan(
			&b.ID, &emitido, &b.Emisor, &b.Detalle, &b.PorRecibir, &justif,
			&b.COES, &provNom, &provCed, &provDirec, &provEmail, &provTel, &provBanco, &provIBAN, &provJustif,
			&montoBruto, &montoIVA, &montoDesc, &gecoSol, &gecoOC,
			&ocFirma, &ocFirmaVive, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		); err != nil {
			return nil, fmt.Errorf("BienesPendientesGECO: error scanning row: %w", err)
		}

		b.FirmasCompletas, err = firmasCompletas(db, "bienes_movimientos", "bien", b.ID)
		if err != nil {
			return nil, err
		}

		if emitido.Year() == periodo && b.FirmasCompletas {
			b.Emitido = emitido
			b.Justif = justif.String
			b.ProvNom = provNom.String
			b.ProvCed = provCed.String
			b.ProvDirec = provDirec.String
			b.ProvEmail = provEmail.String
			b.ProvTel = provTel.String
			b.ProvBanco = provBanco.String
			b.ProvIBAN = provIBAN.String
			b.ProvJustif = provJustif.String
			b.MontoBruto = montoBruto.Float64
			b.MontoIVA = montoIVA.Float64
			b.MontoDesc = montoDesc.Float64
			b.GecoSol = gecoSol.String
			b.GecoOC = gecoOC.String
			b.OCFirma = ocFirma.String
			b.OCFirmaVive = ocFirmaVive.String
			b.AcuseUsuario = acuseUsuario.String
			b.AcuseFecha = acuseFecha.Time
			b.Acuse = acuse.String
			b.AcuseFirma = acuseFirma.String
			b.Pagado = pagado.Time
			b.Notas = notas.String

			bienes = append(bienes, b)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("BienesPendientesGECO: error iterating rows: %w", err)
	}

	return bienes, nil
}

func BienesPendientesOC(db *sql.DB, periodo int) ([]Bien, error) {
	query := `
		SELECT id, emitido, emisor, detalle, por_recibir, justif, 
		       coes, prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, 
		       prov_iban, prov_justif, monto_bruto, monto_iva, monto_desc, 
		       geco_sol, geco_oc, oc_firma, oc_firma_vive, 
		       acuse_usuario, acuse_fecha, acuse, acuse_firma, 
		       pagado, notas
		FROM bienes
		WHERE geco_oc IS NULL
		AND geco_sol IS NOT NULL
		AND coes = TRUE
		ORDER BY emitido
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("BienesPendientesOCS: error fetching bienes: %w", err)
	}
	defer rows.Close()

	var bienes []Bien

	for rows.Next() {
		var b Bien
		var emitido time.Time
		var acuseFecha, pagado sql.NullTime
		var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOC sql.NullString
		var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
		var montoBruto, montoIVA, montoDesc sql.NullFloat64
		var ocFirma, ocFirmaVive sql.NullString
		var justif, notas sql.NullString

		if err := rows.Scan(
			&b.ID, &emitido, &b.Emisor, &b.Detalle, &b.PorRecibir, &justif,
			&b.COES, &provNom, &provCed, &provDirec, &provEmail, &provTel, &provBanco, &provIBAN, &provJustif,
			&montoBruto, &montoIVA, &montoDesc, &gecoSol, &gecoOC,
			&ocFirma, &ocFirmaVive, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		); err != nil {
			return nil, fmt.Errorf("BienesPendientesOCS: error scanning row: %w", err)
		}

		b.FirmasCompletas, err = firmasCompletas(db, "bienes_movimientos", "bien", b.ID)
		if err != nil {
			return nil, err
		}

		if emitido.Year() == periodo && b.FirmasCompletas {
			b.Emitido = emitido
			b.Justif = justif.String
			b.ProvNom = provNom.String
			b.ProvCed = provCed.String
			b.ProvDirec = provDirec.String
			b.ProvEmail = provEmail.String
			b.ProvTel = provTel.String
			b.ProvBanco = provBanco.String
			b.ProvIBAN = provIBAN.String
			b.ProvJustif = provJustif.String
			b.MontoBruto = montoBruto.Float64
			b.MontoIVA = montoIVA.Float64
			b.MontoDesc = montoDesc.Float64
			b.GecoSol = gecoSol.String
			b.GecoOC = gecoOC.String
			b.OCFirma = ocFirma.String
			b.OCFirmaVive = ocFirmaVive.String
			b.AcuseUsuario = acuseUsuario.String
			b.AcuseFecha = acuseFecha.Time
			b.Acuse = acuse.String
			b.AcuseFirma = acuseFirma.String
			b.Pagado = pagado.Time
			b.Notas = notas.String

			bienes = append(bienes, b)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("BienesPendientesOCS: error iterating rows: %w", err)
	}

	return bienes, nil
}

func BienesPendientesDist(db *sql.DB, periodo int) ([]Bien, error) {
	query := `
		SELECT id, emitido, emisor, detalle, por_recibir, justif, coes,
		       prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, 
		       prov_iban, prov_justif, monto_bruto, monto_iva, monto_desc, 
		       geco_sol, geco_oc, oc_firma, oc_firma_vive, 
		       acuse_usuario, acuse_fecha, acuse, acuse_firma, 
		       pagado, notas
		FROM bienes
		WHERE id IN (
		    SELECT bien 
		    FROM bienes_movimientos
		    GROUP BY bien
		    HAVING COUNT(*) > 1
		    AND SUM(CASE WHEN monto IS NULL THEN 1 ELSE 0 END) = COUNT(*)
		)
		AND geco_oc IS NOT NULL
		ORDER BY emitido
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("BienesPendientesDist: error fetching servicios: %w", err)
	}
	defer rows.Close()

	var bienes []Bien

	for rows.Next() {
		var b Bien
		var emitido time.Time
		var acuseFecha, pagado sql.NullTime
		var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOC sql.NullString
		var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
		var montoBruto, montoIVA, montoDesc sql.NullFloat64
		var ocsFirma, ocsFirmaVive sql.NullString
		var justif, notas sql.NullString

		if err := rows.Scan(
			&b.ID, &emitido, &b.Emisor, &b.Detalle, &b.PorRecibir, &justif, &b.COES,
			&provNom, &provCed, &provDirec, &provEmail, &provTel, &provBanco, &provIBAN, &provJustif,
			&montoBruto, &montoIVA, &montoDesc, &gecoSol, &gecoOC,
			&ocsFirma, &ocsFirmaVive, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		); err != nil {
			return nil, fmt.Errorf("BienesPendientesDist: error scanning row: %w", err)
		}

		b.FirmasCompletas, err = firmasCompletas(db, "bienes_movimientos", "bien", b.ID)
		if err != nil {
			return nil, err
		}

		if emitido.Year() == periodo && b.FirmasCompletas {
			b.Emitido = emitido
			b.Justif = justif.String
			b.ProvNom = provNom.String
			b.ProvCed = provCed.String
			b.ProvDirec = provDirec.String
			b.ProvEmail = provEmail.String
			b.ProvTel = provTel.String
			b.ProvBanco = provBanco.String
			b.ProvIBAN = provIBAN.String
			b.ProvJustif = provJustif.String
			b.MontoBruto = montoBruto.Float64
			b.MontoIVA = montoIVA.Float64
			b.MontoDesc = montoDesc.Float64
			b.GecoSol = gecoSol.String
			b.GecoOC = gecoOC.String
			b.OCFirma = ocsFirma.String
			b.OCFirmaVive = ocsFirmaVive.String
			b.AcuseUsuario = acuseUsuario.String
			b.AcuseFecha = acuseFecha.Time
			b.Acuse = acuse.String
			b.AcuseFirma = acuseFirma.String
			b.Pagado = pagado.Time
			b.Notas = notas.String

			bienes = append(bienes, b)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("BienesPendientesDist: error iterating rows: %w", err)
	}

	return bienes, nil
}

func BienesPendientesRecepcion(db *sql.DB, periodo int) ([]Bien, error) {
	query := `
		SELECT id, emitido, emisor, detalle, por_recibir, justif, coes,
		       prov_nom, prov_ced, prov_direc, prov_email, prov_tel, prov_banco, 
		       prov_iban, prov_justif, monto_bruto, monto_iva, monto_desc, 
		       geco_sol, geco_oc, oc_firma, oc_firma_vive, 
		       acuse_usuario, acuse_fecha, acuse, acuse_firma, 
		       pagado, notas
		FROM bienes
		WHERE geco_oc IS NOT NULL
		AND acuse IS NULL
		ORDER BY emitido
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("BienesPendientesRecepcion: error fetching servicios: %w", err)
	}
	defer rows.Close()

	var bienes []Bien

	for rows.Next() {
		var b Bien
		var emitido time.Time
		var acuseFecha, pagado sql.NullTime
		var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOC sql.NullString
		var provNom, provCed, provDirec, provEmail, provTel, provBanco, provIBAN, provJustif sql.NullString
		var montoBruto, montoIVA, montoDesc sql.NullFloat64
		var ocsFirma, ocsFirmaVive sql.NullString
		var justif, notas sql.NullString

		if err := rows.Scan(
			&b.ID, &emitido, &b.Emisor, &b.Detalle, &b.PorRecibir, &justif, &b.COES,
			&provNom, &provCed, &provDirec, &provEmail, &provTel, &provBanco, &provIBAN, &provJustif,
			&montoBruto, &montoIVA, &montoDesc, &gecoSol, &gecoOC,
			&ocsFirma, &ocsFirmaVive, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &notas,
		); err != nil {
			return nil, fmt.Errorf("BienesPendientesRecepcion: error scanning row: %w", err)
		}

		b.FirmasCompletas, err = firmasCompletas(db, "bienes_movimientos", "bien", b.ID)
		if err != nil {
			return nil, err
		}

		if emitido.Year() == periodo && b.FirmasCompletas {
			b.Emitido = emitido
			b.Justif = justif.String
			b.ProvNom = provNom.String
			b.ProvCed = provCed.String
			b.ProvDirec = provDirec.String
			b.ProvEmail = provEmail.String
			b.ProvTel = provTel.String
			b.ProvBanco = provBanco.String
			b.ProvIBAN = provIBAN.String
			b.ProvJustif = provJustif.String
			b.MontoBruto = montoBruto.Float64
			b.MontoIVA = montoIVA.Float64
			b.MontoDesc = montoDesc.Float64
			b.GecoSol = gecoSol.String
			b.GecoOC = gecoOC.String
			b.OCFirma = ocsFirma.String
			b.OCFirmaVive = ocsFirmaVive.String
			b.AcuseUsuario = acuseUsuario.String
			b.AcuseFecha = acuseFecha.Time
			b.Acuse = acuse.String
			b.AcuseFirma = acuseFirma.String
			b.Pagado = pagado.Time
			b.Notas = notas.String

			bienes = append(bienes, b)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("BienesPendientesRecepcion: error iterating rows: %w", err)
	}

	return bienes, nil
}
