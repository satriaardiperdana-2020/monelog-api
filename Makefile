GO ?= go
GOFMT ?= $(shell $(GO) env GOROOT)/bin/gofmt
GO_FILES := $(shell find cmd internal -type f -name '*.go' 2>/dev/null)

.PHONY: build run fmt fmt-check tidy test test-race vet verify check

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

check: fmt-check test vet build verify
