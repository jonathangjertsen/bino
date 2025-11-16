-- +migrate Up
ALTER TABLE status DROP COLUMN short_name;

-- +migrate Down
ALTER TABLE status ADD COLUMN short_name TEXT NULL;
