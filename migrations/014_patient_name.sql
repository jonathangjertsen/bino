-- +migrate Up
ALTER TABLE patient
    ADD COLUMN name TEXT NOT NULL;
;

-- +migrate Down
ALTER TABLE patient DROP COLUMN name;
