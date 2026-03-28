BINARY   := recurly
MODULE   := github.com/built-fast/recurly-cli
VERSION  := $(shell git describe --tags --always --dirty)
LDFLAGS  := -ldflags "-X $(MODULE)/cmd.version=$(VERSION)"

.DEFAULT_GOAL := build

.PHONY: all fmt fmt-check tidy-check vet lint test test-e2e check build clean surface check-surface vuln

all: fmt vet lint test build

fmt:
	gofmt -w .

fmt-check:
	@test -z "$$(gofmt -l .)" || (gofmt -l . && exit 1)

tidy-check:
	@cp go.mod go.mod.bak && cp go.sum go.sum.bak
	@go mod tidy
	@diff -q go.mod go.mod.bak > /dev/null 2>&1 && diff -q go.sum go.sum.bak > /dev/null 2>&1; \
		STATUS=$$?; \
		mv go.mod.bak go.mod; \
		mv go.sum.bak go.sum; \
		if [ $$STATUS -ne 0 ]; then \
			echo "go.mod/go.sum not tidy — run: go mod tidy"; \
			exit 1; \
		fi

vet:
	go vet ./...

lint:
	golangci-lint run

test:
	go test ./...

test-e2e:
	./e2e/run.sh

check: fmt-check vet lint tidy-check check-surface test test-e2e vuln

surface:
	go run ./internal/surface/cmd/gensurface > .surface

check-surface:
	@go run ./internal/surface/cmd/gensurface | diff -u .surface - || (echo "CLI surface has changed. If intentional, run: make surface" && exit 1)

vuln:
	govulncheck ./...

build:
	go build $(LDFLAGS) -o $(BINARY) .

clean:
	rm -f $(BINARY)
