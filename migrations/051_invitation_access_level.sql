-- +migrate Up
ALTER TABLE invitation
    ADD COLUMN access_level INT NOT NULL DEFAULT 1
;
