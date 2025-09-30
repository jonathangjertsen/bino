-- +migrate Up
CREATE TABLE tag (
    id         SERIAL PRIMARY KEY
);
COMMENT ON TABLE tag IS 'Each row represents a possible tag for a patient';

CREATE TABLE tag_language (
    tag_id      INTEGER NOT NULL,
    language_id INTEGER NOT NULL,
    name        TEXT NOT NULL,
    PRIMARY KEY (tag_id, language_id),
    FOREIGN KEY (tag_id) REFERENCES tag(id),
    FOREIGN KEY (language_id) REFERENCES language(id)
);
COMMENT ON TABLE tag_language IS 'Internationalization for tag';
COMMENT ON COLUMN tag_language.name IS 'The name of the tag in the given language';

CREATE TABLE patient_tag (
    patient_id INTEGER NOT NULL,
    tag_id     INTEGER NOT NULL,
    PRIMARY KEY (patient_id, tag_id),
    FOREIGN KEY (patient_id) REFERENCES patient(id),
    FOREIGN KEY (tag_id) REFERENCES tag(id)
);
COMMENT ON TABLE patient_tag IS 'Each row represents a tag on a patient';

-- +migrate Down
DROP TABLE tag;
DROP TABLE tag_language;
DROP TABLE patient_tag;
