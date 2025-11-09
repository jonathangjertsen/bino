-- name: AddPatientEvent :one
INSERT INTO patient_event (patient_id, home_id, event_id, associated_id, note, appuser_id, time)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id
;

-- name: GetEventsForPatient :many
SELECT
    pe.*,
    h.name AS home_name,
    au.display_name AS user_name,
    au.avatar_url AS avatar_url
FROM patient_event AS pe
JOIN home AS h
  ON h.id = pe.home_id
JOIN appuser AS au
  ON au.id = pe.appuser_id
WHERE pe.patient_id = $1
ORDER BY pe.time
;

-- name: GetFirstEventOfTypeForPatient :one
SELECT
  pe.time
FROM patient_event AS pe
WHERE pe.patient_id = $1
  AND pe.event_id = $2
ORDER BY pe.time ASC
;

-- name: SetEventNote :exec
UPDATE patient_event
SET note = $2
WHERE id = $1
;

-- name: DeleteEventsCreatedByUser :exec
DELETE
FROM patient_event
WHERE appuser_id = $1
;
