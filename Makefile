# CU - Central University CLI Tool
# Makefile for local development

# Variables
BINARY_NAME=cuni
BIN_DIR=bin
BUILD_DIR=build
CMD_DIR=cmd/cuni
INSTALL_DIR?=$(shell go env GOPATH)/bin
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE?=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags="-s -w -X main.ver=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

# Default target
.PHONY: all
all: test build

# Clean build directory
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -rf $(BIN_DIR)
	rm -f $(BINARY_NAME)

# Download dependencies
.PHONY: deps
deps:
	go mod download
	go mod verify

# Run tests
.PHONY: test
test:
	go test -v -race -coverprofile=coverage.out ./...

# End-to-end test for install.sh: builds a release archive, serves it over
# localhost and drives the real installer. No network, no published release.
.PHONY: test-install
test-install:
	./test/e2e/install_test.sh

# Run tests with coverage report
.PHONY: test-coverage
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Generate internal/configure/embedded/.env for go:embed.
# Sources, in priority order: ./.env file, then OPDASHBOARD_* / OP_API_URL
# environment variables (CI). Neither present -> remove stale file and build
# a binary with telemetry disabled (noop).
EMBED_ENV=internal/configure/embedded/.env

.PHONY: embed-env
embed-env:
	@mkdir -p internal/configure/embedded
	@if [ -f .env ]; then \
		cp .env $(EMBED_ENV); \
		echo "embed-env: telemetry config from ./.env"; \
	elif [ -n "$$OPDASHBOARD_CLIENT_ID" ] && [ -n "$$OPDASHBOARD_SECRET" ] && [ -n "$$OP_API_URL" ]; then \
		printf 'OPDASHBOARD_CLIENT_ID="%s"\nOPDASHBOARD_SECRET="%s"\nOP_API_URL="%s"\n' \
			"$$OPDASHBOARD_CLIENT_ID" "$$OPDASHBOARD_SECRET" "$$OP_API_URL" > $(EMBED_ENV); \
		echo "embed-env: telemetry config from environment"; \
	else \
		rm -f $(EMBED_ENV); \
		echo "embed-env: no telemetry config found — binary will be built with telemetry disabled"; \
	fi

# Build for current platform
.PHONY: build
build: embed-env
	@mkdir -p $(BIN_DIR)
	go build $(LDFLAGS) -o $(BIN_DIR)/$(BINARY_NAME) ./$(CMD_DIR)

# Install to $(INSTALL_DIR) (defaults to $GOPATH/bin).
# Builds then atomically replaces the existing binary — works even when an MCP
# process holds the old inode open (where `go install` would silently no-op).
# Re-signs on macOS so the new binary keeps its entitlements.
.PHONY: install
install: build
	@mkdir -p $(INSTALL_DIR)
	cp $(BIN_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)
	@if [ "$$(uname)" = "Darwin" ]; then \
		codesign --force --sign - $(INSTALL_DIR)/$(BINARY_NAME) 2>/dev/null || true; \
	fi
	@echo "Installed $(INSTALL_DIR)/$(BINARY_NAME)"
	@echo "If Claude Code (or another host) has the MCP server running, restart it:"
	@echo "  claude mcp restart cuni"

# Build for all platforms
.PHONY: build-all
build-all: clean embed-env
	mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-linux-amd64 ./$(CMD_DIR)
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-linux-arm64 ./$(CMD_DIR)
	
	# macOS AMD64 (Intel)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-amd64 ./$(CMD_DIR)
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-darwin-arm64 ./$(CMD_DIR)
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-windows-amd64.exe ./$(CMD_DIR)
	
	# Windows ARM64
	GOOS=windows GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(VERSION)-windows-arm64.exe ./$(CMD_DIR)
	
	@echo "Built binaries:"
	@ls -la $(BUILD_DIR)/

# Create checksums for all binaries
.PHONY: checksums
checksums: build-all
	cd $(BUILD_DIR) && for file in $(BINARY_NAME)-*; do \
		sha256sum "$$file" > "$$file.sha256"; \
	done
	@echo "Checksums created:"
	@ls -la $(BUILD_DIR)/*.sha256

# Run linting
.PHONY: lint
lint:
	go vet ./...
	go fmt ./...
	golangci-lint run --timeout=5m

# Run gofumpt formatter
.PHONY: gofumpt
gofumpt:
	@which gofumpt > /dev/null || (echo "gofumpt not found. Install it with: make install-gofumpt" && exit 1)
	gofumpt -l -w .

# Install gofumpt
.PHONY: install-gofumpt
install-gofumpt:
	@echo "Installing gofumpt..."
	@go install mvdan.cc/gofumpt@latest
	@echo "✅ gofumpt installed successfully"
	@gofumpt --version

# Run golangci-lint (requires golangci-lint to be installed)
.PHONY: golangci-lint
golangci-lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install it with: make install-golangci-lint" && exit 1)
	golangci-lint run --timeout=5m

# Install golangci-lint
.PHONY: install-golangci-lint
install-golangci-lint:
	@echo "Installing golangci-lint v2.7.2..."
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin v2.7.2
	@echo "✅ golangci-lint installed successfully"
	@golangci-lint --version

# Run all linting tools
.PHONY: lint-all
lint-all: lint gofumpt golangci-lint

# Run the application (requires CU_BFF_COOKIE environment variable)
.PHONY: run
run: build
	$(BIN_DIR)/$(BINARY_NAME) $(ARGS)

# Run with example command
.PHONY: run-help
run-help: build
	$(BIN_DIR)/$(BINARY_NAME) --help

# Show help
.PHONY: help
help:
	@echo "cuni - Central University CLI Tool"
	@echo ""
	@echo "Available commands:"
	@echo "  make build        - Build for current platform"
	@echo "  make embed-env    - Generate embedded telemetry config from .env / env vars"
	@echo "  make build-all    - Build for all platforms"
	@echo "  make test         - Run tests"
	@echo "  make test-install - E2E test for install.sh (macOS/Linux)"
	@echo "  make test-coverage- Run tests with coverage report"
	@echo "  make clean        - Clean build directory"
	@echo "  make deps         - Download dependencies"
	@echo "  make install      - Install to GOPATH/bin"
	@echo "  make lint         - Run basic linting (go vet, go fmt)"
	@echo "  make gofumpt      - Run gofumpt formatter"
	@echo "  make golangci-lint- Run golangci-lint"
	@echo "  make lint-all     - Run all linting tools"
	@echo "  make install-gofumpt - Install gofumpt"
	@echo "  make install-golangci-lint - Install golangci-lint"
	@echo "  make checksums    - Create checksums for all binaries"
	@echo "  make run ARGS=... - Run the application"
	@echo "  make run-help     - Show application help"
	@echo "  make help         - Show this help"
	@echo ""
	@echo "Environment variables:"
	@echo "  VERSION          - Override version (default: git describe)"
	@echo "  CU_BFF_COOKIE    - Cookie for Central University authentication"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make test"
	@echo "  CU_BFF_COOKIE='your-cookie' make run ARGS='fetch courses'"
	@echo "  make build-all && make checksums"

