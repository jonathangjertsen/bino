-- +migrate Up
ALTER TABLE species_language DROP CONSTRAINT species_language_language_id_fkey;
ALTER TABLE appuser_language DROP CONSTRAINT appuser_language_language_id_fkey;
ALTER TABLE tag_language DROP CONSTRAINT tag_language_language_id_fkey;
DROP TABLE language;

-- +migrate Down
