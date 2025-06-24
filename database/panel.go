package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
)

type Panel struct {
	User       User
	Account    Account
	Permission Permission
	// Advanced
	ReadAdvanced  bool
	// Manage users
	Users []User
}

func (p *Panel) Load(db *sqlx.DB, userID, accountID int) error {
	const requiredPermission = ReadAdvanced

	var perm Permission
	if err := perm.load(db, userID, accountID); err != nil {
		return logger.Errorf("error loading user: %v", err)
	}

	if !perm.Active {
		return logger.Errorf("error loading permission: inactive permission")
	}

	if !perm.has(requiredPermission) {
		return logger.Errorf("permission error: required:%d got:%d", requiredPermission, perm.Integer)
	}

	users, err := queryAllUsers(db)
	if err != nil {
		return logger.Errorf("error selecting account from accounts: %v", err)
	}

	p.Users = users

	return nil
}
