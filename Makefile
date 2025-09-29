# CU - Central University CLI Tool
# Makefile for local development

# Variables
BINARY_NAME=cu
BUILD_DIR=build
CMD_DIR=cmd/cli
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags="-s -w -X main.version=$(VERSION)"

# Default target
.PHONY: all
all: test build

# Clean build directory
.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	rm -f bin/cu

# Download dependencies
.PHONY: deps
deps:
	go mod download
	go mod verify

# Run tests
.PHONY: test
test:
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
.PHONY: test-coverage
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Build for current platform
.PHONY: build
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./$(CMD_DIR)

# Install to GOPATH/bin
.PHONY: install
install:
	go install $(LDFLAGS) ./$(CMD_DIR)

# Build for all platforms
.PHONY: build-all
build-all: clean
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

# Run the application (requires CU_BFF_COOKIE environment variable)
.PHONY: run
run: build
	./$(BINARY_NAME) $(ARGS)

# Run with example command
.PHONY: run-help
run-help: build
	./$(BINARY_NAME) --help

# Show help
.PHONY: help
help:
	@echo "CU - Central University CLI Tool"
	@echo ""
	@echo "Available commands:"
	@echo "  make build        - Build for current platform"
	@echo "  make build-all    - Build for all platforms"
	@echo "  make test         - Run tests"
	@echo "  make test-coverage- Run tests with coverage report"
	@echo "  make clean        - Clean build directory"
	@echo "  make deps         - Download dependencies"
	@echo "  make install      - Install to GOPATH/bin"
	@echo "  make lint         - Run linting tools"
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

