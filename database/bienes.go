package database

import (
	"time"
	"fmt"
	"database/sql"
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
