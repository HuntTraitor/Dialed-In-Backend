include .env

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run: builds the binary and runs the program
.PHONY: run
run: build
	@./bin/uwe

## build: builds the binary
.PHONY: build
build:
	@echo 'Building cmd/api...'
	go build -ldflags='-s' -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux/_amd64/api ./cmd/api

## seed: seeds the database
.PHONY: seed
seed:
	@go run cmd/seed/main.go

## db-stats: checks the status of the database to see if you are connected
.PHONY: db-status
db-status:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir="db/migrations" status

## up: runs the up migrations
.PHONY: up
up:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir="db/migrations" up

## reset: resets the migrations
.PHONY: reset
reset:
	@GOOSE_DRIVER=postgres GOOSE_DBSTRING=$(DATABASE_URL) goose -dir="db/migrations" reset

## test-all: runs all unit tests sequentially
.PHONY: test-all
test-all:
	@go test -v -p=1 -count=1 ./...

## test-api: runs all API endpoint tests against a mock database
.PHONY: test-api
test-api:
	@go test -v ./cmd/api/...

## test-internal: runs all internal business logic such as SQL queries against a test database
.PHONY: test-internal
test-internal:
	@go test -v ./internal/...

## test-e2e: spins up a real version of the application and runs tests against a test database
.PHONY: test-e2e
test-e2e:
	@go test -v ./e2e/...

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## tidy: format all .go files
.PHONY: tidy
tidy:
	@echo 'Formatting .go files...'
	go fmt ./...
	@echo 'Tidying module dependencies...'
	go mod tidy

## audit: run quality control checks
audit:
	@echo 'Checking module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Verifying code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	make test-all

