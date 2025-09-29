-- name: GetSpeciesName :many
SELECT s.id, sl.name FROM species AS s
LEFT JOIN species_language AS sl
ON sl.species_id = s.id
HAVING sl.language_id = $1
;

