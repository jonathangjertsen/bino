-- +migrate Up
CREATE TABLE search(
    -- Namespace (e.g. journal, patient, home)
    ns TEXT NOT NULL,
    -- Associated ID (e.g. patient ID)
    associated_id INT,
    -- When it was updated
    updated TIMESTAMPTZ NOT NULL,
    -- Header plaintext
    header text,
    -- Body plaintext
    body text,
    -- Language
    lang regconfig,
    -- full text search vector for header
    fts_header tsvector generated always as (setweight(to_tsvector(lang, COALESCE(header, '')), 'A')) stored,
    -- full text search vector for body
    fts_body tsvector generated always as (setweight(to_tsvector(lang, COALESCE(body, '')), 'C')) stored,
    PRIMARY KEY (ns, associated_id)
);
CREATE INDEX search_fts_header ON search USING gin(fts_header);
CREATE INDEX search_fts_body ON search USING gin(fts_body);

-- +migrate Down
