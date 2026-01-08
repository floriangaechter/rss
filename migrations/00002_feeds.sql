-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS feeds (
  id INTEGER PRIMARY KEY,
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT,
  link TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  modified_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TRIGGER feeds_modified_at
AFTER UPDATE ON feeds
FOR EACH ROW
BEGIN
  UPDATE feeds
  SET modified_at = datetime('now')
  WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS feeds_modified_at;
DROP TABLE IF EXISTS feeds;
-- +goose StatementEnd
