package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
)

type Session struct {
	UserID     int
	AccountID  int
}

func Init(connDvr, connStr string) (*sqlx.DB, error) {
	if connDvr == "" {
		connDvr = "sqlite3"
	}

	if connStr == "" {
		connStr = "./db.db"
	}

	db, err := sqlx.Open(connDvr, connStr)
	if err != nil {
		return nil, logger.Errorf("error opening connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, logger.Errorf("error pinging database: %v", err)
	}

	return db, nil
}

func (s *Session) Validate(db *sqlx.DB, requiredPermission PermissionInteger) error {
	if perm, err := PermissionByUserIDAndAccountID(db, s.UserID, s.AccountID); err != nil {
		return logger.Errorf("error checking permission")
	} else if !perm.Active {
		return logger.Errorf("inactive permission")
	} else if !perm.Has(requiredPermission) {
		return logger.Errorf("incorrect permission, required:%d got:%d", requiredPermission, perm.Integer)
	}

	if active, err := IsUserActive(db, s.UserID); err != nil {
		return logger.Errorf("%v", err)
	} else if !active {
		return logger.Errorf("inactive user")
	}

	if active, err := IsAccountActive(db, s.AccountID); err != nil {
		return logger.Errorf("%v", err)
	} else if !active {
		return logger.Errorf("inactive account")
	}

	return nil
}
