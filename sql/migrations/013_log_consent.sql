-- +migrate Up
ALTER TABLE appuser
    ADD COLUMN logging_consent TIMESTAMPTZ NULL
;

-- +migrate Down
ALTER TABLE appuser DROP COLUMN logging_consent;
