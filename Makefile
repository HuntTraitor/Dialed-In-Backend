include .env

## help: print this help message
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## run: builds the binary and runs the program
run: build
	@./bin/uwe

## build: builds the binary
build:
	@go build -o bin/uwe ./cmd/api/.

## seed: Seeds the database
seed:
	@go run cmd/seed/main.go

## db-stats: Checks the status of the database to see if you are connected
db-status:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir="db/migrations" status

## up: Runs the up migrations
up:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir="db/migrations" up

## reset: Resets the migrations
reset:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir="db/migrations" reset

## test-all: Runs all unit tests sequentially
test-all:
	@go test -v -p=1 -count=1 ./...

## test-api: Runs all API endpoint tests against a mock database
test-api:
	@go test -v ./cmd/api/...

## test-internal: Runs all internal business logic such as SQL queries against a test database
test-internal:
	@go test -v ./internal/...

## test-e2e: Spins up a real version of the application and runs tests against a test database
test-e2e:
	@go test -v ./e2e/...
