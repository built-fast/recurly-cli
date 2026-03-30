SHELL := /bin/bash

# ── Project metadata ──────────────────────────────────────────────
BINARY    := recurly
MODULE    := github.com/built-fast/recurly-cli
BUILD_DIR := ./bin

# ── Version info (injected via ldflags) ───────────────────────────
VERSION := $(shell git describe --tags --always --dirty)
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

LDFLAGS := -s -w \
  -X $(MODULE)/cmd.version=$(VERSION) \
  -X $(MODULE)/cmd.commit=$(COMMIT) \
  -X $(MODULE)/cmd.date=$(DATE)

# ── Go tool aliases ───────────────────────────────────────────────
GOCMD   := go
GOBUILD := CGO_ENABLED=0 $(GOCMD) build -trimpath
GOTEST  := $(GOCMD) test
GOVET   := $(GOCMD) vet
GOFMT   := gofmt
GOMOD   := $(GOCMD) mod

.DEFAULT_GOAL := build

# ── Build ─────────────────────────────────────────────────────────
.PHONY: all build install run clean

all: fmt vet lint test build

build:
	$(GOBUILD) -ldflags '$(LDFLAGS)' -o $(BUILD_DIR)/$(BINARY) .

install:
	$(GOCMD) install -ldflags '$(LDFLAGS)' ./...

run: build
	$(BUILD_DIR)/$(BINARY)

clean:
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html

# ── Format & Lint ─────────────────────────────────────────────────
.PHONY: fmt fmt-check vet lint tidy tidy-check

fmt:
	$(GOFMT) -w .

fmt-check:
	@test -z "$$($(GOFMT) -l .)" || ($(GOFMT) -l . && exit 1)

vet:
	$(GOVET) ./...

lint:
	golangci-lint run

tidy:
	$(GOMOD) tidy

tidy-check:
	@cp go.mod go.mod.bak && cp go.sum go.sum.bak
	@$(GOMOD) tidy
	@diff -q go.mod go.mod.bak > /dev/null 2>&1 && diff -q go.sum go.sum.bak > /dev/null 2>&1; \
		STATUS=$$?; \
		mv go.mod.bak go.mod; \
		mv go.sum.bak go.sum; \
		if [ $$STATUS -ne 0 ]; then \
			echo "go.mod/go.sum not tidy — run: make tidy"; \
			exit 1; \
		fi

# ── Test ──────────────────────────────────────────────────────────
.PHONY: test test-race test-e2e test-coverage coverage

test:
	$(GOTEST) ./...

test-race:
	$(GOTEST) -race ./...

test-e2e:
	./e2e/run.sh

test-coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

coverage: test-coverage
	@open coverage.html 2>/dev/null || xdg-open coverage.html 2>/dev/null || true

# ── Security ──────────────────────────────────────────────────────
.PHONY: vuln secrets security

vuln:
	govulncheck ./...

secrets:
	gitleaks detect --source . --verbose

security: vuln secrets

# ── Surface & Skills ──────────────────────────────────────────────
.PHONY: surface check-surface check-skill-drift

surface:
	go run ./internal/surface/cmd/gensurface > .surface

check-surface:
	@go run ./internal/surface/cmd/gensurface | diff -u .surface - || (echo "CLI surface has changed. If intentional, run: make surface" && exit 1)

check-skill-drift:
	@./scripts/check-skill-drift.sh

# ── CI Gate ───────────────────────────────────────────────────────
.PHONY: check

check: fmt-check vet lint tidy-check check-surface check-skill-drift test test-e2e vuln

# ── Module ────────────────────────────────────────────────────────
.PHONY: verify

verify:
	$(GOMOD) verify

# ── GitHub Actions ────────────────────────────────────────────────
.PHONY: lint-actions

lint-actions:
	actionlint
	zizmor .github/workflows/

# ── Release ───────────────────────────────────────────────────────
.PHONY: test-release

test-release:
	goreleaser release --snapshot --clean

# ── Tools ─────────────────────────────────────────────────────────
.PHONY: tools

tools:
	brew bundle

# ── Help ──────────────────────────────────────────────────────────
.PHONY: help

help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Build:"
	@echo "  build          Build binary to $(BUILD_DIR)/$(BINARY)"
	@echo "  install        Install binary via go install"
	@echo "  run            Build and run"
	@echo "  clean          Remove build artifacts and coverage files"
	@echo ""
	@echo "Format & Lint:"
	@echo "  fmt            Format Go source files"
	@echo "  fmt-check      Check formatting (fails if not formatted)"
	@echo "  vet            Run go vet"
	@echo "  lint           Run golangci-lint"
	@echo "  tidy           Run go mod tidy"
	@echo "  tidy-check     Check that go.mod/go.sum are tidy"
	@echo ""
	@echo "Test:"
	@echo "  test           Run unit tests"
	@echo "  test-race      Run unit tests with race detector"
	@echo "  test-e2e       Run e2e tests (BATS + Prism)"
	@echo "  test-coverage  Generate coverage report"
	@echo "  coverage       Generate and open coverage report"
	@echo ""
	@echo "Security:"
	@echo "  vuln           Run govulncheck"
	@echo "  secrets        Run gitleaks secret scanning"
	@echo "  security       Run all security checks"
	@echo ""
	@echo "Surface & Skills:"
	@echo "  surface        Regenerate .surface snapshot"
	@echo "  check-surface  Verify .surface is up to date"
	@echo "  check-skill-drift  Check SKILL.md matches .surface"
	@echo ""
	@echo "CI:"
	@echo "  check          Run full CI gate"
	@echo ""
	@echo "Other:"
	@echo "  verify         Verify module dependencies"
	@echo "  lint-actions   Lint GitHub Actions workflows"
	@echo "  test-release   Goreleaser dry-run"
	@echo "  tools          Install dev tools via Brewfile"
	@echo "  help           Show this help"
