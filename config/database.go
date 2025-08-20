package config

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

const (
	DEFAULT_DRIVER = "sqlite"
)

func InitDB(dbFile string) (*sql.DB, bool, error) {
	dbFileNonExistent := false
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		dbFileNonExistent = true
	}

	db, err := sql.Open(DEFAULT_DRIVER, dbFile)
	if err != nil {
		return nil, false, err
	}

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return nil, false, err
	}

	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		return nil, false, err
	}

	if err := db.Ping(); err != nil {
		return nil, false, err
	}

	if dbFileNonExistent {
		if err := applySchema(db, schema); err != nil {
			_ = db.Close()
			return nil, false, err
		}
	}

	var firstTimeSetup bool
	err = db.QueryRow("SELECT NOT EXISTS (SELECT 1 FROM users)").Scan(&firstTimeSetup)
	if err != nil {
		return nil, false, err
	}

	return db, firstTimeSetup, nil
}

func applySchema(db *sql.DB, ddl string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ddl); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("error applying schema: %w", err)
	}
	// _, _ = tx.Exec(`PRAGMA user_version = 1`)
	return tx.Commit()
}
