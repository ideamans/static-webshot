.PHONY: all build test unit-test test-cover clean install lint fmt tidy help

VERSION ?= 0.1.0
BINARY := static-webshot
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

all: build

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/staticwebshot

test:
	go test -v ./...

unit-test:
	go test -v -race ./pkg/...

unit-test-coverage:
	go test -v -race -coverprofile=coverage.out ./pkg/...

test-cover:
	go test -cover ./...

clean:
	rm -f $(BINARY) $(BINARY).exe
	rm -f coverage.out coverage.html

install: build
	cp $(BINARY) $(GOPATH)/bin/

lint:
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run ./...

fmt:
	go fmt ./...

tidy:
	go mod tidy

help:
	@echo "Available targets:"
	@echo "  make build              - Build the binary"
	@echo "  make test               - Run all tests"
	@echo "  make unit-test          - Run unit tests with race detector"
	@echo "  make unit-test-coverage - Run unit tests with coverage"
	@echo "  make test-cover         - Run tests with coverage summary"
	@echo "  make clean              - Remove build artifacts"
	@echo "  make install            - Build and install to GOPATH/bin"
	@echo "  make lint               - Run linter"
	@echo "  make fmt                - Format code"
	@echo "  make tidy               - Run go mod tidy"
