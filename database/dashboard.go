package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
)

type Dashboard struct {
	User       User
	Account    Account
	Permission Permission
	// Advanced
	ReadAdvanced bool
}

func (d *Dashboard) Load(db *sqlx.DB, userID int, accountID int) error {
	const requiredPermission = Read

	var perm Permission
	if err := perm.load(db, userID, accountID); err != nil {
		return logger.Errorf("error loading user: %v", err)
	}

	if !perm.Active {
		return logger.Errorf("error loading permission: inactive permission")
	}

	if !perm.Has(requiredPermission) {
		return logger.Errorf("permission error: required:%d got:%d", requiredPermission, perm.Integer)
	}

	var account Account
	if err := account.load(db, accountID); err != nil {
		return logger.Errorf("error loading account: %v", err)
	}

	if account.Active == false {
		return logger.Errorf("error loading account: inactive account")
	}

	var user User
	if err := user.load(db, userID); err != nil {
		return logger.Errorf("error loading user: %v", err)
	}

	if user.Active == false {
		return logger.Errorf("error loading user: inactive user")
	}

	// TODO: Load Servicios
	// TODO: Load Suministros
	// TODO: Load Bienes

	if perm.Has(ReadAdvanced) {
		// TODO: Load Control madre
		d.ReadAdvanced = true
	}

	d.Account = account
	d.User = user
	d.Permission = perm

	return nil
}
