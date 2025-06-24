package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
)

type Distribution struct {
	ID         int     `db:"dist_id"`
	Name       string  `db:"dist_name"`
	Entry      int     `db:"dist_entry"`
	Account    int     `db:"dist_account"`
	ValidUntil int     `db:"dist_valid_until"`
	Amount     float64 `db:"dist_amount"`
	Active     bool    `db:"dist_active"`
}

func (r *Distribution) load(db *sqlx.DB, distributionID int) error {
	var err error = nil
	var request Distribution

	request, err = distributionByID(db, distributionID)
	if err != nil {
		return logger.Errorf("error loading distribution: %v", err)
	}

	*r = request
	return nil
}

func distributionByID(db *sqlx.DB, distributionID int) (Distribution, error) {
	return Distribution{}, nil
}
