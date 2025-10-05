-- name: GetLanguages :many
SELECT * FROM language
ORDER BY id ASC
;

-- name: UpsertUser :one
INSERT INTO appuser (display_name, google_sub, email, avatar_url)
VALUES ($1, $2, $3, $4)
ON CONFLICT (google_sub) DO UPDATE
    SET display_name = EXCLUDED.display_name,
        email        = EXCLUDED.email,
        avatar_url   = EXCLUDED.avatar_url
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
SELECT tag_id, name, default_show FROM tag_language
INNER JOIN tag AS t
    ON t.id = tag_language.tag_id
WHERE language_id = $1
ORDER BY (default_show, tag_id) DESC
;

-- name: GetHomes :many
SELECT * FROM home
ORDER BY name
;

-- name: GetActivePatients :many
SELECT p.id, p.name, p.curr_home_id, COALESCE(sl.name, '???') AS species FROM patient AS p
LEFT JOIN species_language AS sl
    ON sl.species_id = p.species_id
WHERE curr_home_id IS NOT NULL
  AND language_id = $1
;

-- name: GetTagsForActivePatients :many
SELECT pt.patient_id, pt.tag_id, COALESCE(tl.name, '???') AS name from patient_tag AS pt
LEFT JOIN tag_language AS tl
    ON tl.tag_id = pt.tag_id
LEFT JOIN patient AS p
    ON p.id = pt.patient_id
WHERE p.curr_home_id IS NOT NULL
AND tl.language_id = $1
;

-- name: GetAppusers :many
SELECT au.*, ha.home_id FROM appuser AS au
LEFT JOIN home_appuser AS ha
    ON ha.appuser_id = au.id
;

-- name: UpsertHome :exec
INSERT INTO home (name)
VALUES ($1)
ON CONFLICT (id) DO UPDATE
    SET name = EXCLUDED.name
;

-- name: AddUserToHome :exec
INSERT INTO home_appuser (home_id, appuser_id)
VALUES ($1, $2)
;

-- name: RemoveUserFromHome :exec
DELETE FROM home_appuser
WHERE home_id = $1
  AND appuser_id = $2
;

-- name: GetHomesForUser :many
SELECT home_id FROM home_appuser
WHERE appuser_id = $1
;


-- name: SetLoggingConsent :exec
UPDATE appuser SET logging_consent = NOW() + sqlc.arg(period)::INT * INTERVAL '1 day'
WHERE id = $1
;

-- name: RevokeLoggingConsent :exec
UPDATE appuser SET logging_consent = NULL
WHERE id = $1
;

-- name: AddPatient :one
INSERT INTO patient (species_id, name, curr_home_id, status)
VALUES ($1, $2, $3, $4)
RETURNING id
;

-- name: AddPatientTags :exec
INSERT INTO patient_tag (patient_id, tag_id)
VALUES ($1, unnest(@tags::INT[]))
ON CONFLICT (patient_id, tag_id) DO NOTHING
;

-- name: AddPatientEvent :one
INSERT INTO patient_event (patient_id, home_id, event_id, note, time)
VALUES ($1, $2, $3, $4, $5)
RETURNING id
;