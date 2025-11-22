-- +migrate Up
CREATE TABLE image (
    id      SERIAL PRIMARY KEY,
    uuid    TEXT UNIQUE NOT NULL,
    creator INT NOT NULL,
    created TIMESTAMPTZ NOT NULL
);
