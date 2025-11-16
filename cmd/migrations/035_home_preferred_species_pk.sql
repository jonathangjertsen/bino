-- +migrate Up
ALTER TABLE home_preferred_species
    ADD PRIMARY KEY (home_id, species_id)
;

-- +migrate Down
