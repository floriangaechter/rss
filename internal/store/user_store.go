package store

import (
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type password struct {
	plainText *string
	hash      []byte
}

func (p *password) Set(plainTextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plainTextPassword), 12)
	if err != nil {
		return err
	}

	p.plainText = &plainTextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plainTextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plainTextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

type User struct {
	ID       int      `json:"id"`
	Username string   `json:"username"`
	Password password `json:"-"`
}

type Sqlite3UserStore struct {
	db *sql.DB
}

func NewSqlite3UserStore(db *sql.DB) *Sqlite3UserStore {
	return &Sqlite3UserStore{db: db}
}

type UserStore interface {
	CreateUser(*User) error
	GetUserByUsername(username string) (*User, error)
	UpdateUser(*User) error
}

func (s *Sqlite3UserStore) CreateUser(user *User) error {
	query := `
		INSERT INTO
			users (username, password)
		VALUES
			(?, ?) RETURNING id
	`
	err := s.db.QueryRow(query, user.Username, user.Password.hash).Scan(&user.ID)
	if err != nil {
		return err
	}

	return nil
}

func (s *Sqlite3UserStore) GetUserByUsername(username string) (*User, error) {
	user := &User{
		Password: password{},
	}

	query := `
		SELECT
			id,
			username,
			password
		FROM
			users
		WHERE
			username = ?
	`

	err := s.db.QueryRow(query, username).Scan(&user.ID, &user.Username)
	if err != sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Sqlite3UserStore) UpdateUser(user *User) error {
	query := `
		UPDATE users
		SET
			username = ?,
			password = ?
		WHERE
			id = ?
	`

	result, err := s.db.Exec(query, user.Username, user.Password.hash)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
