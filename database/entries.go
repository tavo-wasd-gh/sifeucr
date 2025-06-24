package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
)

type Entry struct {
	ID     int     `db:"dist_id"`
	Year   int     `db:"dist_name"`
	Code   int     `db:"dist_entry"`
	Object string  `db:"dist_account"`
	Amount float64 `db:"dist_valid_until"`
}

func selectEntryByID(db *sqlx.DB, entryID int) (Entry, error) {
	const queryEntryID = `
		SELECT entry_id, entry_year, entry_code, entry_object, entry_amount
		FROM budget_entries WHERE entry_id = ?
	`

	var entry Entry
	err := db.Get(&entry, queryEntryID, entryID)
	if err != nil {
		return Entry{}, logger.Errorf("error querying entry %d: %v", entryID, err)
	}

	return entry, nil
}

func selectCurrentEntries(db *sqlx.DB, entryID int) ([]Entry, error) {
	const queryEntryID = `
		SELECT entry_id, entry_year, entry_code, entry_object, entry_amount
		FROM budget_entries
		WHERE entry_year = CAST(strftime('%Y', 'now') AS INTEGER);
	`

	var entries []Entry
	err := db.Select(&entries, queryEntryID, entryID)
	if err != nil {
		return nil, logger.Errorf("error querying entry %d: %v", entryID, err)
	}

	return entries, nil
}
