-- +migrate Up
ALTER TABLE tag
    ADD COLUMN default_show BOOLEAN NOT NULL DEFAULT FALSE;
COMMENT ON COLUMN tag.default_show IS 'Whether to show this tag by default when creating a patient';

-- +migrate Down
ALTER TABLE tag DROP COLUMN default_show;
