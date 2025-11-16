-- +migrate Up
CREATE TABLE invitation (
    id      TEXT PRIMARY KEY,
    email   TEXT,
    expires TIMESTAMPTZ NOT NULL
);

-- +migrate Down
