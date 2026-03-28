BINARY   := recurly
MODULE   := github.com/built-fast/recurly-cli
VERSION  := $(shell git describe --tags --always --dirty)
LDFLAGS  := -ldflags "-X $(MODULE)/cmd.version=$(VERSION)"

.DEFAULT_GOAL := build

.PHONY: all fmt vet lint test test-e2e check build clean surface check-surface

all: fmt vet lint test build

fmt:
	gofmt -w .

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test ./...

test-e2e:
	./e2e/run.sh

check: test test-e2e check-surface

surface:
	go run ./internal/surface/cmd/gensurface > .surface

check-surface:
	@go run ./internal/surface/cmd/gensurface | diff -u .surface - || (echo "CLI surface has changed. If intentional, run: make surface" && exit 1)

build:
	go build $(LDFLAGS) -o $(BINARY) .

clean:
	rm -f $(BINARY)
