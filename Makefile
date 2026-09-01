.PHONY: build test lint run clean

BINARY=weavelens
GO=go
GOCACHE := $(shell pwd)/.cache/gocache
GOMODCACHE := $(shell pwd)/.cache/gomodcache
export GOCACHE GOMODCACHE

build:
	$(GO) build -o bin/$(BINARY) ./cmd/weavelens

test:
	$(GO) test ./...

lint:
	$(GO) vet ./...
	gofmt -w .

run: build
	./bin/$(BINARY)

clean:
	rm -rf bin/ .cache/
