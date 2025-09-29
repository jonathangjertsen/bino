#!/bin/bash

if ! command -v psql >/dev/null 2>&1; then
  echo "PostgreSQL is not installed or not in PATH."
  exit 1
fi

while true; do
  read -s -p "Password for PostgreSQL user 'bino': " DB_PASS
  echo
  read -s -p "Repeat password: " DB_PASS2
  echo
  [ "$DB_PASS" = "$DB_PASS2" ] && break
  echo "Passwords do not match. Try again."
done

sudo -u postgres psql <<EOF
DROP DATABASE IF EXISTS bino;
DROP SCHEMA IF EXISTS bino CASCADE;
DROP USER IF EXISTS bino;
CREATE DATABASE bino;
CREATE USER bino WITH ENCRYPTED PASSWORD '$DB_PASS';
GRANT ALL PRIVILEGES ON DATABASE bino TO bino;
CREATE SCHEMA bino AUTHORIZATION bino;
GRANT ALL PRIVILEGES ON SCHEMA bino TO bino;
ALTER USER bino SET search_path TO bino;
EOF

echo "Done"
