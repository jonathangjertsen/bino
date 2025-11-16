-- name: GetHomes :many
SELECT * FROM home
ORDER BY name
;

-- name: InsertHome :exec
INSERT INTO home (name)
VALUES ($1)
;

-- name: UpdateHomeName :exec
UPDATE home
SET name = $2
WHERE id = $1
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
SELECT h.* FROM home_appuser AS ha
INNER JOIN home AS h
  ON h.id = ha.home_id
WHERE appuser_id = $1
;

-- name: GetHome :one
SELECT * FROM home
WHERE id = $1
;

-- name: SetHomeCapacity :exec
UPDATE home
SET capacity = $2
WHERE id = $1
;

-- name: SetHomeNote :exec
UPDATE home
SET note = $2
WHERE id = $1
;

-- name: GetPreferredSpecies :many
SELECT hps.home_id, sl.species_id, sl.name FROM home_preferred_species AS hps
JOIN species_language AS sl USING (species_id)
WHERE language_id = $1
;

-- name: GetPreferredSpeciesForHome :many
SELECT hps.species_id, sl.name FROM home_preferred_species AS hps
JOIN species_language AS sl USING(species_id)
WHERE hps.home_id = $1
  AND sl.language_id = $2
ORDER BY hps.sort_order ASC, hps.species_id ASC
;

-- name: AddPreferredSpecies :exec
INSERT INTO home_preferred_species (
  home_id,
  species_id
) VALUES (
  $1,
  $2
)
;

-- name: DeletePreferredSpecies :exec
DELETE FROM home_preferred_species
WHERE home_id    = $1
  AND species_id = $2
;

-- name: UpdatePreferredSpeciesSortOrder :exec
UPDATE home_preferred_species as hps
SET sort_order = v.sort_order
FROM (
  SELECT UNNEST(@species_id::int[]) AS id,
         UNNEST(@orders::int[]) AS sort_order
) AS v
WHERE hps.species_id = v.id
  AND hps.home_id = @home_id
;

-- name: AddHomeUnavailablePeriod :one
INSERT INTO home_unavailable (home_id, from_date, to_date, note)
VALUES ($1, $2, $3, $4)
RETURNING id
;

-- name: UpdateHomeUnavailableFrom :exec
UPDATE home_unavailable
SET from_date = $2
WHERE id = $1
;

-- name: UpdateHomeUnavailableTo :exec
UPDATE home_unavailable
SET to_date = $2
WHERE id = $1
;

-- name: UpdateHomeUnavailableNote :exec
UPDATE home_unavailable
SET note = $2
WHERE id = $1
;

-- name: DeleteHomeUnavailablePeriod :exec
DELETE
FROM home_unavailable
WHERE id = $1
;

-- name: GetHomeUnavailablePeriods :many
SELECT *
FROM home_unavailable
WHERE home_id = $1
  AND to_date + INTERVAL '1 DAY' >= NOW()
ORDER BY to_date
;

-- name: GetAllUnavailablePeriods :many
SELECT *
FROM home_unavailable
WHERE to_date + INTERVAL '1 DAY' >= NOW()
ORDER BY home_id, to_date
;

-- name: GetUnavailablePeriodsInRange :many
SELECT hu.*, h.name
FROM home_unavailable AS hu
INNER JOIN home AS h
  ON hu.home_id = h.id
WHERE from_date <= @range_end
  AND to_date >= @range_begin
;

-- name: GetHomeByName :many
SELECT *
FROM home
WHERE name = $1
;
