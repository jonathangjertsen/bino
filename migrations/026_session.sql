-- +migrate Up
CREATE TABLE session (
    id          TEXT PRIMARY KEY,
    appuser_id  INT NOT NULL,
    expires     TIMESTAMPTZ NOT NULL,
    last_seen   TIMESTAMPTZ NOT NULL
);

-- +migrate Down
