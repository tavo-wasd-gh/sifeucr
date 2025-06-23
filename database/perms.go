package database

type PermissionInteger int

const (
    Read PermissionInteger = 1 << iota // 1 << 0 = 1
    Write                       // 1 << 1 = 2
    ReadOther                   // 1 << 2 = 4
    WriteOther                  // 1 << 3 = 8
    ReadAdvanced                // 1 << 4 = 16
    WriteAdvanced               // 1 << 5 = 32
)
