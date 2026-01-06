-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS feed_items (
  id INTEGER PRIMARY KEY,
  feed_id INTEGER NOT NULL REFERENCES feeds(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT,
  link TEXT NOT NULL,
  published_at TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  modified_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TRIGGER feed_items_modified_at
AFTER UPDATE ON feed_items
FOR EACH ROW
BEGIN
  UPDATE feed_items
  SET modified_at = datetime('now')
  WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS feed_items_modified_at;
DROP TABLE IF EXISTS feed_items;
-- +goose StatementEnd
