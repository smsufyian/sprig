BINARY = dist/sprig

.PHONY: build test clean

build:
	@mkdir -p dist
	go build -o $(BINARY) ./cmd/sprig

test:
	go test ./...

clean:
	rm -rf dist/
