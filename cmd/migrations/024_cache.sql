-- +migrate Up
CREATE TABLE cache (
    key   TEXT PRIMARY KEY,
    value TEXT
);
COMMENT ON TABLE cache IS 'Each row represents a cache entry';

-- +migrate Down
