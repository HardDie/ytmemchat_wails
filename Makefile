# Common developer commands. Run `make help` for the list.

APP_NAME     := ytmemchat
INTERNAL     := ./internal/...
# Exclude the Wails entrypoint (CGO) so package main tests run on CI.
TEST_MAIN_TAGS := nomain
GO           := go
WAILS        := wails
PKG          ?= ./internal/tts
PKGSITE_ADDR ?= localhost:8081
# Exact git tag when HEAD is tagged, otherwise the short commit. Override with BUILD_VERSION=…
BUILD_VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || git rev-parse --short=12 HEAD 2>/dev/null || echo dev)
VERSION_LDFLAGS := -X github.com/HardDie/ytmemchat_wails/bindings/home.buildVersion=$(BUILD_VERSION)
# Ubuntu 24.04+ ships webkit2gtk-4.1; Wails needs this tag instead of 4.0.
WAILS_TAGS := $(shell pkg-config --exists webkit2gtk-4.1 2>/dev/null && echo -tags webkit2_41)

.DEFAULT_GOAL := help

.PHONY: help dev build generate test test-integration test-all \
	vet fmt tidy doc doc-all docs-site frontend-install screenshots clean ci

## help: Show this list
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //' | awk -F': ' '{printf "  %-22s %s\n", $$1, $$2}'

## dev: Run the Wails app with frontend hot reload
dev: require-wails
	$(WAILS) dev $(WAILS_TAGS) -ldflags "$(VERSION_LDFLAGS)"

## build: Production binary for this machine (build/bin)
build: require-wails
	$(WAILS) build $(WAILS_TAGS) -clean -trimpath -ldflags "$(VERSION_LDFLAGS)"

## generate: Regenerate frontend/wailsjs bindings from Go
generate: require-wails
	$(WAILS) generate module

## test: Unit tests for package main, bindings, and internal packages (same as CI, with race)
test:
	$(GO) test -race -count=1 -tags=$(TEST_MAIN_TAGS) .
	$(GO) test -race -count=1 -tags=$(TEST_MAIN_TAGS) ./bindings/...
	$(GO) test -race -count=1 $(INTERNAL)

## test-integration: Integration tests (OS TTS, later HTTP); skips if tools missing
test-integration:
	$(GO) test -tags=integration -count=1 $(INTERNAL)

## test-all: Unit then integration tests
test-all: test test-integration

## ci: What GitHub Actions test.yml runs
ci: test-all

## vet: Go vet on package main (no Wails CGO) and internal packages
vet:
	$(GO) vet -tags=$(TEST_MAIN_TAGS) .
	$(GO) vet -tags=$(TEST_MAIN_TAGS) ./bindings/...
	$(GO) vet $(INTERNAL)

## fmt: Format Go files; fail if any file needed formatting
fmt:
	@files="$$(gofmt -l $$(find . -name '*.go' -not -path './frontend/*'))"; \
	if [ -n "$$files" ]; then echo "$$files"; echo "run: gofmt -w on the files above"; exit 1; fi

## tidy: go mod download and tidy
tidy:
	$(GO) mod download
	$(GO) mod tidy

## doc: Package summary (PKG=./internal/tts)
doc:
	$(GO) doc $(PKG)

## doc-all: Package + all exports (PKG=./internal/tts)
doc-all:
	$(GO) doc -all $(PKG)

## docs-site: HTML godoc at PKGSITE_ADDR (default localhost:8081)
docs-site:
	$(GO) run golang.org/x/pkgsite/cmd/pkgsite@latest -http $(PKGSITE_ADDR)

## frontend-install: npm install in frontend/
frontend-install: require-wails
	npm install --prefix frontend

## screenshots: Refresh README window.gif (Home → Config → Commands → Test)
screenshots:
	cd scripts/screenshots && npm install && npx playwright install chromium && node capture.mjs

## clean: Remove Wails/Go build artifacts
clean:
	$(GO) clean
	rm -rf build/bin frontend/dist

require-wails:
	@test -f wails.json || (echo "wails.json missing: scaffold with wails init before this target"; exit 1)
