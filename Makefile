include .env

run: build
	@./bin/uwe

build:
	@go build -o bin/uwe ./cmd/api/.

seed:
	@go run cmd/seed/main.go

db-status:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_URL) goose -dir="db/migrations" status

up:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_URL) goose -dir="db/migrations" up

reset:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DB_URL) goose -dir="db/migrations" reset

test:
	@go test -v ./cmd/api