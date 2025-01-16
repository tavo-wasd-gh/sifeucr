package db

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
	// _ "github.com/lib/pq"
)

func Init(connDvr, connStr string) (*sql.DB, error) {
	db, err := sql.Open(connDvr, connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}
