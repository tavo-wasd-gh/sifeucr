package database

import (
	"strings"

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

type User struct {
	ID     int    `db:"user_id"`
	Email  string `db:"user_email"`
	Name   string `db:"user_name"`
	Active bool   `db:"user_active"`
}

type Permission struct {
	ID      int  `db:"permission_id"`
	User    int  `db:"permission_user"`
	Account int  `db:"permission_account"`
	Active  bool `db:"permission_active"`
	// Integer
	Integer PermissionInteger `db:"permission_integer"`
}

type Dashboard struct {
	User       User
	Account    Account
	Permission Permission
	// Advanced
	ReadAdvanced  bool
}

func Init(connDvr, connStr string) (*sqlx.DB, error) {
	if connDvr == "" {
		connDvr = "sqlite3"
	}

	if connStr == "" {
		connStr = "./db.db"
	}

	db, err := sqlx.Open(connDvr, connStr)
	if err != nil {
		return nil, logger.Errorf("error opening connection: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, logger.Errorf("error pinging database: %v", err)
	}

	return db, nil
}

func UserEmailByUserID(db *sqlx.DB, userID int) (string, error) {
	var email string

	if err := db.Get(&email, `SELECT user_email FROM users WHERE user_id = ?`, userID); err != nil {
		return "", logger.Errorf("error selecting user_email from users: %v", err)
	}

	return email + "@ucr.ac.cr", nil
}

func UserIDByUserEmail(db *sqlx.DB, email string) (int, error) {
	var userID int

	if at := strings.Index(email, "@"); at != -1 {
		email = email[:at]
	}

	if err := db.Get(&userID, `SELECT user_id FROM users WHERE user_email = ?`, email); err != nil {
		return 0, logger.Errorf("error selecting user_id from users: %v", err)
	}

	return userID, nil
}

func AllowedAccountsByUserID(db *sqlx.DB, userID int) ([]Account, error) {
	const query = `
		SELECT account_id, account_abbr, account_name, account_active
		FROM allowed_accounts
		WHERE user_id = ?
	`

	var accounts []Account
	err := db.Select(&accounts, query, userID)
	if err != nil {
		return nil, logger.Errorf("error querying allowed_accounts for user_id %d: %v", userID, err)
	}

	return accounts, nil
}

func AccountByAccountID(db *sqlx.DB, accountID int) (Account, error) {
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

func UserByUserID(db *sqlx.DB, userID int) (User, error) {
	const queryUserByID = `
		SELECT user_id, user_email, user_name, user_active
		FROM users
		WHERE user_id = ?
	`

	var user User

	if err := db.Get(&user, queryUserByID, userID); err != nil {
		return User{}, logger.Errorf("error selecting account from accounts: %v", err)
	}

	return user, nil
}

func PermissionByUserIDAndAccountID(db *sqlx.DB, userID int, accountID int) (Permission, error) {
	const queryPermissionByUserIDAndAccountID = `
		SELECT permission_id, permission_user, permission_account, permission_integer, permission_active
		FROM permissions
		WHERE permission_user = ? AND permission_account = ?
	`

	var perm Permission

	if err := db.Get(&perm, queryPermissionByUserIDAndAccountID, userID, accountID); err != nil {
		return Permission{}, logger.Errorf("error selecting account from accounts: %v", err)
	}

	return perm, nil
}

func (p Permission) Has(required PermissionInteger) bool {
	return p.Integer&required != 0
}

func (d *Dashboard) LoadData(db *sqlx.DB, userID int, accountID int) error {
	const requiredPermission = Read

	var err error = nil
	var account Account
	var user User
	var perm Permission

	account, err = AccountByAccountID(db, accountID)
	if err != nil {
		return logger.Errorf("error loading account: %v", err)
	}

	if account.Active == false {
		return logger.Errorf("error loading account: inactive account")
	}

	user, err = UserByUserID(db, userID)
	if err != nil {
		return logger.Errorf("error loading user: %v", err)
	}

	if user.Active == false {
		return logger.Errorf("error loading user: inactive user")
	}

	perm, err = PermissionByUserIDAndAccountID(db, userID, accountID)
	if err != nil {
		return logger.Errorf("error loading permission: %v", err)
	}

	if perm.Active == false {
		return logger.Errorf("error loading permission: inactive permission")
	}

	if !perm.Has(requiredPermission) {
		return logger.Errorf("permission error: required:%d got:%d", requiredPermission, perm.Integer)
	}

	d.Account = account
	d.User = user
	d.Permission = perm

	d.ReadAdvanced = perm.Has(ReadAdvanced)

	return nil
}
