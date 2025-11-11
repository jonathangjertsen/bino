-- name: UpsertSearchEntry :exec
INSERT INTO search (ns, associated_url, updated, header, body, lang)
VALUES (
    @namespace,
    @associated_url,
    @updated,
    @header,
    @body,
    @lang
)
ON CONFLICT (ns, associated_url) DO UPDATE SET
    updated        = EXCLUDED.updated,
    header         = EXCLUDED.header,
    body           = EXCLUDED.body,
    lang           = EXCLUDED.lang,
    associated_url = EXCLUDED.associated_url
;

-- name: GetSearchUpdatedTime :one
SELECT updated
FROM search
WHERE ns = @namespace
  AND associated_url = @associated_url
;

-- name: SearchBasic :many
WITH q AS (
  SELECT websearch_to_tsquery(sqlc.arg('lang')::regconfig, sqlc.arg('query')) AS qry
)
SELECT
  i.*,
  (
      i.r_fts_header
    + i.r_fts_body
  )::real AS rank
FROM (
  SELECT
    (sqlc.arg('w_fts_header')::real   * ts_rank(s.fts_header, q.qry))::real AS r_fts_header,
    (sqlc.arg('w_fts_body')::real     * ts_rank(s.fts_body,   q.qry))::real AS r_fts_body,
    ts_headline(sqlc.arg('lang')::regconfig, s.header, q.qry, 'StartSel=[START],StopSel=[STOP],HighlightAll=true')::text AS header_headline,
    ts_headline(sqlc.arg('lang')::regconfig, s.body,   q.qry, 'StartSel=[START],StopSel=[STOP],MaxFragments=5,MinWords=3,MaxWords=10,FragmentDelimiter=[CUT]')::text AS body_headline,
    s.ns,
    s.associated_url,
    s.updated
  FROM search s
  CROSS JOIN q
  WHERE (
        q.qry @@ s.fts_header
     OR q.qry @@ s.fts_body
  )
) i
ORDER BY rank DESC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset')
;

-- name: SearchAdvanced :many
WITH q AS (
  SELECT websearch_to_tsquery(sqlc.arg('lang')::regconfig, sqlc.arg('query')) AS qry
)
SELECT
  i.*,
  (
      i.r_fts_header
    + i.r_fts_body
    + i.r_sim_header
    + i.r_sim_body
    + i.r_ilike_header
    + i.r_ilike_body
    + i.r_recency
  )::real AS rank
FROM (
  SELECT
    (sqlc.arg('w_fts_header')::real   * ts_rank(s.fts_header, q.qry))::real AS r_fts_header,
    (sqlc.arg('w_fts_body')::real     * ts_rank(s.fts_body,   q.qry))::real AS r_fts_body,
    (sqlc.arg('w_sim_header')::real   * f.sim_header)::real                 AS r_sim_header,
    (sqlc.arg('w_sim_body')::real     * f.sim_body)::real                   AS r_sim_body,
    (sqlc.arg('w_ilike_header')::real * f.ilike_header)::real               AS r_ilike_header,
    (sqlc.arg('w_ilike_body')::real   * f.ilike_body)::real                 AS r_ilike_body,
    (sqlc.arg('w_recency')::real      * f.recency)::real                    AS r_recency,
    COALESCE(s.header, '') AS header,
    COALESCE(s.body, '') AS body,
    ts_headline(sqlc.arg('lang')::regconfig, s.header, q.qry, 'StartSel=[START],StopSel=[STOP],HighlightAll=true')::text AS header_headline,
    ts_headline(sqlc.arg('lang')::regconfig, s.body,   q.qry, 'StartSel=[START],StopSel=[STOP],MaxFragments=5,MinWords=3,MaxWords=10,FragmentDelimiter=[CUT]')::text AS body_headline,
    s.ns,
    s.associated_url,
    s.updated
  FROM search s
  CROSS JOIN q
  CROSS JOIN LATERAL (
    SELECT
      similarity(lower(s.header), lower(sqlc.arg('query'))) AS sim_header,
      similarity(lower(s.body),   lower(sqlc.arg('query'))) AS sim_body,
      CASE WHEN s.header ILIKE ('%' || sqlc.arg('query') || '%') THEN 1 ELSE 0 END AS ilike_header,
      CASE WHEN s.body   ILIKE ('%' || sqlc.arg('query') || '%') THEN 1 ELSE 0 END AS ilike_body,
      exp(
        - GREATEST(0, EXTRACT(EPOCH FROM (now() - s.updated))) /
          (sqlc.arg('recency_half_life_days')::real * 86400.0)
      ) AS recency
  ) f
  WHERE (
        q.qry @@ s.fts_header
     OR q.qry @@ s.fts_body
     OR f.ilike_header = 1
     OR f.ilike_body = 1
     OR f.sim_header > sqlc.arg('simthreshold')::real
     OR f.sim_body   > sqlc.arg('simthreshold')::real
  )
) i
ORDER BY rank DESC
LIMIT sqlc.arg('limit')
OFFSET sqlc.arg('offset')
;
