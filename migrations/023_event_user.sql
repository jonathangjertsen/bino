-- +migrate Up
ALTER TABLE patient_event
    ADD COLUMN appuser_id INT NOT NULL
    DEFAULT 1
;
ALTER TABLE patient_event
    ADD CONSTRAINT patient_event_appuser_id_fkey
    FOREIGN KEY (appuser_id)
    REFERENCES appuser (id)
;

-- +migrate Down
