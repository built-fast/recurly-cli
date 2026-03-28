BINARY   := recurly
MODULE   := github.com/built-fast/recurly-cli
VERSION  := $(shell git describe --tags --always --dirty)
LDFLAGS  := -ldflags "-X $(MODULE)/cmd.version=$(VERSION)"

.DEFAULT_GOAL := build

.PHONY: all fmt vet lint test build clean

all: fmt vet lint test build

fmt:
	gofmt -w .

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test ./...

build:
	go build $(LDFLAGS) -o $(BINARY) .

clean:
	rm -f $(BINARY)
