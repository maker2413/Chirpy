createdb:
	docker exec -it postgres17 createdb --username=postgres --owner=postgres chirpy

dropdb:
	docker exec -it postgres17 dropdb chirpy

startpostgres:
	docker run --rm --name postgres17 -p 5432:5432 \
	-e POSTGRES_USER=postgres \
	-e POSTGRES_PASSWORD=postgres \
	-d postgres:17-alpine

stoppostgres:
	docker stop postgres12

migrateup:
	cd sql/schema/ && \
	goose postgres postgres://postgres:postgres@localhost:5432/chirpy up

migratedown:
	cd sql/schema/ && \
	goose postgres postgres://postgres:postgres@localhost:5432/chirpy down
