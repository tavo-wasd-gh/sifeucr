package config

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(connDvr, connStr string) (*sql.DB, bool, error) {
	db, err := sql.Open(connDvr, connStr)
	if err != nil {
		return nil, false, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, false, err
	}

    if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
        return nil, false, err
    }

	var firstTimeSetup bool
	err = db.QueryRow("SELECT NOT EXISTS (SELECT 1 FROM users)").Scan(&firstTimeSetup)
	if err != nil {
		return nil, false, err
	}

	if err := db.Ping(); err != nil {
		return nil, false, err
	}

	return db, firstTimeSetup, nil
}
