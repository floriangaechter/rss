package store

import (
	"database/sql"
)

type FeedItem struct {
	ID          int    `json:"id"`
	FeedID      int    `json:"feedID"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	PublishedAt string `json:"publishedAt"`
	ReadAt      string `json:"readAt"`
}

type Sqlite3FeedItemStore struct {
	db *sql.DB
}

func NewSqlite3FeedItemStore(db *sql.DB) *Sqlite3FeedItemStore {
	return &Sqlite3FeedItemStore{db: db}
}

type FeedItemStore interface {
	CreateFeedItem(*FeedItem) (*FeedItem, error)
	GetFeedItemByID(id int64) (*FeedItem, error)
	UpdateFeedItem(*FeedItem) error
}

func (sqlite3 *Sqlite3FeedItemStore) CreateFeedItem(feedItem *FeedItem) (*FeedItem, error) {
	tx, err := sqlite3.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		INSERT INTO feed_items (
			feed_id,
			title,
			description,
			link,
			published_at
		)
		VALUES (
			?,
			?,
			?,
			?,
			?
		)
		RETURNING id;
	`
	err = tx.QueryRow(query, feedItem.FeedID, feedItem.Title, feedItem.Description, feedItem.Link, feedItem.PublishedAt).Scan(&feedItem.ID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return feedItem, nil
}

func (sqlite3 *Sqlite3FeedItemStore) GetFeedItemByID(id int64) (*FeedItem, error) {
	feedItem := &FeedItem{}

	query := `
		SELECT
			id,
			title,
			description,
			link,
			published_at
		FROM
			feed_items
		WHERE
			id = ?
	`
	err := sqlite3.db.QueryRow(query, id).Scan(&feedItem.ID, &feedItem.Title, &feedItem.Description, &feedItem.Link, &feedItem.PublishedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return feedItem, nil
}

func (sqlite3 *Sqlite3FeedItemStore) UpdateFeedItem(feedItem *FeedItem) error {
	tx, err := sqlite3.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	query := `
		UPDATE
			feed_items
		SET
			read_at = ?
		WHERE id = ?
	`
	result, err := tx.Exec(query, feedItem.ReadAt)
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
