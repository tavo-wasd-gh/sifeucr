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
	// Ejecutado
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

func bienesInit(db *sql.DB, c string) ([]Bien, error) {
	var bienes []Bien

	query := `
		SELECT id, emitido, emisor, detalle, por_recibir, justif, coes,
		       prov_nom, prov_ced, prov_direc, prov_email, prov_tel,
		       prov_banco, prov_iban, prov_justif, monto_bruto, monto_iva, monto_desc,
		       geco_sol, geco_oc, oc_firma, oc_firma_vive,
		       acuse_usuario, acuse_fecha, acuse, acuse_firma,
		       pagado, notas
		FROM bienes
		WHERE cuenta = ? 
		ORDER BY emitido
	`

	rows, err := db.Query(query, c)
	if err != nil {
		return nil, fmt.Errorf("bienesInit: error querying bienes: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var b Bien
		var acuseFecha, pagado sql.NullTime
		var acuseUsuario, acuse, acuseFirma, gecoSol, gecoOC sql.NullString
		var montoBruto, montoIVA, montoDesc sql.NullFloat64

		err := rows.Scan(
			&b.ID, &b.Emitido, &b.Emisor, &b.Detalle, &b.PorRecibir, &b.Justif, &b.COES,
			&b.ProvNom, &b.ProvCed, &b.ProvDirec, &b.ProvEmail, &b.ProvTel,
			&b.ProvBanco, &b.ProvIBAN, &b.ProvJustif, &montoBruto, &montoIVA, &montoDesc,
			&gecoSol, &gecoOC, &b.OCFirma, &b.OCFirmaVive,
			&acuseUsuario, &acuseFecha, &acuse, &acuseFirma,
			&pagado, &b.Notas,
		)
		if err != nil {
			return nil, fmt.Errorf("bienesInit: error scanning row: %w", err)
		}

		b.MontoBruto = montoBruto.Float64
		b.MontoIVA = montoIVA.Float64
		b.MontoDesc = montoDesc.Float64
		b.GecoSol = gecoSol.String
		b.GecoOC = gecoOC.String
		b.AcuseUsuario = acuseUsuario.String
		b.AcuseFecha = acuseFecha.Time
		b.Acuse = acuse.String
		b.AcuseFirma = acuseFirma.String
		b.Pagado = pagado.Time

		b.Movimientos, err = bienMovimientosInit(db, b.ID)
		if err != nil {
			return nil, fmt.Errorf("bienesInit: error fetching movimientos for bien %d: %w", b.ID, err)
		}

		b.FirmasCompletas, err = firmasCompletas(db, "bienes_movimientos", "bien", b.ID)

		bienes = append(bienes, b)
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

		err := rows.Scan(
			&m.ID, &m.Bien, &m.Usuario, &m.Cuenta, &m.Presupuesto, &m.Monto, &firma,
		)
		if err != nil {
			return nil, fmt.Errorf("bienMovimientosInit: error scanning row: %w", err)
		}

		m.Firma = firma.String
		movimientos = append(movimientos, m)
	}

	return movimientos, nil
}
