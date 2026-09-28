BIN_DIR ?= bin
BINARY ?= $(BIN_DIR)/forgedash
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "v0.1.0")
COMMIT_SHA ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")
BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

LDFLAGS := -X github.com/forgedash/forgedash/internal/interfaces/cli.Version=$(VERSION) \
           -X github.com/forgedash/forgedash/internal/interfaces/cli.CommitSHA=$(COMMIT_SHA) \
           -X github.com/forgedash/forgedash/internal/interfaces/cli.BuildDate=$(BUILD_DATE)

.PHONY: all build run demo test test-race fmt vet lint clean

all: build

build:
	@mkdir -p $(BIN_DIR)
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/forgedash

run: build
	./$(BINARY)

demo: build
	./$(BINARY) --demo

test:
	go test ./...

test-race:
	go test -race ./...

fmt:
	gofmt -s -w .

vet:
	go vet ./...

lint: fmt vet

clean:
	rm -rf $(BIN_DIR)
