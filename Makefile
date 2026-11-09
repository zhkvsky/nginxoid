# Build, run, and code quality workflows for nginxoid CLI.
GO ?= go
GOLANGCI_LINT ?= golangci-lint
ARGS ?=
PORT ?= 8000
BINARY ?= bin/nginxoid
CMD_PATH ?= ./cmd/nginxoid
LOG_SERVER_PATH ?= ./cmd/log-server

.PHONY: install build run serve format check test test-race clean

install: ## Download project dependencies
	$(GO) mod download

build: ## Compile the CLI binary
	$(GO) build -trimpath -o $(BINARY) $(CMD_PATH)

run: ## Run the CLI with optional ARGS (e.g. make run ARGS="--path ...")
	$(GO) run $(CMD_PATH) $(ARGS)

serve: ## Start local HTTP server serving sample NGINX logs
	$(GO) run $(LOG_SERVER_PATH) --port $(PORT)

format: ## Format codebase using gofumpt and goimports
	$(GOLANGCI_LINT) fmt

check: ## Verify go.mod tidy, formatting, and run linters
	$(GO) mod tidy -diff
	$(GOLANGCI_LINT) config verify
	$(GOLANGCI_LINT) fmt --diff
	$(GOLANGCI_LINT) run

test: ## Run unit tests and generate HTML coverage report
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out
	$(GO) tool cover -html=coverage.out -o coverage.html

test-race: ## Run unit tests with data race detector enabled
	$(GO) test -race ./...

clean: ## Remove build artifacts and coverage files
	rm -rf bin coverage.out coverage.html