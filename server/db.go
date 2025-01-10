package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func fillData(data *Data, id_cuenta string) error {
	var (
		servicios   []Servicios
		suministros []Suministros
		bienes      []Bienes
		ajustes     []Ajustes
		donaciones  []Donaciones
	)

	query := `SELECT * FROM cuentas WHERE id_cuenta = $1`
	rows, err := db.Query(query, id_cuenta)
	if err != nil {
		return err
	}
	defer rows.Close()

	var cuenta Cuenta
	if err := fillStruct(rows, &cuenta); err != nil {
		return err
	}

	cuenta.P1Total = cuenta.P1Servicios + cuenta.P1Suministros + cuenta.P1Bienes
	cuenta.P2Total = cuenta.P2Servicios + cuenta.P2Suministros + cuenta.P2Bienes

	query = `SELECT * FROM servicios WHERE id_cuenta = $1`
	rows, err = db.Query(query, id_cuenta)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var servicio Servicios
		err := rows.Scan(
			&servicio.IDServicios,
			&servicio.Emitido,
			&servicio.IDCuenta,
			&servicio.Detalle,
			&servicio.MontoBruto,
			&servicio.MontoIVA,
			&servicio.MontoDesc,
			&servicio.JustifServ,
			&servicio.ProvNom,
			&servicio.ProvCed,
			&servicio.ProvDirec,
			&servicio.ProvEmail,
			&servicio.ProvTel,
			&servicio.ProvBanco,
			&servicio.ProvIBAN,
			&servicio.JustifProv,
			&servicio.COES,
			&servicio.GecoSol,
			&servicio.GecoOCS,
			&servicio.PorEjecutar,
			&servicio.Ejecutado,
			&servicio.Pagado,
			&servicio.Notas,
		)
		if err != nil {
			return err
		}

		servicios = append(servicios, servicio)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	query = `SELECT * FROM suministros WHERE id_cuenta = $1`
	rows, err = db.Query(query, id_cuenta)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var suministro Suministros
		err := rows.Scan(
			&suministro.IDSuministros,
			&suministro.Emitido,
			&suministro.IDCuenta,
			&suministro.JustifSum,
			&suministro.COES,
			&suministro.Geco,
			&suministro.Notas,
		)
		if err != nil {
			return err
		}

		desgloseQuery := `SELECT * FROM suministros_desglose WHERE id_suministros = $1`
		desgloseRows, err := db.Query(desgloseQuery, suministro.IDSuministros)
		if err != nil {
			return err
		}
		defer desgloseRows.Close()

		var desglose []SuministrosDesglose
		var montoBrutoTotal float64
		for desgloseRows.Next() {
			var item SuministrosDesglose
			err := desgloseRows.Scan(
				&item.ID,
				&item.IDSuministros,
				&item.IDItem,
				&item.Nombre,
				&item.Cantidad,
				&item.MontoBrutoUnidad,
			)
			if err != nil {
				return err
			}
			desglose = append(desglose, item)
			montoBrutoTotal += item.MontoBrutoUnidad * float64(item.Cantidad)
		}
		if err := desgloseRows.Err(); err != nil {
			return err
		}

		suministro.Desglose = desglose
		suministro.MontoBrutoTotal = montoBrutoTotal
		suministros = append(suministros, suministro)
	}

	if err := rows.Err(); err != nil {
		return err
	}

	query = `SELECT * FROM bienes WHERE id_cuenta = $1`
	rows, err = db.Query(query, id_cuenta)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var bien Bienes
		err := rows.Scan(
			&bien.IDBienes,
			&bien.Emitido,
			&bien.IDCuenta,
			&bien.Detalle,
			&bien.MontoBruto,
			&bien.MontoIVA,
			&bien.MontoDesc,
			&bien.JustifBien,
			&bien.ProvNom,
			&bien.ProvCed,
			&bien.ProvDirec,
			&bien.ProvEmail,
			&bien.ProvTel,
			&bien.ProvBanco,
			&bien.ProvIBAN,
			&bien.JustifProv,
			&bien.COES,
			&bien.GecoSol,
			&bien.GecoOC,
			&bien.Recibido,
			&bien.Notas,
		)
		if err != nil {
			return err
		}

		bienes = append(bienes, bien)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	query = `SELECT * FROM ajustes WHERE id_cuenta = $1`
	rows, err = db.Query(query, id_cuenta)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var ajuste Ajustes
		err := rows.Scan(
			&ajuste.IDAjustes,
			&ajuste.Emitido,
			&ajuste.IDCuenta,
			&ajuste.Partida,
			&ajuste.Detalle,
			&ajuste.MontoBruto,
			&ajuste.Notas,
		)
		if err != nil {
			return err
		}

		ajustes = append(ajustes, ajuste)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	*data = Data{
		Cuenta:               cuenta,
		Periodo:              time.Now().Year(),
		Servicios:            servicios,
		Suministros:          suministros,
		Bienes:               bienes,
		Ajustes:              ajustes,
		Donaciones:           donaciones,
	}

	return nil
}

func calcularEmitido(data *Data, tipo string) (float64, error) {
	switch tipo {
	case "PGeneral":
		return calcularPGeneral(data), nil
	case "P1Servicios":
		return calcularServicios(data, data.Cuenta.P1Validez), nil
	case "P1Suministros":
		return calcularSuministros(data, data.Cuenta.P1Validez), nil
	case "P1Bienes":
		return calcularBienes(data, data.Cuenta.P1Validez), nil
	case "P1Total":
		return calcularTotal(data, data.Cuenta.P1Validez), nil
	case "P2Servicios":
		return calcularServicios(data, data.Cuenta.P2Validez), nil
	case "P2Suministros":
		return calcularSuministros(data, data.Cuenta.P2Validez), nil
	case "P2Bienes":
		return calcularBienes(data, data.Cuenta.P2Validez), nil
	case "P2Total":
		return calcularTotal(data, data.Cuenta.P2Validez), nil
	default:
		return 0, fmt.Errorf("tipo '%s' no reconocido", tipo)
	}
}

func calcularPGeneral(data *Data) float64 {
	var total float64
	for _, servicio := range data.Servicios {
		total += servicio.MontoBruto
	}
	return total
}

func calcularServicios(data *Data, validez sql.NullTime) float64 {
	var total float64
	for _, servicio := range data.Servicios {
		if servicio.Emitido.Valid && validez.Valid && servicio.Emitido.Time.Before(validez.Time) {
			total += servicio.MontoBruto
		}
	}
	return total
}

func calcularSuministros(data *Data, validez sql.NullTime) float64 {
	var total float64
	for _, suministro := range data.Suministros {
		if suministro.Emitido.Valid && validez.Valid && suministro.Emitido.Time.Before(validez.Time) {
			for _, desglose := range suministro.Desglose {
				total += desglose.MontoBrutoUnidad * float64(desglose.Cantidad)
			}
		}
	}
	return total
}

func calcularBienes(data *Data, validez sql.NullTime) float64 {
	var total float64
	for _, bien := range data.Bienes {
		if bien.Emitido.Valid && validez.Valid && bien.Emitido.Time.Before(validez.Time) {
			total += bien.MontoBruto
		}
	}
	return total
}

func calcularAjustes(data *Data, validez sql.NullTime) float64 {
	var total float64
	for _, ajuste := range data.Ajustes {
		if ajuste.Emitido.Valid && validez.Valid && ajuste.Emitido.Time.Before(validez.Time) {
			total += ajuste.MontoBruto
		}
	}
	return total
}

func calcularDonaciones(data *Data, validez sql.NullTime) float64 {
	var total float64
	for _, donacion := range data.Donaciones {
		if donacion.Emitido.Valid && validez.Valid && donacion.Emitido.Time.Before(validez.Time) {
			total += donacion.MontoBruto
		}
	}
	return total
}

func calcularTotal(data *Data, validez sql.NullTime) float64 {
	totalServicios := calcularServicios(data, validez)
	totalBienes := calcularBienes(data, validez)
	totalSuministros := calcularSuministros(data, validez)
	totalAjustes := calcularAjustes(data, validez)
	totalDonaciones := calcularDonaciones(data, validez)
	return totalServicios + totalBienes + totalSuministros + totalAjustes + totalDonaciones
}

func fillStruct(rows *sql.Rows, dest interface{}) error {
	destValue := reflect.ValueOf(dest).Elem()
	destType := destValue.Type()

	columnMap := make(map[string]int)
	for i := 0; i < destType.NumField(); i++ {
		field := destType.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag != "" {
			columnMap[dbTag] = i
		}
	}

	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	values := make([]interface{}, len(columns))
	fieldPointers := make([]interface{}, len(columns))

	for i, columnName := range columns {
		if fieldIndex, found := columnMap[columnName]; found {
			field := destValue.Field(fieldIndex)
			fieldPointers[i] = field.Addr().Interface()
		} else {
			var placeholder interface{}
			fieldPointers[i] = &placeholder
		}
		values[i] = fieldPointers[i]
	}

	if rows.Next() {
		if err := rows.Scan(values...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}
	}

	return nil
}
