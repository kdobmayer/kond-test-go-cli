.PHONY: build test lint clean

BINARY := pipeline
GOFLAGS := -trimpath
VERSION := $(shell git describe --tags --always 2>/dev/null || echo dev)
COMMIT  := $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.buildDate=$(DATE)

build:
	go build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY) .

test:
	go test -race -cover ./...

lint:
	go vet ./...
	@if command -v staticcheck >/dev/null 2>&1; then staticcheck ./...; fi

clean:
	rm -f $(BINARY)
	go clean -testcache
