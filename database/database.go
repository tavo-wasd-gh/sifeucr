package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
)

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

func SetupNeeded(db *sqlx.DB) (bool, error) {
	var count int
	err := db.Get(&count, "SELECT COUNT(*) FROM users")
	if err != nil {
		return false, logger.Errorf("failed to query users: %v", err)
	}
	return count == 0, nil
}
