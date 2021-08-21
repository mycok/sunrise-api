## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY:lint
lint: fmt
	golint ./...

.PHONY: vet
vet: lint
	go vet ./...

## format: format & lint all go files
.PHONY: format
format: vet

.PHONY: confirm
confirm:
	@echo -n 'Are you sure? [Y/N] ' && read ans && [ $${ans:-N} = y ]

## start: run the ./cmd/api/ application
.PHONY: start
start:
	go run ./cmd/api/

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo '.....Creating migration files for ${name}.....'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo '.....Running up migrations.....'
	migrate -path ./migrations -database ${SUNRISE_API_POSTGRES_DSN} up

