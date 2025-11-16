-- +migrate Up
CREATE OR REPLACE FUNCTION search_match_advanced(
    s            search,
    tsq          tsquery,
    query        text,
    simthreshold real,
    min_created  timestamptz,
    max_created  timestamptz,
    min_updated  timestamptz,
    max_updated  timestamptz
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
    )
    AND (NOT s.skipped)
    AND (min_created IS NULL OR s.created >= min_created)
    AND (max_created IS NULL OR s.created <= max_created)
    AND (min_updated IS NULL OR s.updated >= min_updated)
    AND (max_updated IS NULL OR s.updated <= max_updated)
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
