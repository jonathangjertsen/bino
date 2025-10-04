-- +migrate Up
ALTER TABLE home_appuser
    ADD PRIMARY KEY (appuser_id, home_id)
;

-- +migrate Down
