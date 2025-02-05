package database

import (
	"time"
	"fmt"
	"database/sql"
)

type Suministros struct {
	ID int
	// Solicitud
	Emitido     time.Time
	Emisor      string
	Cuenta      string
	Presupuesto string
	Justif      string
	Firma       string
	// COES
	COES bool
	// OSUM
	MontoBrutoTotal float64
	GECO            string
	// Recibido
	AcuseUsuario string
	AcuseFecha   time.Time
	Acuse        string
	AcuseFirma   string
	// Final
	Notas string
	// Runtime
	Desglose []SuministroDesglose
	UsuarioLoggeado string
	CuentaLoggeada string
}

type SuministroDesglose struct {
	ID          int
	Suministros int
	// Artículo
	Nombre        string
	Articulo      string
	Agrupacion    string
	Cantidad      int
	MontoUnitario float64
}

func suministrosInit(db *sql.DB, c string, periodo int) ([]Suministros, error) {
	var suministros []Suministros

	query := `
	SELECT id, emitido, emisor, presupuesto, justif, firma, coes, 
	monto_bruto_total, geco, acuse_usuario, acuse_fecha, 
	acuse, acuse_firma, notas
	FROM suministros
	WHERE cuenta = ?
	ORDER BY emitido DESC
	`

	rows, err := db.Query(query, c)
	if err != nil {
		return nil, fmt.Errorf("suministrosInit: error querying suministros: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var s Suministros
		var firma sql.NullString
		var acuseFecha sql.NullTime
		var acuseUsuario, acuse, acuseFirma sql.NullString
		var geco sql.NullString
		var montoBrutoTotal sql.NullFloat64
		var notas sql.NullString

		err := rows.Scan(
			&s.ID, &s.Emitido, &s.Emisor, &s.Presupuesto, &s.Justif, &firma, &s.COES,
			&montoBrutoTotal, &geco, &acuseUsuario, &acuseFecha,
			&acuse, &acuseFirma, &notas,
		)
		if err != nil {
			return nil, fmt.Errorf("suministrosInit: error scanning row: %w", err)
		}

		s.Firma = firma.String
		s.MontoBrutoTotal = montoBrutoTotal.Float64
		s.GECO = geco.String
		s.AcuseUsuario = acuseUsuario.String
		s.AcuseFecha = acuseFecha.Time
		s.Acuse = acuse.String
		s.AcuseFirma = acuseFirma.String
		s.Notas = notas.String

		if s.Emitido.Year() == periodo {
			s.Desglose, err = suministroDesgloseInit(db, s.ID)
			if err != nil {
				return nil, fmt.Errorf("suministrosInit: error fetching desglose for suministro %d: %w", s.ID, err)
			}

			suministros = append(suministros, s)
		}
	}

	return suministros, nil
}

func suministroDesgloseInit(db *sql.DB, suministroID int) ([]SuministroDesglose, error) {
	var desgloseList []SuministroDesglose

	query := `
	SELECT id, suministros, nombre, articulo, agrupacion, cantidad, monto_unitario
	FROM suministros_desglose
	WHERE suministros = ?
	`

	rows, err := db.Query(query, suministroID)
	if err != nil {
		return nil, fmt.Errorf("suministroDesgloseInit: error querying desglose: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var d SuministroDesglose
		err := rows.Scan(
			&d.ID, &d.Suministros, &d.Nombre, &d.Articulo, &d.Agrupacion, &d.Cantidad, &d.MontoUnitario,
		)
		if err != nil {
			return nil, fmt.Errorf("fetchSuministroDesglose: error scanning row: %w", err)
		}
		desgloseList = append(desgloseList, d)
	}

	return desgloseList, nil
}

func NuevoSuministro(db *sql.DB, suministro Suministros) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	presupuesto, err := presupuestoActual(db, suministro.Cuenta)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO suministros (emitido, emisor, cuenta, presupuesto, justif, firma, coes)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`
	var suministroID int
	err = tx.QueryRow(
		query,
		suministro.Emitido,
		suministro.Emisor,
		suministro.Cuenta,
		presupuesto,
		suministro.Justif,
		suministro.Firma,
		suministro.COES,
	).Scan(&suministroID)

	if err != nil {
		return err
	}

	desgloseQuery := `
		INSERT INTO suministros_desglose (suministros, nombre, articulo, agrupacion, cantidad, monto_unitario)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	for _, item := range suministro.Desglose {
		_, err = tx.Exec(
			desgloseQuery,
			suministroID,
			item.Nombre,
			item.Articulo,
			item.Agrupacion,
			item.Cantidad,
			item.MontoUnitario,
		)
		if err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func LeerSuministro(db *sql.DB, id, cuenta string) (Suministros, error) {
	var s Suministros
	var acuseFecha sql.NullTime
	var acuseUsuario, acuse, acuseFirma, geco, notas, firma sql.NullString
	var montoBrutoTotal sql.NullFloat64

	err := db.QueryRow(`
		SELECT id, emitido, emisor, cuenta, presupuesto, justif, firma, coes, 
		monto_bruto_total, geco, acuse_usuario, acuse_fecha, acuse, acuse_firma, notas
		FROM suministros WHERE id = ?`, id).
		Scan(
			&s.ID, &s.Emitido, &s.Emisor, &s.Cuenta, &s.Presupuesto, &s.Justif, &firma, &s.COES,
			&montoBrutoTotal, &geco, &acuseUsuario, &acuseFecha, &acuse, &acuseFirma, &notas,
		)

	if err != nil {
		if err == sql.ErrNoRows {
			return Suministros{}, fmt.Errorf("LeerSuministros: suministro con ID '%s' no encontrado", id)
		}
		return Suministros{}, fmt.Errorf("LeerSuministros: error al obtener suministro: %w", err)
	}

	s.Firma = firma.String
	s.AcuseUsuario = acuseUsuario.String
	s.AcuseFecha = acuseFecha.Time
	s.Acuse = acuse.String
	s.AcuseFirma = acuseFirma.String
	s.GECO = geco.String
	s.MontoBrutoTotal = montoBrutoTotal.Float64
	s.Notas = notas.String

	rows, err := db.Query(`
		SELECT id, suministros, nombre, articulo, agrupacion, cantidad, monto_unitario 
		FROM suministros_desglose 
		WHERE suministros = ?`, id)
	if err != nil {
		return Suministros{}, fmt.Errorf("LeerSuministros: error al obtener desglose: %w", err)
	}
	defer rows.Close()

	var desglose []SuministroDesglose
	for rows.Next() {
		var d SuministroDesglose
		var montoUnitario sql.NullFloat64

		if err := rows.Scan(&d.ID, &d.Suministros, &d.Nombre, &d.Articulo, &d.Agrupacion, &d.Cantidad, &montoUnitario); err != nil {
			return Suministros{}, fmt.Errorf("LeerSuministros: error al escanear desglose: %w", err)
		}

		d.MontoUnitario = montoUnitario.Float64
		desglose = append(desglose, d)
	}

	if err := rows.Err(); err != nil {
		return Suministros{}, fmt.Errorf("LeerSuministros: error al recorrer desglose: %w", err)
	}

	s.Desglose = desglose
	s.CuentaLoggeada = cuenta

	if s.Cuenta != cuenta && cuenta != "COES" && cuenta != "SF" {
		return Suministros{}, fmt.Errorf("LeerSuministros: cuenta '%s' no tiene acceso a este suministro", cuenta)
	}

	return s, nil
}

func ConfirmarEjecutadoSuministros(db *sql.DB, id, usuario, cuenta string, fecha time.Time, acuse, firma string) error {
	now := time.Now()
	oneMonthAgo := now.AddDate(0, -1, 0)

	if fecha.After(now) || fecha.Before(oneMonthAgo) {
		return fmt.Errorf("ConfirmarEjecutadoSuministros: invalid date")
	}

	_, err := UsuarioAcreditado(db, usuario, cuenta)
	if err != nil {
		return fmt.Errorf("ConfirmarEjecutadoSuministros: usuario %s no acreditado para cuenta %s: %w", usuario, cuenta, err)
	}

	var suministroID int
	err = db.QueryRow("SELECT id FROM suministros WHERE id = ? AND cuenta = ?", id, cuenta).Scan(&suministroID)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("ConfirmarEjecutadoSuministros: no matching suministro found for id %s and cuenta %s", id, cuenta)
		}
		return fmt.Errorf("ConfirmarEjecutadoSuministros: error retrieving suministro for id %s: %w", id, err)
	}

	query := `UPDATE suministros
		SET acuse_usuario = ?, acuse_fecha = ?, acuse = ?, acuse_firma = ?
		WHERE id = ?;`

	_, err = db.Exec(query, usuario, fecha, acuse, firma, suministroID)
	if err != nil {
		return fmt.Errorf("ConfirmarEjecutadoSuministros: failed to update suministro with id %d: %w", suministroID, err)
	}

	return nil
}

func SuministrosPendientesCOES(db *sql.DB, periodo int) ([]Suministros, error) {
	query := `
		SELECT id, emitido, emisor, cuenta, presupuesto, justif, firma,
		       coes, monto_bruto_total, geco, 
		       acuse_usuario, acuse_fecha, acuse, acuse_firma, notas
		FROM suministros
		WHERE coes = FALSE
		ORDER BY emitido DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("SuministrosPendientesCOES: error fetching suministros: %w", err)
	}
	defer rows.Close()

	var suministros []Suministros

	for rows.Next() {
		var s Suministros
		var emitido time.Time
		var acuseFecha sql.NullTime
		var acuseUsuario, acuse, acuseFirma, geco sql.NullString
		var montoBrutoTotal sql.NullFloat64
		var notas sql.NullString

		if err := rows.Scan(
			&s.ID, &emitido, &s.Emisor, &s.Cuenta, &s.Presupuesto, &s.Justif, &s.Firma,
			&s.COES, &montoBrutoTotal, &geco,
			&acuseUsuario, &acuseFecha, &acuse, &acuseFirma, &notas,
		); err != nil {
			return nil, fmt.Errorf("SuministrosPendientesCOES: error scanning row: %w", err)
		}

		if emitido.Year() == periodo {
			s.Emitido = emitido
			s.MontoBrutoTotal = montoBrutoTotal.Float64
			s.GECO = geco.String
			s.AcuseUsuario = acuseUsuario.String
			s.AcuseFecha = acuseFecha.Time
			s.Acuse = acuse.String
			s.AcuseFirma = acuseFirma.String
			s.Notas = notas.String

			suministros = append(suministros, s)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("SuministrosPendientesCOES: error iterating rows: %w", err)
	}

	return suministros, nil
}

func AprobarSuministroCOES(db *sql.DB, id string) error {
	_, err := db.Exec(`UPDATE suministros SET coes = TRUE WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("AprobarSuministroCOES: failed to update service: %w", err)
	}
	return nil
}
