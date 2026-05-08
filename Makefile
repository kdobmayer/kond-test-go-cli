.PHONY: build test clean

build:
	go build -o tasks .

test:
	go test ./...

clean:
	rm -f tasks
