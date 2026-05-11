.PHONY: build test lint clean

BINARY := pipeline
GOFLAGS := -trimpath
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -X github.com/kdobmayer/kond-test-go-cli/cmd.version=$(VERSION) \
           -X github.com/kdobmayer/kond-test-go-cli/cmd.commit=$(COMMIT) \
           -X github.com/kdobmayer/kond-test-go-cli/cmd.buildDate=$(BUILD_DATE)

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
