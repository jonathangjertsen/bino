-- +migrate Up
ALTER TABLE appuser
  ADD COLUMN IF NOT EXISTS google_sub TEXT NOT NULL,
  ADD COLUMN IF NOT EXISTS email TEXT NOT NULL;

-- +migrate Down
ALTER TABLE appuser DROP COLUMN google_sub, email;
