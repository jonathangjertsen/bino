-- +migrate Up
ALTER TABLE invitation
    ADD COLUMN home INT
;
