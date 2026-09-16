GO ?= go
GOFMT ?= $(shell $(GO) env GOROOT)/bin/gofmt
MIGRATE ?= migrate
SQLC ?= sqlc
GO_FILES := $(shell find cmd internal -type f -name '*.go' 2>/dev/null)

.PHONY: build run fmt fmt-check tidy test test-race test-integration vet verify check sqlc-generate sqlc-check sqlc-vet oapi-generate oapi-check migrate-up migrate-down runtime-grants

build:
	mkdir -p bin
	$(GO) build -o bin/monelog-api ./cmd/api

run:
	$(GO) run ./cmd/api

fmt:
	$(GOFMT) -w $(GO_FILES)

fmt-check:
	@test -z "$$($(GOFMT) -l $(GO_FILES))"

tidy:
	$(GO) mod tidy

test:
	$(GO) test ./...

test-race:
	$(GO) test -race ./...

vet:
	$(GO) vet ./...

verify:
	$(GO) mod verify

sqlc-generate:
	$(SQLC) generate

sqlc-check: sqlc-generate
	git diff --exit-code -- internal/repository/sqlc

sqlc-vet:
	$(SQLC) vet

oapi-generate:
	$(GO) generate ./api

oapi-check: oapi-generate
	git diff --exit-code -- internal/api

migrate-up:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required" >&2; exit 1)
	$(MIGRATE) -path db/migrations -database "$(DATABASE_URL)" up

migrate-down:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required" >&2; exit 1)
	$(MIGRATE) -path db/migrations -database "$(DATABASE_URL)" down 1

runtime-grants:
	@test -n "$(DATABASE_URL)" || (echo "DATABASE_URL is required" >&2; exit 1)
	@test -n "$(RUNTIME_DB_ROLE)" || (echo "RUNTIME_DB_ROLE is required" >&2; exit 1)
	psql "$(DATABASE_URL)" -v ON_ERROR_STOP=1 -v runtime_role="$(RUNTIME_DB_ROLE)" -f db/roles/runtime.sql

test-integration:
	@test -n "$(TEST_DATABASE_URL)" || (echo "TEST_DATABASE_URL is required" >&2; exit 1)
	$(GO) test -tags=integration ./internal/repository/sqlc

check: fmt-check test vet build verify sqlc-vet sqlc-check oapi-check
