package main

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

var db *sql.DB

func fillData(data *Data, id_cuenta string) error {
	var (
		cuenta Cuenta
		servicios []Servicios
		suministros []Suministros
		bienes []Bienes
		ajustes []Ajustes
		donaciones []Donaciones
	)

	query := `SELECT * FROM cuentas WHERE id_cuenta = $1`
	rows, err := db.Query(query, id_cuenta)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(
			&cuenta.IDCuenta,
			&cuenta.Nombre,
			&cuenta.Presidencia,
			&cuenta.Tesoreria,
			&cuenta.PGeneral,
			&cuenta.P1Servicios,
			&cuenta.P1Suministros,
			&cuenta.P1Bienes,
			&cuenta.P1Validez,
			&cuenta.P2Servicios,
			&cuenta.P2Suministros,
			&cuenta.P2Bienes,
			&cuenta.P2Validez,
			&cuenta.TEEU,
			&cuenta.COES,
		)
		if err != nil {
			return err
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

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
		Cuenta: cuenta,
		Periodo: time.Now().Year(),
		Servicios: servicios,
		Suministros: suministros,
		Bienes: bienes,
		Ajustes: ajustes,
		Donaciones: donaciones,
		PGeneralEmitido: 0,
		P1ServiciosEmitido: 0,
		P1SuministrosEmitido: 0,
		P1BienesEmitido: 0,
		P1Total: 0,
		P1Emitido: 0,
		P2ServiciosEmitido: 0,
		P2SuministrosEmitido: 0,
		P2BienesEmitido: 0,
		P2Total: 0,
		P2Emitido: 0,
	}

	return nil
}
