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
	Presupuesto string
	Justif      string
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

func suministrosInit(db *sql.DB, c string) ([]Suministros, error) {
	var suministros []Suministros

	query := `
	SELECT id, emitido, emisor, presupuesto, justif, coes, 
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
		var acuseFecha sql.NullTime
		var acuseUsuario, acuse, acuseFirma sql.NullString
		var geco sql.NullString
		var montoBrutoTotal sql.NullFloat64

		err := rows.Scan(
			&s.ID, &s.Emitido, &s.Emisor, &s.Presupuesto, &s.Justif, &s.COES,
			&montoBrutoTotal, &geco, &acuseUsuario, &acuseFecha,
			&acuse, &acuseFirma, &s.Notas,
		)
		if err != nil {
			return nil, fmt.Errorf("suministrosInit: error scanning row: %w", err)
		}

		s.MontoBrutoTotal = montoBrutoTotal.Float64
		s.GECO = geco.String
		s.AcuseUsuario = acuseUsuario.String
		s.AcuseFecha = acuseFecha.Time
		s.Acuse = acuse.String
		s.AcuseFirma = acuseFirma.String

		s.Desglose, err = suministroDesgloseInit(db, s.ID)
		if err != nil {
			return nil, fmt.Errorf("suministrosInit: error fetching desglose for suministro %d: %w", s.ID, err)
		}

		suministros = append(suministros, s)
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
