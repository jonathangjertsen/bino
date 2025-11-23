-- name: RegisterFile :one
INSERT
INTO file
  (uuid, accessibility, creator, created, filename, mimetype, size)
VALUES 
  (@uuid, @accessibility, @creator, @created, @filename, @mimetype, @size)
RETURNING id
;

-- name: DeregisterFile :exec
DELETE
FROM file
WHERE id = @id
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
