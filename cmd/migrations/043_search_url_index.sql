-- +migrate Up
ALTER TABLE search DROP CONSTRAINT search_pkey;
ALTER TABLE search ADD PRIMARY KEY (ns, associated_url);
