package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
)

type Panel struct {
	Users []User
}

func (p *Panel) Load(db *sqlx.DB) error {
	users, err := queryAllUsers(db)
	if err != nil {
		return logger.Errorf("error selecting account from accounts: %v", err)
	}

	p.Users = users

	return nil
}
