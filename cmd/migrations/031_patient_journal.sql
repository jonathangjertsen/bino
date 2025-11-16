-- +migrate Up
ALTER TABLE patient
    ADD COLUMN journal_url TEXT NULL;

-- +migrate Down
