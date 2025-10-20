-- +migrate Up
ALTER TABLE patient
    DROP CONSTRAINT patient_curr_home_id_fkey;
ALTER TABLE patient_event
    DROP CONSTRAINT patient_event_home_id_fkey;

-- +migrate Down
