-- +migrate Up
CREATE TABLE language (
    id         SERIAL PRIMARY KEY,
    short_name TEXT NOT NULL,
    self_name  TEXT NOT NULL     
);
COMMENT ON TABLE language IS 'Each row is a supported language';
COMMENT ON COLUMN language.short_name IS 'A short name used to refer to this language in code';
COMMENT ON COLUMN language.self_name IS 'The name of this language, in that language';

CREATE TABLE species (
    id              SERIAL PRIMARY KEY,
    short_name      TEXT NOT NULL,
    scientific_name TEXT NOT NULL
);
COMMENT ON TABLE species IS 'Each row is a species';
COMMENT ON COLUMN species.short_name      IS 'A short name used to refer to this species in code';
COMMENT ON COLUMN species.scientific_name IS 'The scientific name for this species';

CREATE TABLE species_language (
    species_id  INTEGER NOT NULL,
    language_id INTEGER NOT NULL,
    name        TEXT NOT NULL,
    PRIMARY KEY (species_id, language_id),
    FOREIGN KEY (species_id) REFERENCES species(id),
    FOREIGN KEY (language_id) REFERENCES language(id)
);
COMMENT ON TABLE species_language IS 'Internationalization for species';
COMMENT ON COLUMN species_language.name IS 'The name of the species in the given language';

CREATE TABLE appuser (
    id           SERIAL PRIMARY KEY,
    username     TEXT NOT NULL,
    display_name TEXT NOT NULL
);
COMMENT ON TABLE appuser                 IS 'Each row is an app user (person)';
COMMENT ON COLUMN appuser.username       IS 'The user''s username, for login';
COMMENT ON COLUMN appuser.display_name   IS 'The user''s display name';

CREATE TABLE home (
    id       SERIAL PRIMARY KEY,
    name     TEXT NOT NULL
);
COMMENT ON TABLE home IS 'Each row is a rehab home';

CREATE TABLE home_appuser (
    appuser_id INTEGER NOT NULL,
    home_id INTEGER NOT NULL,
    FOREIGN KEY (appuser_id) REFERENCES appuser(id),
    FOREIGN KEY (home_id) REFERENCES home(id)
);
COMMENT ON TABLE home_appuser IS 'Each row associates an app user with a rehab home';

CREATE TABLE status (
    id         SERIAL PRIMARY KEY,
    short_name TEXT NOT NULL
);
COMMENT ON TABLE status             IS 'Each row represents a possible status for a patient';
COMMENT ON COLUMN status.short_name IS 'A short name that may be used to refer to this status in code'; 

CREATE TABLE status_language (
    status_id   INTEGER NOT NULL,
    language_id INTEGER NOT NULL,
    name        TEXT NOT NULL,
    PRIMARY KEY (status_id, language_id),
    FOREIGN KEY (status_id) REFERENCES status(id),
    FOREIGN KEY (language_id) REFERENCES language(id)
);
COMMENT ON TABLE status_language IS 'Internationalization for status';
COMMENT ON COLUMN status_language.name IS 'The name of the status in the given language';

CREATE TABLE patient (
    id             SERIAL PRIMARY KEY,
    species_id     INTEGER NOT NULL,
    curr_status_id INTEGER NOT NULL,
    curr_home_id   INTEGER NULL,
    FOREIGN KEY (species_id) REFERENCES species(id),
    FOREIGN KEY (curr_status_id) REFERENCES status(id),
    FOREIGN KEY (curr_home_id) REFERENCES home(id)
);
COMMENT ON TABLE patient IS 'Each row represents a patient';

CREATE TABLE event (
    id         SERIAL PRIMARY KEY,
    short_name TEXT NOT NULL
);
COMMENT ON TABLE event IS 'Each row represents an event that may occur to a patient';

CREATE TABLE event_language (
    event_id    INTEGER NOT NULL,
    language_id INTEGER NOT NULL,
    name        TEXT NOT NULL,
    PRIMARY KEY (event_id, language_id),
    FOREIGN KEY (event_id) REFERENCES event(id),
    FOREIGN KEY (language_id) REFERENCES language(id)
);
COMMENT ON TABLE event_language IS 'Internationalization for event';
COMMENT ON COLUMN event_language.name IS 'The name of the event in the given language';

CREATE TABLE patient_event (
    id         SERIAL PRIMARY KEY,
    event_id   INTEGER NOT NULL,
    patient_id INTEGER NOT NULL,
    home_id    INTEGER NOT NULL,
    note       TEXT NOT NULL,
    FOREIGN KEY (event_id) REFERENCES event(id),
    FOREIGN KEY (patient_id) REFERENCES patient(id),
    FOREIGN KEY (home_id) REFERENCES home(id)
);
COMMENT ON TABLE patient_event IS 'Each row represents an event that has occurred to a specific patient';

-- +migrate Down
DROP SCHEMA IF EXISTS bino CASCADE;
