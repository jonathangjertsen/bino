-- +migrate Up
CREATE TABLE appuser_language (
    appuser_id  INTEGER PRIMARY KEY,
    language_id INTEGER NOT NULL,
    FOREIGN KEY (appuser_id) REFERENCES appuser(id),
    FOREIGN KEY (language_id) REFERENCES language(id)
);
COMMENT ON TABLE appuser_language IS 'Language preference for each user';

-- +migrate Down
DROP TABLE appuser_language;
