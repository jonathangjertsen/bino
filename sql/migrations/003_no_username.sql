-- +migrate Up
ALTER TABLE appuser DROP COLUMN username;

-- +migrate Down
ALTER TABLE appuser ADD COLUMN username TEXT NOT NULL;
