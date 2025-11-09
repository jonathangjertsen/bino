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

-- name: GetTagsWithLanguage :many
SELECT tag_id, name FROM tag_language
WHERE language_id = $1
ORDER BY (tag_id)
;

-- name: GetTagName :one
SELECT name FROM tag_language
WHERE language_id = $1
  AND tag_id = $2
;

-- name: GetTagsWithLanguageCheckin :many
SELECT tag_id, name, default_show FROM tag_language
INNER JOIN tag AS t
    ON t.id = tag_language.tag_id
WHERE language_id = $1
ORDER BY (default_show, tag_id) DESC
;

-- name: GetTagsForActivePatients :many
SELECT pt.patient_id, pt.tag_id, COALESCE(tl.name, '???') AS name from patient_tag AS pt
LEFT JOIN tag_language AS tl
    ON tl.tag_id = pt.tag_id
LEFT JOIN patient AS p
    ON p.id = pt.patient_id
WHERE p.curr_home_id IS NOT NULL
AND tl.language_id = $1
ORDER BY p.curr_home_id, p.sort_order, p.id
;

-- name: GetTagsForFormerPatients :many
SELECT pt.patient_id, pt.tag_id, COALESCE(tl.name, '???') AS name from patient_tag AS pt
LEFT JOIN tag_language AS tl
    ON tl.tag_id = pt.tag_id
LEFT JOIN patient AS p
    ON p.id = pt.patient_id
WHERE p.curr_home_id IS NULL
AND tl.language_id = $1
;

-- name: GetTagsForPatient :many
SELECT pt.tag_id, COALESCE(tl.name, '???') AS name from patient_tag AS pt
LEFT JOIN tag_language AS tl
    ON tl.tag_id = pt.tag_id
WHERE pt.patient_id = $1
  AND tl.language_id = $2
;

-- name: AddPatientTags :exec
INSERT INTO patient_tag (patient_id, tag_id)
VALUES ($1, unnest(@tags::INT[]))
ON CONFLICT (patient_id, tag_id) DO NOTHING
;

-- name: DeletePatientTag :exec
DELETE FROM patient_tag
WHERE patient_id = $1
  AND tag_id = $2
;

-- name: GetTagsForCurrentPatientsForHome :many
SELECT pt.patient_id, pt.tag_id, COALESCE(tl.name, '???') AS name
FROM patient_tag AS pt
LEFT JOIN tag_language as tl
  ON tl.tag_id = pt.tag_id
LEFT JOIN patient as p
  ON p.id = pt.patient_id
WHERE p.curr_home_id = $1
  AND tl.language_id = $2
;
