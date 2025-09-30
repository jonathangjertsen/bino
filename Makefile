PHONY = init_db init_tables backend all sqlc psql

all: sqlc backend

init_db:
	./scripts/init_db.sh

init_tables:
	psql -U bino -d bino -h localhost -f sql/migrations/000_init.sql

backend: $(shell find . -name "*.go") go.mod
	go build -o backend github.com/jonathangjertsen/bino/cmd/backend

run-backend: backend
	./backend

sqlc:
	sqlc generate

psql:
	@PGPASSWORD=${BINO_DB_PASSWORD} psql -U bino -d bino -h localhost - 

session_key:
	openssl rand -base64 32 > secret/session_key

css:
	npx @tailwindcss/cli -i ./static/input.css -o ./static/styles.css
	cp node_modules/flyonui/flyonui.js ./static/flyonui.js

templ:
	go tool templ generate

