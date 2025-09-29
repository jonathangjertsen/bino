.PHONY = init_db init_tables

init_db:
	./scripts/init_db.sh

init_tables:
	psql -U bino -d bino -h localhost -f schema.sql
