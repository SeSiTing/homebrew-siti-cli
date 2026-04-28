# siti-cli Makefile
# Convenience targets for local development. Release is driven by goreleaser
# in CI (see .github/workflows/publish-on-version-bump.yml).

BINARY      := siti
PKG         := github.com/SeSiTing/siti-cli
VERSION     := $(shell grep -oE '^var version = "[0-9]+\.[0-9]+\.[0-9]+"' version.go | grep -oE '[0-9]+\.[0-9]+\.[0-9]+')
LDFLAGS     := -s -w -X main.version=$(VERSION)
GO          ?= go

.PHONY: help build install run test vet tidy fmt lint clean completions snapshot release-check

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## Build the siti binary into ./siti
	$(GO) build -ldflags '$(LDFLAGS)' -o $(BINARY) .

install: ## Install siti into $GOBIN (or $GOPATH/bin)
	$(GO) install -ldflags '$(LDFLAGS)' .

run: ## Run via `go run .` (forward args with: make run ARGS="ai list")
	$(GO) run . $(ARGS)

test: ## Run unit tests
	$(GO) test -race -count=1 ./...

vet: ## Run go vet
	$(GO) vet ./...

tidy: ## Tidy modules and verify nothing changed (CI-friendly)
	$(GO) mod tidy
	@git diff --exit-code go.mod go.sum || (echo "go.mod/go.sum dirty after tidy — commit the changes" && exit 1)

fmt: ## Format Go code
	$(GO) fmt ./...

lint: vet ## Lint (currently just vet — extend with golangci-lint if added)

completions: build ## Regenerate shell completions
	mkdir -p completions
	./$(BINARY) completion zsh  > completions/_siti
	./$(BINARY) completion bash > completions/siti.bash

snapshot: ## Local goreleaser dry-run (requires goreleaser binary)
	goreleaser release --snapshot --clean --skip=publish

release-check: ## Verify version.go version vs current git tag
	@echo "version.go:  $(VERSION)"
	@echo "latest tag:  $$(git describe --tags --abbrev=0 2>/dev/null || echo none)"

clean: ## Remove build artifacts
	rm -rf $(BINARY) dist/
