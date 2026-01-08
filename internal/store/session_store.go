package store

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"time"
)

type Session struct {
	Token     string
	UserID    int
	ExpiresAt time.Time
}

type SessionStore interface {
	CreateSession(userID int, expiresIn time.Duration) (*Session, error)
	GetSession(token string) (*Session, error)
	DeleteSession(token string) error
	DeleteUserSessions(userID int) error
}

type Sqlite3SessionStore struct {
	db *sql.DB
}

func NewSqlite3SessionStore(db *sql.DB) *Sqlite3SessionStore {
	return &Sqlite3SessionStore{db: db}
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *Sqlite3SessionStore) CreateSession(userID int, expiresIn time.Duration) (*Session, error) {
	token, err := generateToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(expiresIn)
	query := `
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES (?, ?, ?)
	`
	_, err = s.db.Exec(query, token, userID, expiresAt.Format(time.RFC3339))
	if err != nil {
		return nil, err
	}

	return &Session{
		Token:     token,
		UserID:    userID,
		ExpiresAt: expiresAt,
	}, nil
}

func (s *Sqlite3SessionStore) GetSession(token string) (*Session, error) {
	session := &Session{}
	var expiresAtStr string

	query := `
		SELECT
			token,
			user_id,
			expires_at
		FROM
			sessions
		WHERE
			token = ?
		AND
			expires_at > datetime('now')
	`
	err := s.db.QueryRow(query, token).Scan(&session.Token, &session.UserID, &expiresAtStr)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	expiresAt, err := time.Parse(time.RFC3339, expiresAtStr)
	if err != nil {
		return nil, err
	}
	session.ExpiresAt = expiresAt

	return session, nil
}

func (s *Sqlite3SessionStore) DeleteSession(token string) error {
	query := `
		DELETE FROM
			sessions
		WHERE
			token = ?
	`
	_, err := s.db.Exec(query, token)
	return err
}

func (s *Sqlite3SessionStore) DeleteUserSessions(userID int) error {
	query := `
		DELETE FROM
			sessions
		WHERE
			user_id = ?
	`
	_, err := s.db.Exec(query, userID)
	return err
}
