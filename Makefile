# Common developer commands. Run `make help` for the list.
# Wails targets need a scaffolded project (wails.json). Go test targets work today.

APP_NAME     := ytmemchat
INTERNAL     := ./internal/...
GO           := go
WAILS        := wails
PKG          ?= ./internal/tts
PKGSITE_ADDR ?= localhost:8081

.DEFAULT_GOAL := help

.PHONY: help dev build generate test test-integration test-all \
	vet fmt tidy doc doc-all docs-site frontend-install clean ci

## help: Show this list
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/^## //' | awk -F': ' '{printf "  %-22s %s\n", $$1, $$2}'

## dev: Run the Wails app with frontend hot reload
dev: require-wails
	$(WAILS) dev

## build: Production binary for this machine (build/bin)
build: require-wails
	$(WAILS) build -clean -trimpath

## generate: Regenerate frontend/wailsjs bindings from Go
generate: require-wails
	$(WAILS) generate module

## test: Unit tests for internal packages (same as CI, with race)
test:
	$(GO) test -race -count=1 $(INTERNAL)

## test-integration: Integration tests (OS TTS, later HTTP); skips if tools missing
test-integration:
	$(GO) test -tags=integration -count=1 $(INTERNAL)

## test-all: Unit then integration tests
test-all: test test-integration

## ci: What GitHub Actions test.yml runs
ci: test-all

## vet: Go vet on internal packages
vet:
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

## clean: Remove Wails/Go build artifacts
clean:
	$(GO) clean
	rm -rf build/bin frontend/dist

require-wails:
	@test -f wails.json || (echo "wails.json missing: scaffold with wails init before this target"; exit 1)
