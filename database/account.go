package database

import (
	"github.com/tavo-wasd-gh/webtoolkit/logger"
	"github.com/jmoiron/sqlx"
)

type Account struct {
	ID   string `db:"id"`
	Name string `db:"name"`
	TEEU bool   `db:"teeu"`
	COES bool   `db:"coes"`
	// Runtime
	Budgets []Budget
}

func accountByID(db *sqlx.DB, id string) (Account, error) {
	var account Account

	if err := db.Get(&account, "SELECT * FROM accounts WHERE id = $1", id); err != nil {
		return account, logger.Errorf("error querying account by id '%s': %v", id, err)
	}

	budgets, err := budgetsByAccount(db, id); 
	if err != nil {
		return account, logger.Errorf("error querying budgets by account '%s': %v", id, err)
	}

	account.Budgets = budgets

	return account, nil
}

func accountsByUser(db *sqlx.DB, email string) ([]Account, error) {
	var accounts []Account

	query := `
	SELECT a.id, a.name, a.teeu, a.coes
	FROM accounts a
	INNER JOIN permissions p ON a.id = p.account
	WHERE p.user = ?
	`

	if err := db.Select(&accounts, query, email); err != nil {
		return nil, logger.Errorf("error querying accounts by user '%s': %v", email, err)
	}

	if len(accounts) == 0 {
		return nil, logger.Errorf("no accounts found for user '%s'", email)
	}

	return accounts, nil
}
