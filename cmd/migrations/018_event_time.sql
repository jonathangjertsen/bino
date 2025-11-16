-- +migrate Up
ALTER TABLE patient_event ADD COLUMN TIME TIMESTAMPTZ NOT NULL;

-- +migrate Down
