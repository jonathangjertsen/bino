-- +migrate Up
ALTER TABLE appuser
    ADD COLUMN has_gdrive_access BOOLEAN NOT NULL DEFAULT FALSE;
;

-- +migrate Down
