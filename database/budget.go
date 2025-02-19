package database

import (
	"time"

	"github.com/tavo-wasd-gh/webtoolkit/logger"
	"github.com/jmoiron/sqlx"
)

type Budget struct {
	ID       string     `db:"id"`
	Valid    *time.Time `db:"valid"`
	Services float64    `db:"services"`
	Supplies float64    `db:"supplies"`
	Goods    float64    `db:"goods"`
}

func budgetsByAccount(db *sqlx.DB, account string) ([]Budget, error) {
	var budgets []Budget

	if err := db.Select(&budgets, "SELECT * FROM budgets WHERE account = $1", account); err != nil {
		return nil, logger.Errorf("error initializing budgets: %v", err)
	}

	return budgets, nil
}
