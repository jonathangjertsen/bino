-- +migrate Up
ALTER TABLE patient_event ADD COLUMN associated_id INT NULL;

-- +migrate Down
