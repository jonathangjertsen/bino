-- +migrate Up
ALTER TABLE appuser ADD COLUMN avatar_url TEXT NULL;

-- +migrate Down
