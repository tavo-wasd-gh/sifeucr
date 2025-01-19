package session

type Cuenta struct {
	ID          string `db:"id"`
	Privilegio  uint64 `db:"privilegio"`
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
	Servicios   []Servicio
	Suministros []Suministros
	Bienes      []Bien
	Donaciones  []Donacion
	Ajustes     []Ajuste
}
