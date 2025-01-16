package auth

import (
	"fmt"
	"net/http"
	"database/sql"

	"github.com/tavo-wasd-gh/gosmtp"
)

type User struct {
	email string
	passw string
	asked string
}

func (u *User) Validate(w http.ResponseWriter, r *http.Request, db *sql.DB) error {
	return nil
}

func availableAccounts(db *sql.DB, email string) ([]string, error) {
	query := `
		SELECT id_cuenta
		FROM cuentas
		WHERE presidencia = $1 OR tesoreria = $1
	`

	rows, err := db.Query(query, email)
	if err != nil {
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var cuenta string
		if err := rows.Scan(&cuenta); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}
		results = append(results, cuenta)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	if len(results) == 0 {
		return nil, fmt.Errorf("no matching accounts found for email: %s", email)
	}

	return results, nil
}

func validateUser(db *sql.DB, email, passwd string, asked ...string) (string, error) {
	accounts, err := availableAccounts(db, email)
	if err != nil {
		return "", err
	}

	s := smtp.Client("smtp.ucr.ac.cr", "587", passwd)
	if err := s.Validate(email); err != nil {
		return err
	}

	if len(asked) > 0 && asked[0] != "" {
		for _, account := range accounts {
			if account == asked[0] {
				return asked[0], nil
			}
		}
	} else if len(accounts) == 1 {
		return accounts[0], nil
	}

	return "", fmt.Errorf("error validating user: multiple accounts available")
}
