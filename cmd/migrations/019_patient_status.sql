-- +migrate Up
ALTER TABLE patient ADD COLUMN status INT NOT NULL DEFAULT 0;

-- +migrate Down
