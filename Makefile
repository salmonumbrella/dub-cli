.PHONY: build test lint fmt setup

# Version info
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS := -s -w
LDFLAGS += -X github.com/salmonumbrella/dub-cli/internal/cmd.Version=$(VERSION)
LDFLAGS += -X github.com/salmonumbrella/dub-cli/internal/cmd.Commit=$(COMMIT)
LDFLAGS += -X github.com/salmonumbrella/dub-cli/internal/cmd.Date=$(DATE)

build:
	go build -ldflags "$(LDFLAGS)" -o dub ./cmd/dub

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/dub

test:
	go test -v ./...

lint:
	golangci-lint run

fmt:
	gofumpt -w .
	goimports -w .

setup:
	@command -v lefthook >/dev/null 2>&1 || (echo "Installing lefthook..." && go install github.com/evilmartians/lefthook@latest)
	lefthook install
