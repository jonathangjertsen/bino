-- +migrate Up
ALTER TABLE patient ADD COLUMN sort_order INT NOT NULL DEFAULT 2147483647;

-- +migrate Down
