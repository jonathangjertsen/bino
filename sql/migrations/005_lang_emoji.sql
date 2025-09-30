-- +migrate Up
UPDATE language SET short_name='🇳🇴' WHERE short_name='nb';
UPDATE language SET short_name='🇬🇧' WHERE short_name='en';

-- +migrate Down
UPDATE language SET short_name='nb' WHERE short_name='🇳🇴';
UPDATE language SET short_name='en' WHERE short_name='🇬🇧';
