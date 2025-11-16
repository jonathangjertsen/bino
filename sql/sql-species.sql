-- name: GetSpecies :many
SELECT * FROM species
ORDER BY id;

-- name: GetSpeciesLanguage :many
SELECT * FROM species_language
ORDER BY (species_id, language_id);

-- name: GetNameOfSpecies :one
SELECT name FROM species_language
WHERE species_id = $1
  AND language_id = $2
;

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

-- name: GetSpeciesByName :many
SELECT species_id
FROM species_language
WHERE name = $1
;
