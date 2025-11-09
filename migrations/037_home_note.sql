-- +migrate Up
ALTER TABLE home
    ADD COLUMN note TEXT NOT NULL DEFAULT ''
;

-- +migrate Down
