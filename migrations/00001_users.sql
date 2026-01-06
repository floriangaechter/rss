-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY,
  username TEXT UNIQUE NOT NULL,
  email TEXT NOT NULL,
  password TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  modified_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TRIGGER users_modified_at
AFTER UPDATE ON users
FOR EACH ROW
BEGIN
  UPDATE users
  SET modified_at = datetime('now')
  WHERE id = OLD.id;
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS users_modified_at;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
