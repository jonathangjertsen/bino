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
SELECT * FROM appuser
WHERE id = $1;
