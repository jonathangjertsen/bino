.PHONY = backend

BUILD_KEY := $(shell tr -dc A-Za-z0-9 </dev/urandom | head -c 8)

backend: $(shell find . -name "*.go") go.mod
	sqlc generate
	go tool templ generate
	go generate ./...
	sass styles.scss static/gen.css -q
	go build -ldflags="-X 'main.BuildKey=${BUILD_KEY}'" -o backend github.com/jonathangjertsen/bino
	./backend

init_db:
	./scripts/init_db.sh

init_tables:
	psql -U bino -d bino -h localhost -f sql/migrations/000_init.sql


psql:
	@PGPASSWORD=${BINO_DB_PASSWORD} psql -U bino -d bino -h localhost - 

session_key:
	openssl rand -base64 32 > secret/session_key
