-- name: GetLanguages :many
SELECT * FROM language
ORDER BY id ASC
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

-- name: GetSpecies :many
SELECT * FROM species
ORDER BY id;

-- name: GetSpeciesLanguage :many
SELECT * FROM species_language
ORDER BY (species_id, language_id);

-- name: AddSpecies :one
INSERT INTO species (scientific_name)
VALUES ($1)
RETURNING id
;

-- name: UpsertSpeciesLanguage :exec
INSERT INTO species_language (species_id, language_id, name)
VALUES ($1, $2, $3)
ON CONFLICT (species_id, language_id) DO UPDATE
    SET name = EXCLUDED.name
;

-- name: GetSpeciesWithLanguage :many
SELECT species_id, name FROM species_language
WHERE language_id = $1
ORDER BY (species_id)
;