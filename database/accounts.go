package database

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
)

type Account struct {
	ID     int    `db:"account_id"`
	Abbr   string `db:"account_abbr"`
	Name   string `db:"account_name"`
	Active bool   `db:"account_active"`
}

func (a *Account) load(db *sqlx.DB, accountID int) error {
	var err error = nil
	var account Account

	account, err = accountByAccountID(db, accountID)
	if err != nil {
		return logger.Errorf("error loading account: %v", err)
	}

	*a = account
	return nil
}

func AllowedAccountsByUserID(db *sqlx.DB, userID int) ([]Account, error) {
	const queryAllowedAccountsByID = `
		SELECT account_id, account_abbr, account_name, account_active
		FROM allowed_accounts
		WHERE user_id = ?
	`

	var accounts []Account
	err := db.Select(&accounts, queryAllowedAccountsByID, userID)
	if err != nil {
		return nil, logger.Errorf("error querying allowed_accounts for user_id %d: %v", userID, err)
	}

	return accounts, nil
}

func accountByAccountID(db *sqlx.DB, accountID int) (Account, error) {
	const queryAccountByID = `
		SELECT account_id, account_abbr, account_name, account_active
		FROM accounts
		WHERE account_id = ?
	`

	var account Account

	if err := db.Get(&account, queryAccountByID, accountID); err != nil {
		return Account{}, logger.Errorf("error selecting account from accounts: %v", err)
	}

	return account, nil
}

func IsAccountActive(db *sqlx.DB, accountID int) (bool, error) {
	const query = `
		SELECT account_active
		FROM accounts
		WHERE account_id = ?
	`

	var isActive bool
	if err := db.Get(&isActive, query, accountID); err != nil {
		return false, logger.Errorf("failed to check if account is active: %v", err)
	}

	return isActive, nil
}
