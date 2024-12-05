include .env

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## run: builds the binary and runs the program
.PHONY: run
run: build
	@./bin/uwe

## build: builds the binary
.PHONY: build
build:
	@go build -o bin/uwe ./cmd/api/.

## seed: Seeds the database
.PHONY: seed
seed:
	@go run cmd/seed/main.go

## db-stats: Checks the status of the database to see if you are connected
.PHONY: db-status
db-status:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir="db/migrations" status

## up: Runs the up migrations
.PHONY: up
up:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir="db/migrations" up

## reset: Resets the migrations
.PHONY: reset
reset:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir="db/migrations" reset

## test-all: Runs all unit tests sequentially
.PHONY: test-all
test-all:
	@go test -v -p=1 -count=1 ./...

## test-api: Runs all API endpoint tests against a mock database
.PHONY: test-api
test-api:
	@go test -v ./cmd/api/...

## test-internal: Runs all internal business logic such as SQL queries against a test database
.PHONY: test-internal
test-internal:
	@go test -v ./internal/...

## test-e2e: Spins up a real version of the application and runs tests against a test database
.PHONY: test-e2e
test-e2e:
	@go test -v ./e2e/...
