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
	// Final
	Ejecutado sql.NullTime
	Pagado    sql.NullTime
	Notas     sql.NullString
	// Runtime
	FirmasCompletas bool
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
	query := `SELECT
	s.id,
	s.emitido,s.emisor,s.detalle,s.por_ejecutar,s.justif,
	s.coes,
	s.prov_nom,s.prov_ced,s.prov_direc,s.prov_email,s.prov_tel,s.prov_banco,s.prov_iban,s.prov_justif,s.monto_bruto,s.monto_iva,s.monto_desc,s.geco_sol,s.geco_ocs,
	s.ocs_firma,s.ocs_firma_vive,
	s.ejecutado,s.pagado,s.notas
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
			&s.Ejecutado, &s.Pagado, &s.Notas,
		); err != nil {
			return nil, fmt.Errorf("serviciosInit: error scanning row: %w", err)
		}

		validez := s.Emitido.Year()
		if validez == periodo {
			queryFirmas := `SELECT firma FROM servicios_movimientos WHERE servicio = ?`

			rows, err := db.Query(queryFirmas, s.ID)
			if err != nil {
				return nil, fmt.Errorf("serviciosInit: error querying database: %w", err)
			}
			defer rows.Close()

			var firmas []string
			s.FirmasCompletas = true

			for rows.Next() {
				var f sql.NullString
				if err := rows.Scan(&f); err != nil {
					return nil, fmt.Errorf("serviciosInit: error scanning row: %w", err)
				}

				if f.Valid {
					firmas = append(firmas, f.String)
				} else {
					s.FirmasCompletas = false
				}
			}

			servicios = append(servicios, s)
		}
	}

	return servicios, nil
}

func NuevoServicio (db *sql.DB, servicio Servicio, movimientos []ServicioMovimiento) error {
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
