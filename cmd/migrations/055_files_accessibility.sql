-- +migrate Up
-- 0 is private to creator
-- 1 is shown to logged-in users only
-- 2 is public
ALTER TABLE image RENAME TO file;
ALTER TABLE file ADD COLUMN accessibility INT NOT NULL DEFAULT 1;
