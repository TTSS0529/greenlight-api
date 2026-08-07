-include .envrc

# ==================================================================================== #
# HELPERS
# ==================================================================================== #

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [y/N] ' && read ans && [ $${ans:-N} = y ]

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	@go run ./cmd/api -db-dsn=${GREENLIGHT_DB_DSN}

## db/psql: connect to the database using psql
.PHONY: db/psql
db/psql:
	@psql ${GREENLIGHT_DB_DSN}

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}...'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	@migrate -path ./migrations -database ${GREENLIGHT_DB_DSN} up

# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #

## audit: tidy and vendor dependencies and format, vet and test all code
.PHONY: audit
audit: vendor
	@echo 'Formatting code...'
	go fmt ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests...'
	go test -race -vet=off ./...

## vendor: tidy and vendor dependencies
.PHONY: vendor
vendor:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies...'
	go mod vendor

## test: run integration tests
.PHONY: test
test:
	docker compose -f docker-compose.test.yml up --abort-on-container-exit --exit-code-from test

## benchmark: run benchmark tests
.PHONY: benchmark
benchmark:
	go test -bench=. -benchmem ./...

# ==================================================================================== #
# BUILD
# ==================================================================================== #

current_time = $(shell date --iso-8601=seconds)
git_description = $(shell git describe --always --dirty --tags --long)
linker_flags = '-s -X main.buildTime=${current_time} -X main.version=${git_description}'

## build/api: build the cmd/api application
.PHONY: build/api
build/api:
	@echo 'Building cmd/api...'
	go build -ldflags=${linker_flags} -o=./bin/api ./cmd/api
	GOOS=linux GOARCH=amd64 go build -ldflags=${linker_flags} -o=./bin/linux_amd64/api ./cmd/api

# ==================================================================================== #
# DEPLOY
# ==================================================================================== #

## docker/deploy: build docker images, run database migrations and start services
.PHONY: docker/build
docker/build:
	docker compose build \
		--build-arg BUILD_TIME="${current_time}" \
		--build-arg VERSION="${git_description}"

## docker/up: build and start all services
.PHONY: docker/up
docker/up:
	docker compose up -d db
	docker compose run --rm migration
	docker compose run --rm seed
	docker compose up -d api caddy

## docker/deploy: build and start all services
.PHONY: docker/deploy
docker/deploy: docker/build docker/up

## docker/down: stop all services
.PHONY: docker/down
docker/down:
	docker compose down

## docker/down/v: stop services and remove volumes
.PHONY: docker/down/v
docker/down/v:
	docker compose down -v

## docker/redeploy: rebuild and redeploy
.PHONY: docker/redeploy
docker/redeploy: docker/down docker/build docker/up

## docker/reset: rebuild everything from scratch
.PHONY: docker/reset
docker/reset: docker/down/v docker/build docker/up

## docker/logs: print log infos
.PHONY: docker/logs
docker/logs:
	docker compose logs -f