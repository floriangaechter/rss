package store

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
)

func Open(logger *log.Logger) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "database/rss.sqlite")
	if err != nil {
		return nil, fmt.Errorf("db: open %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("db: ping %w", err)
	}

	logger.Printf("db: connection ok")

	return db, err
}

func MigrateFS(db *sql.DB, migrationFS fs.FS, dir string) error {
	goose.SetBaseFS(migrationFS)
	defer func() {
		goose.SetBaseFS(nil)
	}()
	return Migrate(db, dir)
}

func Migrate(db *sql.DB, dir string) error {
	err := goose.SetDialect("sqlite3")
	if err != nil {
		return fmt.Errorf("db: migrate %w", err)
	}

	err = goose.Up(db, dir)
	if err != nil {
		return fmt.Errorf("db: goose up %w", err)
	}
	return nil
}
