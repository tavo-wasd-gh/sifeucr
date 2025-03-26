package database

import (
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
)

type User struct {
	ID       string     `db:"id"`
	Name     string     `db:"name"`
	Created  *time.Time `db:"created"`
	Disabled *time.Time `db:"disabled"`
	// Runtime
	Perms             []Permission
	AvailableAccounts []Account
	Account           Account
}

type Permission struct {
	Account string
	Integer int
}

func (u *User) Login(db *sqlx.DB) error {
	var err error

	if u.ID == "" {
		return logger.Errorf("email is required to log in")
	}

	if u.AvailableAccounts, err = accountsByUser(db, u.ID); err != nil {
		return logger.Errorf("error querying available accounts for '%s': %v", u.ID, err)
	}

	if u.Account.ID != "" {
		accountFound := false

		for _, acc := range u.AvailableAccounts {
			if acc.ID == u.Account.ID {
				accountFound = true
				break
			}
		}

		if !accountFound {
			return logger.Errorf("account ID '%s' is not available for user '%s'", u.Account.ID, u.ID)
		}
	} else {
		if len(u.AvailableAccounts) == 1 {
			u.Account.ID = u.AvailableAccounts[0].ID
		} else {
			return nil
		}
	}

	// Account.ID is defined and matches one AvailableAccounts
	// Or, it is not defined but there is only one AvailableAccounts

	if err = db.Get(u, "SELECT * FROM users WHERE id = ?", u.ID); err != nil {
		return logger.Errorf("error querying user by id '%s': %v", u.ID, err)
	}

	if u.Account, err = accountByID(db, u.Account.ID); err != nil {
		return logger.Errorf("error querying account by id '%s': %v", u.Account.ID, err)
	}

	// TODO: fill requests and other structs

	return nil
}

func UserExists(db *sqlx.DB, user string) (bool, error) {
	var exists bool

	err := db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM users WHERE id = ?)`, user)
	if err != nil {
		return false, logger.Errorf("error checking user existence: %v", err)
	}

	return exists, nil
}

func (u *User) FirstSetupUser(db *sqlx.DB) error {
	if err := u.insertUser(db); err != nil {
		return logger.Errorf("failed to create setup user '%s': %v", u.ID, err)
	}
	return nil
}

func (u *User) insertUser(db *sqlx.DB) error {
	exists, err := UserExists(db, u.ID)

	if err != nil {
		return logger.Errorf("error checking if user with id '%s' already exists", u.ID)
	}

	if exists {
		return logger.Errorf("user with id '%s' already exists", u.ID)
	}

	_, err = db.Exec(
		`INSERT INTO users (id, name) VALUES (?, ?)`,
		u.ID, u.Name,
	)
	if err != nil {
		return logger.Errorf("error inserting user: %v", err)
	}

	if err := u.updatePermissions(db); err != nil {
		return logger.Errorf("error setting permissions: %v", err)
	}

	return nil
}

func (u *User) updatePermissions(db *sqlx.DB) error {
	for _, perm := range u.Perms {
		_, err := db.Exec(
			`DELETE FROM permissions WHERE user = ? AND account = ?`,
			u.ID, perm.Account,
		)
		if err != nil {
			return logger.Errorf("error deleting old permission for account '%s': %v", perm.Account, err)
		}

		_, err = db.Exec(
			`INSERT INTO permissions (user, account, permission_integer)
			 VALUES (?, ?, ?)`,
			u.ID, perm.Account, perm.Integer,
		)
		if err != nil {
			return logger.Errorf("error inserting permission for account '%s': %v", perm.Account, err)
		}
	}
	return nil
}
