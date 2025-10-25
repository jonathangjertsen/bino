-- +migrate Up
ALTER TABLE home
    ADD COLUMN capacity INT NOT NULL DEFAULT 0
;

-- +migrate Down
