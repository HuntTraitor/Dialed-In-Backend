include .env

run: build
	@./bin/uwe

build:
	@go build -o bin/uwe ./cmd/api/.

seed:
	@go run cmd/seed/main.go

db-status:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir="db/migrations" status

up:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir="db/migrations" up

reset:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir="db/migrations" reset

test-all:
	@go test -v -p=1 -count=1 ./...

test-api:
	@go test -v ./cmd/api/...

test-internal:
	@go test -v ./internal/...

test-e2e:
	@go test -v ./e2e/...
