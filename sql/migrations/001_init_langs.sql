-- +migrate Up
INSERT INTO language (id, short_name, self_name) VALUES
    (1, 'nb', 'Norsk bokmål'),
    (2, 'en', 'English');

-- +migrate Down
TRUNCATE TABLE language;
