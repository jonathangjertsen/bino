-- +migrate Up
ALTER TABLE patient DROP COLUMN curr_status_id;
DROP TABLE status_language;
DROP TABLE status;

-- +migrate Down
