BINARY := tvnamer
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build test clean install lint fmt vet

build:
	go build $(LDFLAGS) -o $(BINARY) .

test:
	go test ./...

test-verbose:
	go test -v ./...

fmt:
	go fmt ./...

vet:
	go vet ./...

lint: fmt vet

clean:
	rm -f $(BINARY)

install:
	go install $(LDFLAGS) .
