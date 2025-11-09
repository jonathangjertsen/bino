-- +migrate Up
CREATE TABLE home_unavailable (
    id SERIAL PRIMARY KEY,
    home_id INT NOT NULL,
    from_date DATE NULL,
    to_date DATE NULL,
    note TEXT NULL
);

-- +migrate Down
