package main

import (
	"time"
)

type Cuenta struct {
	IDCuenta      string    `json:"id_cuenta" db:"id_cuenta"`
	Nombre        string    `json:"nombre" db:"nombre"`
	PGeneral      float64   `json:"p_general" db:"p_general"`
	P1Servicios   float64   `json:"p1_servicios" db:"p1_servicios"`
	P1Suministros float64   `json:"p1_suministros" db:"p1_suministros"`
	P1Bienes      float64   `json:"p1_bienes" db:"p1_bienes"`
	P1Validez     time.Time `json:"p1_validez" db:"p1_validez"`
	P2Servicios   float64   `json:"p2_servicios" db:"p2_servicios"`
	P2Suministros float64   `json:"p2_suministros" db:"p2_suministros"`
	P2Bienes      float64   `json:"p2_bienes" db:"p2_bienes"`
	P2Validez     time.Time `json:"p2_validez" db:"p2_validez"`
	TEEU          bool      `json:"teeu" db:"teeu"`
	COES          bool      `json:"coes" db:"coes"`
}

type Servicios struct {
	IDServicios int       `json:"id_servicios" db:"id_servicios"`
	Emitido     time.Time `json:"emitido" db:"emitido"`
	IDCuenta    string    `json:"id_cuenta" db:"id_cuenta"`
	Detalle     string    `json:"detalle" db:"detalle"`
	MontoBruto  float64   `json:"monto_bruto" db:"monto_bruto"`
	MontoIVA    float64   `json:"monto_iva" db:"monto_iva"`
	MontoDesc   float64   `json:"monto_desc" db:"monto_desc"`
	JustifServ  string    `json:"justif_serv" db:"justif_serv"`
	ProvNom     string    `json:"prov_nom" db:"prov_nom"`
	ProvCed     string    `json:"prov_ced" db:"prov_ced"`
	ProvDirec   string    `json:"prov_direc" db:"prov_direc"`
	ProvEmail   string    `json:"prov_email" db:"prov_email"`
	ProvTel     string    `json:"prov_tel" db:"prov_tel"`
	ProvBanco   string    `json:"prov_banco" db:"prov_banco"`
	ProvIBAN    string    `json:"prov_iban" db:"prov_iban"`
	JustifProv  string    `json:"justif_prov" db:"justif_prov"`
	COES        bool      `json:"coes" db:"coes"`
	GecoSol     string    `json:"geco_sol" db:"geco_sol"`
	GecoOCS     string    `json:"geco_ocs" db:"geco_ocs"`
	PorEjecutar time.Time `json:"por_ejecutar" db:"por_ejecutar"`
	Ejecutado   time.Time `json:"ejecutado" db:"ejecutado"`
	Pagado      time.Time `json:"pagado" db:"pagado"`
	Notas       string    `json:"notas" db:"notas"`
}

type Suministros struct {
	IDSuministros int       `json:"id_suministros" db:"id_suministros"`
	Emitido       time.Time `json:"emitido" db:"emitido"`
	IDCuenta      string    `json:"id_cuenta" db:"id_cuenta"`
	JustifSum     string    `json:"justif_sum" db:"justif_sum"`
	COES          bool      `json:"coes" db:"coes"`
	Geco          string    `json:"geco" db:"geco"`
	Notas         string    `json:"notas" db:"notas"`
}

type SuministrosDesglose struct {
	ID               int     `json:"id" db:"id"`
	IDSuministros    int     `json:"id_suministros" db:"id_suministros"`
	IDItem           string  `json:"id_item" db:"id_item"`
	Nombre           string  `json:"nombre" db:"nombre"`
	Cantidad         int     `json:"cantidad" db:"cantidad"`
	MontoBrutoUnidad float64 `json:"monto_bruto_unidad" db:"monto_bruto_unidad"`
}

type Bienes struct {
	IDBienes   int       `json:"id_bienes" db:"id_bienes"`
	Emitido    time.Time `json:"emitido" db:"emitido"`
	IDCuenta   string    `json:"id_cuenta" db:"id_cuenta"`
	Detalle    string    `json:"detalle" db:"detalle"`
	MontoBruto float64   `json:"monto_bruto" db:"monto_bruto"`
	MontoIVA   float64   `json:"monto_iva" db:"monto_iva"`
	MontoDesc  float64   `json:"monto_desc" db:"monto_desc"`
	JustifBien string    `json:"justif_bien" db:"justif_bien"`
	ProvNom    string    `json:"prov_nom" db:"prov_nom"`
	ProvCed    string    `json:"prov_ced" db:"prov_ced"`
	ProvDirec  string    `json:"prov_direc" db:"prov_direc"`
	ProvEmail  string    `json:"prov_email" db:"prov_email"`
	ProvTel    string    `json:"prov_tel" db:"prov_tel"`
	ProvBanco  string    `json:"prov_banco" db:"prov_banco"`
	ProvIBAN   string    `json:"prov_iban" db:"prov_iban"`
	JustifProv string    `json:"justif_prov" db:"justif_prov"`
	COES       bool      `json:"coes" db:"coes"`
	GecoSol    string    `json:"geco_sol" db:"geco_sol"`
	GecoOC     string    `json:"geco_oc" db:"geco_oc"`
	Recibido   time.Time `json:"recibido" db:"recibido"`
	Notas      string    `json:"notas" db:"notas"`
}

type Ajustes struct {
	IDAjustes  int       `json:"id_ajustes" db:"id_ajustes"`
	Emitido    time.Time `json:"emitido" db:"emitido"`
	IDCuenta   string    `json:"id_cuenta" db:"id_cuenta"`
	Partida    string    `json:"partida" db:"partida"`
	Detalle    string    `json:"detalle" db:"detalle"`
	MontoBruto float64   `json:"monto_bruto" db:"monto_bruto"`
	Notas      string    `json:"notas" db:"notas"`
}

type Donaciones struct {
	IDBienes        int       `json:"id_bienes" db:"id_bienes"`
	Emitido         time.Time `json:"emitido" db:"emitido"`
	IDCuentaSalida  string    `json:"id_cuenta_salida" db:"id_cuenta_salida"`
	PartidaSalida   string    `json:"partida_salida" db:"partida_salida"`
	IDCuentaEntrada string    `json:"id_cuenta_entrada" db:"id_cuenta_entrada"`
	PartidaEntrada  string    `json:"partida_entrada" db:"partida_entrada"`
	Detalle         string    `json:"detalle" db:"detalle"`
	MontoBruto      float64   `json:"monto_bruto" db:"monto_bruto"`
	CartaCOES       string    `json:"carta_coes" db:"carta_coes"`
	Notas           string    `json:"notas" db:"notas"`
}

type Dashboard struct {
	Cuenta      Cuenta `json:"id_cuenta" db:"id_cuenta"`
	Periodo     uint8
	Servicios   []Servicios
	Suministros []Suministros
	Bienes      []Bienes
	Ajustes     []Ajustes
	Donaciones  []Donaciones
}
