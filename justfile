default: build

sqlc:
	sqlc generate

templ:
	go tool templ generate

enum:
	go generate ./...

sass:
	sass styles.scss cmd/static/gen.css -q

gen: sqlc templ enum sass

build: gen
	go build -ldflags="-X 'main.BuildKey=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 8)'" -o backend github.com/jonathangjertsen/bino/cmd

run: build
	./backend

init_db:
	./scripts/init_db.sh

init_tables:
	psql -U bino -d bino -h localhost -f sql/migrations/000_init.sql

psql:
	@PGPASSWORD=${BINO_DB_PASSWORD} psql -U bino -d bino -h localhost - 

session_key:
	openssl rand -base64 32 > secret/session_key

dbuild:
	docker build -t bino .

drun: dbuild
	docker run --rm -p 8080:8080 \
		-v $(pwd)/config.json:/main/config.json \
		-v $(pwd)/secret:/main/secret \
		bino