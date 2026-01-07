// Package store
package store

import (
	"database/sql"
)

type Feed struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Items       []Item `json:"items"`
}

type Item struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	PublishedAt string `json:"publishedAt"`
}

type Sqlite3FeedStore struct {
	db *sql.DB
}

func NewSqlite3FeedStore(db *sql.DB) *Sqlite3FeedStore {
	return &Sqlite3FeedStore{db: db}
}

type FeedStore interface {
	CreateFeed(*Feed) (*Feed, error)
	GetFeedByID(id int64) (*Feed, error)
	UpdateFeed(*Feed) error
	DeleteFeedByID(id int64) error
}

func (sqlite3 *Sqlite3FeedStore) CreateFeed(feed *Feed) (*Feed, error) {
	tx, err := sqlite3.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		INSERT INTO feeds (
			title,
			description,
			link
		)
		VALUES (
			?,
			?,
			?
		)
		RETURNING id;
	`
	err = tx.QueryRow(query, feed.Title, feed.Description, feed.Link).Scan(&feed.ID)
	if err != nil {
		return nil, err
	}

	// items should be pulled in periodically and automatically

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return feed, nil
}

func (sqlite3 *Sqlite3FeedStore) GetFeedByID(id int64) (*Feed, error) {
	feed := &Feed{}
	query := `
		SELECT
			id,
			title,
			description,
			link
		FROM
			feeds
		WHERE
			id = ?
	`
	err := sqlite3.db.QueryRow(query, id).Scan(&feed.ID, &feed.Title, &feed.Description, &feed.Link)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	itemQuery := `
		SELECT
			id,
			title,
			description,
			link,
			published_at
		FROM
			feed_items
		WHERE
			feed_id = ?
		ORDER BY published_at
	`
	rows, err := sqlite3.db.Query(itemQuery, id)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	for rows.Next() {
		var item Item
		err = rows.Scan(
			&item.ID,
			&item.Title,
			&item.Description,
			&item.Link,
			&item.PublishedAt,
		)
		if err != nil {
			return nil, err
		}
		feed.Items = append(feed.Items, item)
	}

	return feed, nil
}

func (sqlite3 *Sqlite3FeedStore) UpdateFeed(feed *Feed) error {
	tx, err := sqlite3.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		UPDATE
			feeds
		SET
			title = ?,
			description = ?,
			link = ?
		WHERE id = ?
	`
	result, err := tx.Exec(query, feed.Title, feed.Description, feed.Link, feed.ID)
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

	return tx.Commit()
}

func (sqlite3 *Sqlite3FeedStore) DeleteFeedByID(id int64) error {
	tx, err := sqlite3.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		DELETE FROM
			feeds
		WHERE id = ?
	`

	result, err := sqlite3.db.Exec(query, id)
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

	return tx.Commit()
}
