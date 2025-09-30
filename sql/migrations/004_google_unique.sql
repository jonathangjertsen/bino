-- +migrate Up
ALTER TABLE appuser ADD CONSTRAINT appuser_google_sub_key UNIQUE (google_sub);

-- +migrate Down

