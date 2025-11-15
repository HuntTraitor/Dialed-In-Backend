-include .env

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
	@./bin/api

## build: builds the binary
.PHONY: build
build:
	@echo 'Building cmd/api...'
	GIT_COMMIT=$(shell git rev-parse HEAD)
	GIT_DIRTY=$(shell test -n "`git status --porcelain`" && echo "+CHANGES" || true)
	BUILD_DATE=$(shell TZ=UTC0 git show --quiet --date='format-local:%Y-%m-%dT%T%z' --format="%cd")
	go build -ldflags="-X github.com/hunttraitor/dialed-in-backend/internal/vcs.revision=${GIT_COMMIT}${GIT_DIRTY} -X github.com/hunttraitor/dialed-in-backend/internal/vcs.time=${BUILD_DATE}" -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux/_amd64/api ./cmd/api

## seed: seeds the database
.PHONY: seed
seed:
	@docker compose exec -T db psql -U ${POSTGRES_USER} -d ${POSTGRES_DB} -f /tmp/seeds.sql

## db-stats: checks the status of the database to see if you are connected
.PHONY: db-status
db-status:
	@docker compose exec -T app goose -dir=db/migrations postgres "$(MIGRATION_URL)" status

## up: runs the up migrations
.PHONY: up
up:
	@docker compose exec -T app goose -dir=db/migrations postgres "$(MIGRATION_URL)" up

## reset: resets the migrations
.PHONY: reset
reset:
	@docker compose exec -T app goose -dir=db/migrations postgres "$(MIGRATION_URL)" reset

## test-all: runs all unit tests sequentially
.PHONY: test-all
test-all:
	@docker compose exec -T app go test -v -p=1 -count=1 -buildvcs=false ./...

## test-api: runs all API endpoint tests against a mock database
.PHONY: test-api
test-api:
	@docker compose exec -T app go test -v -buildvcs=false ./cmd/api/...

## test-internal: runs all internal business logic such as SQL queries against a test database
.PHONY: test-internal
test-internal:
	@docker compose exec -T app go test -v -buildvcs=false ./internal/...

## test-e2e: spins up a real version of the application and runs tests against a test database
.PHONY: test-e2e
test-e2e:
	@docker compose exec -T app go test -v -buildvcs=false ./e2e/...

## docker-up: runs the docker container
.PHONY: docker-up
docker-up:
	docker compose up --build

## docker-down: shuts down the docker container
.PHONY: docker-down
docker-down:
	docker compose down

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

# ==================================================================================== #
# PRODUCTION
# ==================================================================================== #

production_host_ip = $(PROD_HOST)

## prod/deploy: deploy to production
.PHONY: prod/deploy
prod/deploy:
	ssh -o StrictHostKeyChecking=no \
	    -o ServerAliveInterval=30 \
	    -o ServerAliveCountMax=10 \
	    $(PROD_USER)@$(production_host_ip) '\
		cd /root/app/Dialed-In-Backend && \
		git pull && \
		docker compose -f production-compose.yml pull app && \
		docker compose -f production-compose.yml up -d app \
	'

## prod/connect: connect to the production server
.PHONY: prod/connect
prod/connect:
	ssh root@${production_host_ip}


