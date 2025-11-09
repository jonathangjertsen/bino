-- +migrate Up
ALTER TABLE appuser ADD COLUMN access_level INT NOT NULL DEFAULT 0;

-- +migrate Down
