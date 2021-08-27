# Include variables from the .envrc file
include .envrc

# ======================================================================== #
# HELPERS
# ======================================================================== #

## help: print this help message.
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [Y/N] ' && read ans && [ $${ans:-N} = y ]

# ======================================================================== #
# CODE QUALITY CONTROL
# ======================================================================== #

## audit: perform all code quality checks and module dependency resolution & verification.
.PHONY: audit
audit: format
	@echo '.....Tidying, verifying and resolving module dependencies.....'
	go mod tidy
	go mod verify

.PHONY: fmt
fmt:
	@echo '.....Formatting go code.....'
	go fmt ./...

.PHONY:lint
lint: fmt
	@echo '.....Linting go code.....'
	golint ./...

.PHONY: vet
vet: lint
	@echo '.....Vetting go code.....'
	go vet ./...

## format: format & lint all go files.
.PHONY: format
format: vet

# ======================================================================== #
# DEVELOPMENT
# ======================================================================== #

## start: run the ./cmd/api/ application.
.PHONY: start
start:
	@go run ./cmd/api/ -db-dsn=${SUNRISE_API_POSTGRES_DSN}

## db/migrations/new name=$1: create a new database migration.
.PHONY: db/migrations/new
db/migrations/new:
	@echo '.....Creating migration files for ${name}.....'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations.
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo '.....Running up migrations.....'
	migrate -path ./migrations -database ${SUNRISE_API_POSTGRES_DSN} up

