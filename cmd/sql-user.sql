-- name: CreateUser :one
INSERT INTO appuser (
  display_name,
  google_sub,
  email, 
  avatar_url,
  access_level
) VALUES (
  $1,
  $2,
  $3,
  $4,
  $5
)
RETURNING id
;

-- name: UpdateUser :exec
UPDATE appuser
SET display_name = @display_name,
    google_sub = @google_sub,
    avatar_url = @avatar_url
WHERE id = $1
  AND email = $2
;

-- name: GetUserIDByEmail :one
SELECT id
FROM appuser
WHERE email = $1
;

-- name: GetUser :one
SELECT au.*, COALESCE(al.language_id, 1)
FROM appuser AS au
LEFT JOIN appuser_language AS al
ON au.id = al.appuser_id
WHERE id = $1
;

-- name: GetAppusers :many
SELECT au.*, ha.home_id FROM appuser AS au
LEFT JOIN home_appuser AS ha
    ON ha.appuser_id = au.id
ORDER BY au.id
;

-- name: SetUserLanguage :exec
INSERT INTO appuser_language (appuser_id, language_id)
VALUES ($1, $2)
ON CONFLICT (appuser_id) DO UPDATE
    SET language_id = EXCLUDED.language_id
;

-- name: SetLoggingConsent :exec
UPDATE appuser SET logging_consent = NOW() + sqlc.arg(period)::INT * INTERVAL '1 day'
WHERE id = $1
;

-- name: RevokeLoggingConsent :exec
UPDATE appuser SET logging_consent = NULL
WHERE id = $1
;


-- name: SetUserGDriveAccess :exec
UPDATE appuser
SET has_gdrive_access = $2
WHERE id = $1
;

-- name: ClearAllUserGDriveAccess :exec
UPDATE appuser
SET has_gdrive_access = FALSE
WHERE 1
;

-- name: GetAppusersForHome :many
SELECT au.*
FROM home_appuser AS hau
INNER JOIN appuser AS au
  ON hau.appuser_id = au.id
WHERE home_id = $1
;

-- name: GetHomesWithDataForUser :many
SELECT h.*
FROM home AS h
INNER JOIN home_appuser AS hau
  ON hau.home_id = h.id
WHERE appuser_id = $1
;

-- name: RemoveHomesForAppuser :exec
DELETE
FROM home_appuser
WHERE appuser_id = $1
;

-- name: DeleteSessionsForUser :exec
DELETE
FROM session
WHERE appuser_id = $1
;

-- name: ScrubAppuser :exec
UPDATE appuser SET
  display_name = 'Deleted user (id = ' || id || ')',
  google_sub = '',
  email = '',
  logging_consent = NULL,
  avatar_url = '',
  has_gdrive_access = FALSE 
WHERE id = $1
;

-- name: DeleteAppuser :exec
DELETE
FROM appuser
WHERE id = $1
;

-- name: DeleteAppuserLanguage :exec
DELETE
FROM appuser_language
WHERE appuser_id = $1
;
