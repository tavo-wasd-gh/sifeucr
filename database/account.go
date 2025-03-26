package database

import (
	"github.com/jmoiron/sqlx"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
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

	budgets, err := budgetsByAccount(db, id)
	if err != nil {
		return account, logger.Errorf("error querying budgets by account '%s': %v", id, err)
	}

	account.Budgets = budgets

	return account, nil
}

func accountsByUser(db *sqlx.DB, email string) ([]Account, error) {
	var accounts []Account

	query := `
	SELECT a.*
	FROM accounts a
	JOIN permissions p ON a.id = p.account
	WHERE p.user = ?;
	`

	if err := db.Select(&accounts, query, email); err != nil {
		return nil, logger.Errorf("error querying accounts by user '%s': %v", email, err)
	}

	if len(accounts) == 0 {
		return nil, logger.Errorf("no accounts found for user '%s'", email)
	}

	return accounts, nil
}

func (a *Account) Register(db *sqlx.DB) error {
	var exists bool

	err := db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM accounts WHERE id = ?)`, a.ID)
	if err != nil {
		return logger.Errorf("error checking account existence: %v", err)
	}

	if exists {
		return logger.Errorf("account with id '%s' already exists", a.ID)
	}

	_, err = db.Exec(
		`INSERT INTO accounts (id, name, teeu, coes) VALUES (?, ?, ?, ?)`,
		a.ID, a.Name, a.TEEU, a.COES,
	)
	if err != nil {
		return logger.Errorf("error inserting account: %v", err)
	}

	return nil
}
