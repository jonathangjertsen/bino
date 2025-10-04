.PHONY = backend

backend: $(shell find . -name "*.go") go.mod
	sqlc generate
	go tool templ generate
	go generate ./...
	go build -o backend github.com/jonathangjertsen/bino/cmd/backend
	./backend

init_db:
	./scripts/init_db.sh

init_tables:
	psql -U bino -d bino -h localhost -f sql/migrations/000_init.sql


psql:
	@PGPASSWORD=${BINO_DB_PASSWORD} psql -U bino -d bino -h localhost - 

session_key:
	openssl rand -base64 32 > secret/session_key
