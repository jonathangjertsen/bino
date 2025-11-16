-- +migrate Up
ALTER TABLE search ADD COLUMN extra_data TEXT;
