start:
	sqlc init

generate:
	sqlc generate

init:
	docker run -it --rm --network host --volume "/Users/george/workspace/bhe/db:/db" migrate/migrate:v4.17.0 create -ext sql -dir /db/migrations init_schema

up:
	docker run -it --rm --network host --volume ./db:/db migrate/migrate:v4.17.0 -path=/db/migrations -database "postgresql://user:rocketman1@localhost:5432/bhe?sslmode=disable" up

down:
	docker run -it --rm --network host --volume ./db:/db migrate/migrate:v4.17.0 -path=/db/migrations -database "postgresql://user:rocketman1@localhost:5432/bhe?sslmode=disable" down

# specially for CI/CD
up_ci:
	docker run --rm --network host --volume ./db:/db migrate/migrate:v4.17.0 -path=/db/migrations -database "postgresql://user:rocketman1@localhost:5432/bhe?sslmode=disable" up

test:
	go test -v -cover ./...

.PHONY: start generate init up down test up_ci
