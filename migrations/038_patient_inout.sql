-- +migrate Up
ALTER TABLE patient ADD COLUMN time_checkin  TIMESTAMPTZ NULL;
ALTER TABLE patient ADD COLUMN time_checkout TIMESTAMPTZ NULL;

-- +migrate Down
