package main

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func fillData(data *Data, id_cuenta string) error {
	var (
		suministros []Suministros
	)

	rows, err := db.Query(`SELECT * FROM cuentas WHERE id_cuenta = $1`, id_cuenta)
	if err != nil {
		return err
	}
	defer rows.Close()

	var cuenta Cuenta
	if err := Scan(rows, &cuenta); err != nil {
		return err
	}

	cuenta.P1Total = cuenta.P1Servicios + cuenta.P1Suministros + cuenta.P1Bienes
	cuenta.P2Total = cuenta.P2Servicios + cuenta.P2Suministros + cuenta.P2Bienes

	rows, err = db.Query(`SELECT * FROM servicios WHERE id_cuenta = $1`, id_cuenta)
	if err != nil {
		return err
	}
	defer rows.Close()

	var servicios []Servicios
	if err := Scan(rows, &servicios); err != nil {
		return err
	}

	rows, err = db.Query(`SELECT * FROM suministros WHERE id_cuenta = $1`, id_cuenta)
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

	rows, err = db.Query(`SELECT * FROM bienes WHERE id_cuenta = $1`, id_cuenta)
	if err != nil {
		return err
	}
	defer rows.Close()

	var bienes []Bienes
	if err := Scan(rows, &bienes); err != nil {
		return err
	}

	rows, err = db.Query(`SELECT * FROM ajustes WHERE id_cuenta = $1`, id_cuenta)
	if err != nil {
		return err
	}
	defer rows.Close()

	var ajustes []Ajustes
	if err := Scan(rows, &ajustes); err != nil {
		return err
	}

	rows, err = db.Query(`SELECT * FROM donaciones WHERE id_cuenta_salida = $1`, id_cuenta)
	if err != nil {
		return err
	}
	defer rows.Close()

	var donaciones []Donaciones

	if err := Scan(rows, &donaciones); err != nil {
		return err
	}

	rows, err = db.Query(`SELECT * FROM donaciones WHERE id_cuenta_entrada = $1`, id_cuenta)
	if err != nil {
		return err
	}
	defer rows.Close()

	if err := Scan(rows, &donaciones); err != nil {
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

func calcularRestante(data *Data, tipo string) (float64, error) {
	switch tipo {
	case "PGeneral":
		return data.Cuenta.PGeneral - calcularPGeneral(data), nil
	case "P1Servicios":
		return data.Cuenta.P1Servicios - calcularServicios(data, data.Cuenta.P1Validez), nil
	case "P1Suministros":
		return data.Cuenta.P1Suministros - calcularSuministros(data, data.Cuenta.P1Validez), nil
	case "P1Bienes":
		return data.Cuenta.P1Bienes - calcularBienes(data, data.Cuenta.P1Validez), nil
	case "P1Total":
		return data .Cuenta.P1Total - calcularTotal(data, data.Cuenta.P1Validez), nil
	case "P2Servicios":
		return data.Cuenta.P2Servicios - calcularServicios(data, data.Cuenta.P2Validez), nil
	case "P2Suministros":
		return data.Cuenta.P2Suministros - calcularSuministros(data, data.Cuenta.P2Validez), nil
	case "P2Bienes":
		return data.Cuenta.P2Bienes - calcularBienes(data, data.Cuenta.P2Validez), nil
	case "P2Total":
		return data.Cuenta.P2Total - calcularTotal(data, data.Cuenta.P2Validez), nil
	default:
		return 0, fmt.Errorf("tipo '%s' no reconocido", tipo)
	}
}

func calcularPGeneral(data *Data) float64 {
	var total float64

	for _, servicio := range data.Servicios {
		total += servicio.MontoBruto
	}

	for _, suministro := range data.Suministros {
		total += suministro.MontoBrutoTotal
	}

	for _, bien := range data.Bienes {
		total += bien.MontoBruto
	}

	for _, ajuste := range data.Ajustes {
		total += ajuste.MontoBruto
	}

	for _, donacion := range data.Donaciones {
		total += donacion.MontoBruto
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

	for _, ajuste := range data.Ajustes {
		if ajuste.Emitido.Valid && validez.Valid && ajuste.Emitido.Time.Before(validez.Time) {
			total += ajuste.MontoBruto
		}
	}

	for _, donacion := range data.Donaciones {
		if donacion.Emitido.Valid && validez.Valid && donacion.Emitido.Time.Before(validez.Time) {
			if donacion.IDCuentaEntrada == data.Cuenta.IDCuenta {
				total += donacion.MontoBruto
			} else if donacion.IDCuentaSalida == data.Cuenta.IDCuenta {
				total -= donacion.MontoBruto
			} else {
				log.Println("Neither")
			}
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

func Scan(rows *sql.Rows, dest interface{}) error {
	destValue := reflect.ValueOf(dest)

	// Validate `dest` is a pointer
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer to a struct or slice")
	}
	destValue = destValue.Elem()

	if destValue.Kind() == reflect.Struct {
		// Handle a single struct
		return fillSingleStruct(rows, destValue)
	} else if destValue.Kind() == reflect.Slice {
		// Handle a slice of structs
		return fillSlice(rows, destValue)
	}

	return fmt.Errorf("dest must be a pointer to a struct or slice of structs")
}

// fillSingleStruct handles populating a single struct from one row.
func fillSingleStruct(rows *sql.Rows, destValue reflect.Value) error {
	destType := destValue.Type()

	// Map struct fields by their `db` tags
	columnMap := make(map[string]int)
	for i := 0; i < destType.NumField(); i++ {
		field := destType.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag != "" {
			columnMap[dbTag] = i
		}
	}

	// Get database column names
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Prepare pointers for scanning
	values := make([]interface{}, len(columns))
	for i, columnName := range columns {
		if fieldIndex, found := columnMap[columnName]; found {
			field := destValue.Field(fieldIndex)
			if field.CanSet() {
				values[i] = field.Addr().Interface()
			} else {
				var placeholder interface{}
				values[i] = &placeholder
			}
		} else {
			var placeholder interface{}
			values[i] = &placeholder
		}
	}

	// Read and scan a single row
	if !rows.Next() {
		return sql.ErrNoRows // or fmt.Errorf("no rows found")
	}
	if err := rows.Scan(values...); err != nil {
		return fmt.Errorf("failed to scan row: %w", err)
	}

	return nil
}

// fillSlice handles populating a slice of structs from multiple rows.
func fillSlice(rows *sql.Rows, destValue reflect.Value) error {
	sliceElementType := destValue.Type().Elem()

	if sliceElementType.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a slice of structs")
	}

	// Map struct fields by their `db` tags
	columnMap := make(map[string]int)
	for i := 0; i < sliceElementType.NumField(); i++ {
		field := sliceElementType.Field(i)
		dbTag := field.Tag.Get("db")
		if dbTag != "" {
			columnMap[dbTag] = i
		}
	}

	// Get database column names
	columns, err := rows.Columns()
	if err != nil {
		return fmt.Errorf("failed to get columns: %w", err)
	}

	// Iterate over rows and populate the slice
	for rows.Next() {
		newStruct := reflect.New(sliceElementType).Elem()

		// Prepare pointers for scanning
		values := make([]interface{}, len(columns))
		for i, columnName := range columns {
			if fieldIndex, found := columnMap[columnName]; found {
				field := newStruct.Field(fieldIndex)
				if field.CanSet() {
					values[i] = field.Addr().Interface()
				} else {
					var placeholder interface{}
					values[i] = &placeholder
				}
			} else {
				var placeholder interface{}
				values[i] = &placeholder
			}
		}

		// Scan the row into the struct
		if err := rows.Scan(values...); err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		// Append the struct to the slice
		destValue.Set(reflect.Append(destValue, newStruct))
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("error iterating rows: %w", err)
	}

	return nil
}
