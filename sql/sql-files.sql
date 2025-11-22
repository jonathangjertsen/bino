-- name: RegisterFiles :one
WITH ins AS (
  INSERT INTO file (uuid, accessibility, creator, created)
  SELECT u, @accessibility, @creator, @created
  FROM unnest(@uuids::text[]) AS u
  RETURNING id
)
SELECT array_agg(id)::int[] AS ids
FROM ins
;

-- name: GetFilesForUser :many
SELECT *
FROM file
WHERE
      creator = @creator
  AND accessibility >= @accessibility
ORDER BY created DESC
;

-- name: GetFileByID :one
SELECT *
FROM file
WHERE id = @id
;
