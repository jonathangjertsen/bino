-- +migrate Up
ALTER TABLE patient_event ADD COLUMN event_id INT NOT NULL;

-- +migrate Down
