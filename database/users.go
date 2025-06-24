package database

import (
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"github.com/tavo-wasd-gh/webtoolkit/logger"
)

type User struct {
	ID     int    `db:"user_id"`
	Email  string `db:"user_email"`
	Name   string `db:"user_name"`
	Active bool   `db:"user_active"`
}

func (u *User) load(db *sqlx.DB, userID int) error {
	var err error = nil
	var user User

	user, err = userByUserID(db, userID)
	if err != nil {
		return logger.Errorf("error loading user: %v", err)
	}

	*u = user
	return nil
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

func userByUserID(db *sqlx.DB, userID int) (User, error) {
	const queryUserByID = `
		SELECT user_id, user_email, user_name, user_active
		FROM users
		WHERE user_id = ?
	`

	var user User

	if err := db.Get(&user, queryUserByID, userID); err != nil {
		return User{}, logger.Errorf("error selecting user from users: %v", err)
	}

	return user, nil
}

func queryAllUsers(db *sqlx.DB) ([]User, error) {
	const queryActiveUsers = "SELECT user_id, user_email, user_name, user_active FROM active_users"

	var users []User
	if err := db.Select(&users, queryActiveUsers); err != nil {
		return nil, logger.Errorf("error selecting account from accounts: %v", err)
	}

	return users, nil
}

func AddUser(db *sqlx.DB, userID, accountID int, newUserEmail, newUserName string) error {
	query := `
		INSERT INTO users (user_email, user_name, user_active)
		VALUES (:user_email, :user_name, :user_active)
	`

	if at := strings.Index(newUserEmail, "@"); at != -1 {
		newUserEmail = newUserEmail[:at]
	}

	user := User{
		Email: newUserEmail,
		Name: newUserName,
		Active: true,
	}

	tx, err := db.Beginx()
	if err != nil {
		return logger.Errorf("failed to begin transaction: %v", err)
	}

	_, err = tx.NamedExec(query, user)
	if err != nil {
		tx.Rollback()
		return logger.Errorf("insert failed, rollback: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return logger.Errorf("commit failed: %v", err)
	}

	return nil
}
