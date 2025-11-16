-- +migrate Up
ALTER TABLE species DROP COLUMN short_name;

-- +migrate Down
ALTER TABLE species ADD COLUMN short_name TEXT NULL;
