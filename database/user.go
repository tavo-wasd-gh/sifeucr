package database

import (
	"time"

	"github.com/tavo-wasd-gh/webtoolkit/logger"
	"github.com/jmoiron/sqlx"
)

type User struct {
	ID       string     `db:"id"`
	Name     string     `db:"name"`
	Created  *time.Time `db:"created"`
	Disabled *time.Time `db:"disabled"`
	// Runtime
	AvailableAccounts []Account
	Account           Account
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
