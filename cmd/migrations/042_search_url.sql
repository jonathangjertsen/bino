-- +migrate Up
ALTER TABLE search ADD COLUMN associated_url TEXT;
