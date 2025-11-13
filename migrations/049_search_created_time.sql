-- +migrate Up
ALTER TABLE search ADD COLUMN created timestamptz NOT NULL DEFAULT now();
UPDATE search SET created = updated;
