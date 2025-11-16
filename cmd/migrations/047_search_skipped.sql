-- +migrate Up
ALTER TABLE search ADD COLUMN skipped BOOLEAN;

CREATE OR REPLACE FUNCTION search_match_advanced(
    s            search,
    tsq          tsquery,
    query        text,
    simthreshold real
)
RETURNS boolean
LANGUAGE sql
STABLE
AS $sma$
    SELECT (
           (tsq @@ s.fts_header)
        OR (tsq @@ s.fts_body)
        OR (s.header ILIKE ('%' || query || '%'))
        OR (s.body   ILIKE ('%' || query || '%'))
        OR (similarity(lower(s.header), lower(query)) > simthreshold)
        OR (similarity(lower(s.body),   lower(query)) > simthreshold)
    ) AND NOT s.skipped
$sma$;

CREATE OR REPLACE FUNCTION search_match_basic(
    s   search,
    tsq tsquery
)
RETURNS boolean
LANGUAGE sql
STABLE
AS $smb$
    SELECT (
           (tsq @@ s.fts_header)
        OR (tsq @@ s.fts_body)
    ) AND NOT s.skipped
$smb$;
