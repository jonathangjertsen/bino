-- name: GetLanguages :many
SELECT * FROM language;

-- name: GetSpeciesName :many
SELECT s.id, sl.name FROM species AS s
LEFT JOIN species_language AS sl
ON sl.species_id = s.id
HAVING sl.language_id = $1
;

-- name: UpsertUser :one
INSERT INTO appuser (display_name, google_sub, email)
VALUES ($1, $2, $3)
ON CONFLICT (google_sub) DO UPDATE
    SET display_name = EXCLUDED.display_name,
        email        = EXCLUDED.email
RETURNING id;

-- name: GetUser :one
SELECT au.*, COALESCE(al.language_id, 1) FROM appuser AS au
LEFT JOIN appuser_language AS al
ON au.id = al.appuser_id
WHERE id = $1
;

-- name: SetUserLanguage :exec
INSERT INTO appuser_language (appuser_id, language_id)
VALUES ($1, $2)
ON CONFLICT (appuser_id) DO UPDATE
    SET language_id = EXCLUDED.language_id
;
