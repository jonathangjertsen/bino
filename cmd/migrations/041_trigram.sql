-- +migrate Up
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX search_header_trgm ON search
USING gin (lower(header) gin_trgm_ops);

CREATE INDEX search_body_trgm ON search
USING gin (lower(body) gin_trgm_ops);
