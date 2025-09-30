-- +migrate Up
ALTER TABLE event DROP COLUMN short_name;

-- +migrate Down
ALTER TABLE event ADD COLUMN short_name TEXT NULL;
