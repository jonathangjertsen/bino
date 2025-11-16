-- +migrate Up
ALTER TABLE invitation 
    ADD COLUMN created TIMESTAMPTZ NOT NULL DEFAULT NOW()
;

-- +migrate Down
