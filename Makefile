include .env

start:
	sqlc init

generate:
	sqlc generate

init:
	docker run -it --rm --network host --volume "$(WD)" migrate/migrate:v4.17.0 create -ext sql -dir /db/migrations init_schema

up:
	docker run -it --rm --network host --volume ./db:/db migrate/migrate:v4.17.0 -path=/db/migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" up

down:
	docker run -it --rm --network host --volume ./db:/db migrate/migrate:v4.17.0 -path=/db/migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable" down

.PHONY: generate init up down
