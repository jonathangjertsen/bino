-- name: GetSession :one
SELECT *
FROM session
WHERE id = $1
;

-- name: InsertSession :exec
INSERT INTO session (id, appuser_id, expires, last_seen)
VALUES ($1, $2, $3, NOW())
;

-- name: UpdateSessionLastSeen :exec
UPDATE session
SET last_seen = NOW()
WHERE id = $1
;

-- name: RevokeSession :exec
DELETE FROM session
WHERE id = $1
;

-- name: RevokeAllSessionsForUser :execresult
DELETE FROM session
WHERE appuser_id = $1
;

-- name: RevokeAllOtherSessionsForUser :execresult
DELETE FROM session
WHERE id != $1
  AND appuser_id == $2
;

-- name: DeleteStaleSessions :execresult
DELETE FROM session
WHERE expires < NOW()
  -- Try to make sure users log in at the start of a session rather than the middle of it 
  OR (
    (expires < NOW() + INTERVAL '20 minutes')
    AND last_seen < NOW() - INTERVAL '5 minutes'
  )
;
