-- +migrate Up
CREATE TABLE home_preferred_species (
    home_id INT,
    species_id INT,
    sort_order INT,
    PRIMARY KEY(home_id, species_id)
);

-- +migrate Down
