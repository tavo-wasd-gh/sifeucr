package session

import (
	tr "github.com/tavo-wasd-gh/sifeucr/user/transactions"
)

type Cuenta struct {
	ID          string `db:"id"`
	Nombre      string `db:"nombre"`
	Presidencia string `db:"presidencia"`
	Tesoreria   string `db:"tesoreria"`
	PGID        string `db:"pg"`
	P1ID        string `db:"p1"`
	P2ID        string `db:"p2"`
	PG          Presupuesto
	P1          Presupuesto
	P2          Presupuesto
	TEEU        bool `db:"teeu"`
	COES        bool `db:"coes"`
	Periodo     int
	Servicios   []tr.Servicio
	Suministros []tr.Suministros
	Bienes      []tr.Bien
	Donaciones  []tr.Donacion
	Ajustes     []tr.Ajuste
}
