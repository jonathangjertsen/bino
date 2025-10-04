-- +migrate Up
ALTER TABLE patient_event DROP COLUMN event_id;
DROP TABLE event_language;
DROP TABLE event;

-- +migrate Down
