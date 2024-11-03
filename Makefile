DB_URL=postgresql://user:rocketman1@localhost:5432/bhe?sslmode=disable

start:
	sqlc init

generate:
	sqlc generate

init:
	docker run -it --rm --network host --volume "/Users/george/workspace/bhe/db:/db" migrate/migrate:v4.17.0 create -ext sql -dir /db/migrations init_schema

migrateup:
	docker run -it --rm --network host --volume ./db:/db migrate/migrate:v4.17.0 -path=/db/migrations -database "$(DB_URL)" -verbose up

migratedown:
	docker run -it --rm --network host --volume ./db:/db migrate/migrate:v4.17.0 -path=/db/migrations -database "$(DB_URL)" -verbose down

migrateup1:
	docker run -it --rm --network host --volume ./db:/db migrate/migrate:v4.17.0 -path=/db/migrations -database "$(DB_URL)" -verbose up 1

migratedown1:
	docker run -it --rm --network host --volume ./db:/db migrate/migrate:v4.17.0 -path=/db/migrations -database "$(DB_URL)" -verbose down 1

# specially for CI/CD
up_ci:
	docker run --rm --network host --volume ./db:/db migrate/migrate:v4.17.0 -path=/db/migrations -database "$(DB_URL)" -verbose up

test:
	go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/repository.go bhe/db/sqlc Repository

.PHONY: start generate init migrateup migratedown migrateup1 migratedown1 test up_ci server mock
