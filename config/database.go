package config

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(connDvr, connStr string) (*sql.DB, error) {
	db, err := sql.Open(connDvr, connStr)
	if err != nil {
		return nil, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, err
	}

    if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
        return nil, err
    }

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
