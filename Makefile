.PHONY: build test clean install

VERSION := 0.1.0
BINARY := pvrt
LDFLAGS := -ldflags "-X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/pvrt

test:
	go test -v ./...

test-cover:
	go test -cover ./...

clean:
	rm -f $(BINARY)
	rm -f coverage.out

install: build
	cp $(BINARY) $(GOPATH)/bin/

lint:
	golangci-lint run

fmt:
	go fmt ./...

tidy:
	go mod tidy
