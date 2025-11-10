-- name: UpsertSearchEntry :exec
INSERT INTO search (ns, associated_id, updated, header, body, lang)
VALUES (
    @namespace,
    @associated_id,
    @updated,
    @header,
    @body,
    @lang
)
ON CONFLICT (ns, associated_id) DO UPDATE SET
    updated = EXCLUDED.updated,
    header  = EXCLUDED.header,
    body    = EXCLUDED.body,
    lang    = EXCLUDED.lang
;

-- name: GetSearchUpdatedTime :one
SELECT updated
FROM search
WHERE ns = @namespace
  AND associated_id = @associated_id
;

-- name: Search :many
WITH q AS (
  SELECT websearch_to_tsquery(@lang, @query) AS qry
)
SELECT
	(ts_rank(fts_header, q.qry) + ts_rank(fts_body, q.qry))::DOUBLE PRECISION AS rank,
	s.header,
	ts_headline('norwegian', s.header, q.qry, 'StartSel=[START],StopSel=[STOP],HighlightAll=true')::TEXT AS header_headline,
	ts_headline('norwegian', s.body, q.qry, 'StartSel=[START],StopSel=[STOP],MaxFragments=5,MinWords=3,MaxWords=10')::TEXT AS body_headline,
	s.associated_id,
	s.updated
FROM search AS s, q
WHERE (
	   q.qry @@s.fts_body
	OR q.qry @@s.fts_header
)
ORDER BY RANK desc
LIMIT sqlc.narg('limit')
OFFSET sqlc.narg('offset')
;
