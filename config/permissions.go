package config

const (
	Read          int64 = 1 << iota // 1 << 0 = 1
	Write                           // 1 << 1 = 2
	ReadOther                       // 1 << 2 = 4
	WriteOther                      // 1 << 3 = 8
	ReadAdvanced                    // 1 << 4 = 16
	WriteAdvanced                   // 1 << 5 = 32
)

type PermissionType struct {
	Name  string
	Label string
	Bit   int64
}

var PermissionTypes = []PermissionType{
	{"read", "Lectura", Read},
	{"write", "Escritura", Write},
	{"readOther", "Lectura Otros", ReadOther},
	{"writeOther", "Escritura Otros", WriteOther},
	{"readAdvanced", "Lectura Avanzado", ReadAdvanced},
	{"writeAdvanced", "Escritura Avanzado", WriteAdvanced},
}

func HasPermission(userPerm, requiredPerm int64) bool {
	return userPerm&requiredPerm == requiredPerm
}
