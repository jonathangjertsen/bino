-- name: GetActivePatients :many
SELECT
  p.id,
  p.name,
  p.curr_home_id,
  p.status,
  p.journal_url,
  p.time_checkin,
  p.time_checkout,
  COALESCE(sl.name, '???') AS species
FROM patient AS p
LEFT JOIN species_language AS sl
    ON sl.species_id = p.species_id
WHERE curr_home_id IS NOT NULL
  AND language_id = $1
ORDER BY p.curr_home_id, p.sort_order, p.id
;

-- name: GetPatientsByJournalURL :many
SELECT
  p.id
FROM patient AS p
WHERE p.journal_url LIKE CONCAT('%', @lookup::TEXT, '%')
;

-- name: GetFormerPatients :many
SELECT
  p.id,
  p.name,
  p.curr_home_id,
  p.status,
  p.journal_url,
  p.time_checkin,
  p.time_checkout,
  COALESCE(sl.name, '???') AS species
FROM patient AS p
LEFT JOIN species_language AS sl
  ON sl.species_id = p.species_id
WHERE curr_home_id IS NULL
  AND sl.language_id = $1
ORDER BY p.id DESC
;

-- name: AddPatient :one
INSERT INTO patient (species_id, name, curr_home_id, status, time_checkin)
VALUES ($1, $2, $3, $4, NOW())
RETURNING id
;

-- name: AddPatients :many
INSERT INTO patient (species_id, name, curr_home_id, status, journal_url, time_checkin)
SELECT UNNEST(@species::int[]),
       UNNEST(@name::text[]),
       UNNEST(@curr_home_id::int[]),
       UNNEST(@status::int[]),
       UNNEST(@journal_url::text[]),
       NOW()
RETURNING id
;

-- name: MovePatient :exec
UPDATE patient
SET curr_home_id = $2
WHERE id = $1
;

-- name: GetPatient :one
SELECT * FROM patient
WHERE id = $1
;

-- name: GetPatientWithSpecies :one
SELECT p.*, sl.name AS species_name FROM patient AS p
JOIN species_language AS sl
  ON sl.species_id = p.species_id
WHERE p.id = $1
  AND sl.language_id = $2
;

-- name: GetCurrentPatientsForHome :many
SELECT p.*, sl.name AS species_name FROM patient AS p
JOIN species_language AS sl
  ON sl.species_id = p.species_id
WHERE p.curr_home_id = $1
  AND sl.language_id = $2
ORDER BY p.sort_order, p.id
;

-- name: SetPatientStatus :exec
UPDATE patient
SET status = $2
WHERE id = $1
;

-- name: SetPatientName :exec
UPDATE patient
SET name = $2
WHERE id = $1
;

-- name: SetPatientJournal :execresult
UPDATE patient
SET journal_url = $2
WHERE id = $1
;

-- name: UpdatePatientSortOrder :exec
UPDATE patient as p
SET sort_order = v.sort_order
FROM (
  SELECT UNNEST(@ids::int[]) AS id,
         UNNEST(@orders::int[]) AS sort_order
) AS v
WHERE p.id = v.id
;

-- name: CheckoutPatient :exec
UPDATE patient
SET time_checkout = $2
WHERE id = $1
;
