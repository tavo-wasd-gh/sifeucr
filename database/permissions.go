package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
)

type PermissionInteger int

type Permission struct {
	ID        int  `db:"permission_id"`
	UserID    int  `db:"permission_user"`
	AccountID int  `db:"permission_account"`
	Active    bool `db:"permission_active"`
	// Integer
	Integer PermissionInteger `db:"permission_integer"`
}

const (
	Read          PermissionInteger = 1 << iota // 1 << 0 = 1
	Write                                       // 1 << 1 = 2
	ReadOther                                   // 1 << 2 = 4
	WriteOther                                  // 1 << 3 = 8
	ReadAdvanced                                // 1 << 4 = 16
	WriteAdvanced                               // 1 << 5 = 32
)

func PermissionByUserIDAndAccountID(db *sqlx.DB, userID int, accountID int) (Permission, error) {
	const queryPermissionByUserIDAndAccountID = `
		SELECT permission_id, permission_user, permission_account, permission_integer, permission_active
		FROM permissions
		WHERE permission_user = ? AND permission_account = ?
	`

	var perm Permission

	if err := db.Get(&perm, queryPermissionByUserIDAndAccountID, userID, accountID); err != nil {
		return Permission{}, logger.Errorf("error selecting account from accounts: %v", err)
	}

	return perm, nil
}

func (p Permission) Has(required PermissionInteger) bool {
	return p.Integer&required != 0
}
