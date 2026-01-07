package store

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", "../../database/test.sqlite")
	if err != nil {
		t.Fatalf("db: open %v", err)
	}

	if err := db.Ping(); err != nil {
		t.Fatalf("db: ping %v", err)
	}

	err = Migrate(db, "../../migrations/")
	if err != nil {
		t.Fatalf("db: migration %v", err)
	}

	_, err = db.Exec(`
		DELETE FROM feed_items;
		DELETE FROM feeds;
	`)
	if err != nil {
		t.Fatalf("db: truncate %v", err)
	}

	return db
}

func TestCreateFeed(t *testing.T) {
	db := setupTestDB(t)
	defer func() { _ = db.Close() }()

	store := NewSqlite3FeedStore(db)

	tests := []struct {
		name    string
		feed    *Feed
		wantErr bool
	}{
		{
			name: "valid feed",
			feed: &Feed{
				Title:       "Feed Title",
				Description: "Feed Description",
				Link:        "https://example.com/rss.xml",
			},
			wantErr: false,
		},
		{
			name: "missing title",
			feed: &Feed{
				Description: "Feed Description",
				Link:        "https://example.com/rss.xml",
			},
			wantErr: true,
		},
		{
			name: "missing link",
			feed: &Feed{
				Title:       "Feed Title",
				Description: "Feed Description",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdFeed, err := store.CreateFeed(tt.feed)

			if tt.wantErr {
				assert.Nil(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.feed.Title, createdFeed.Title)
			assert.Equal(t, tt.feed.Description, createdFeed.Description)
			assert.Equal(t, tt.feed.Link, createdFeed.Link)

			retrieved, err := store.GetFeedByID(int64(createdFeed.ID))
			require.NoError(t, err)
			assert.Equal(t, createdFeed.Title, retrieved.Title)
			assert.Equal(t, createdFeed.Description, retrieved.Description)
			assert.Equal(t, createdFeed.Link, retrieved.Link)
		})
	}
}
