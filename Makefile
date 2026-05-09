.PHONY: build test lint clean

BINARY := pipeline
GOFLAGS := -trimpath

build:
	go build $(GOFLAGS) -o $(BINARY) .

test:
	go test -race -cover ./...

lint:
	go vet ./...
	@if command -v staticcheck >/dev/null 2>&1; then staticcheck ./...; fi

clean:
	rm -f $(BINARY)
	go clean -testcache
