package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
)

type Dashboard struct {
	User       User
	Account    Account
	// Advanced
	ReadAdvanced bool
}

func (d *Dashboard) Load(db *sqlx.DB, userID int, accountID int) error {
	perm, err := PermissionByUserIDAndAccountID(db, userID, accountID);
	if err != nil {
		return logger.Errorf("error loading user: %v", err)
	}

	account, err :=  accountByAccountID(db, accountID)
	if err != nil {
		return logger.Errorf("error loading account: %v", err)
	}

	user, err := userByUserID(db, userID);
	if err != nil {
		return logger.Errorf("error loading user: %v", err)
	}

	// TODO: Load Servicios
	// TODO: Load Suministros
	// TODO: Load Bienes

	d.ReadAdvanced = perm.Has(ReadAdvanced)

	d.User = user
	d.Account = account

	return nil
}
