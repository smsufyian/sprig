BINARY    = dist/sprig
MODULE    = github.com/smsufyian/sprig
VERSION   = $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT    = $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
DATE      = $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS   = -X $(MODULE)/internal/version.Version=$(VERSION) \
            -X $(MODULE)/internal/version.Commit=$(COMMIT) \
            -X $(MODULE)/internal/version.BuildDate=$(DATE)

.PHONY: build test lint generate clean install

build:
	@mkdir -p dist
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/sprig
	@echo "Built $(BINARY)"

test:
	go test -race -coverprofile=coverage.out ./...

lint:
	golangci-lint run

generate:
	go generate ./...

clean:
	rm -rf dist/ coverage.out

install:
	go install -ldflags "$(LDFLAGS)" ./cmd/sprig

# Cross-compile for all release targets
build-all:
	@mkdir -p dist
	GOOS=linux  GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/sprig_linux_amd64   ./cmd/sprig
	GOOS=linux  GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/sprig_linux_arm64   ./cmd/sprig
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o dist/sprig_darwin_amd64  ./cmd/sprig
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o dist/sprig_darwin_arm64  ./cmd/sprig
	@echo "Cross-compiled all targets to dist/"
