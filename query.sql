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

-- name: GetStatuses :many
SELECT * FROM status
ORDER BY id;

-- name: GetStatusesLanguage :many
SELECT * FROM status_language
ORDER BY (status_id, language_id);

-- name: AddStatus :one
INSERT INTO status DEFAULT VALUES
RETURNING id
;
-- name: UpsertStatusLanguage :exec
INSERT INTO status_language (status_id, language_id, name)
VALUES ($1, $2, $3)
ON CONFLICT (status_id, language_id) DO UPDATE
    SET name = EXCLUDED.name
;

-- name: GetStatusWithLanguage :many
SELECT status_id, name FROM status_language
WHERE language_id = $1
ORDER BY (status_id)
;

-- name: GetEvents :many
SELECT * FROM event
ORDER BY id;

-- name: GetEventsLanguage :many
SELECT * FROM event_language
ORDER BY (event_id, language_id);

-- name: AddEvent :one
INSERT INTO event DEFAULT VALUES
RETURNING id
;
-- name: UpsertEventLanguage :exec
INSERT INTO event_language (event_id, language_id, name)
VALUES ($1, $2, $3)
ON CONFLICT (event_id, language_id) DO UPDATE
    SET name = EXCLUDED.name
;

-- name: GetEventWithLanguage :many
SELECT event_id, name FROM event_language
WHERE language_id = $1
ORDER BY (event_id)
;

-- name: GetTags :many
SELECT * FROM tag
ORDER BY id;

-- name: GetTagsLanguage :many
SELECT * FROM tag_language
ORDER BY (tag_id, language_id);

-- name: AddTag :one
INSERT INTO tag (default_show)
    VALUES ($1)
RETURNING id
;

-- name: UpdateTagDefaultShown :exec
UPDATE tag SET default_show = $1
WHERE id = $2
;

-- name: UpsertTagLanguage :exec
INSERT INTO tag_language (tag_id, language_id, name)
VALUES ($1, $2, $3)
ON CONFLICT (tag_id, language_id) DO UPDATE
    SET name = EXCLUDED.name
;

-- name: GetTagWithLanguage :many
SELECT tag_id, name FROM tag_language
WHERE language_id = $1
ORDER BY (tag_id)
;

-- name: GetTagWithLanguageCheckin :many
SELECT tag_id, name FROM tag_language
INNER JOIN tag AS t
    ON t.id = tag_language.tag_id
WHERE language_id = $1
    AND t.default_show
ORDER BY (tag_id)
;
