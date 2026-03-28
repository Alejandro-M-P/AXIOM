.PHONY: test test-unit test-race test-coverage test-verbose help

# ============================================
# AXIOM - Makefile for Development
# ============================================

# Go parameters
GOCMD=go
GOTEST=$(GOCMD) test
GOBUILD=$(GOCMD) build

# Directories
CMD=cmd/axiom
INTERNAL=internal
TESTS=tests

# ============================================
# Tests
# ============================================

## test: Run all tests with race detector
test:
	@echo "🔍 Running all tests..."
	@echo ""
	$(GOTEST) -v -race ./$(TESTS)/...
	$(GOTEST) -v -race ./$(INTERNAL)/...
	$(GOTEST) -v -race ./$(CMD)/...
	@echo ""
	@echo "✅ All tests passed!"

## test-unit: Run unit tests without race detector (faster)
test-unit:
	@echo "🔍 Running unit tests..."
	$(GOTEST) -v ./$(TESTS)/...
	$(GOTEST) -v ./$(INTERNAL)/...
	$(GOTEST) -v ./$(CMD)/...

## test-race: Run tests with race detector (comprehensive)
test-race:
	@echo "🏃 Running tests with race detector..."
	$(GOTEST) -v -race ./$(TESTS)/...
	$(GOTEST) -v -race ./$(INTERNAL)/...
	$(GOTEST) -v -race ./$(CMD)/...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "📊 Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./$(TESTS)/...
	$(GOTEST) -v -race -coverprofile=coverage.out ./$(INTERNAL)/...
	$(GOTEST) -v -race -coverprofile=coverage.out ./$(CMD)/...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo ""
	@echo "📄 Coverage report: coverage.html"

## test-verbose: Run tests with verbose output
test-verbose:
	@echo "📝 Running verbose tests..."
	$(GOTEST) -v ./...

# ============================================
# Build
# ============================================

## build: Build the binary
build:
	@echo "🔨 Building axiom..."
	$(GOBUILD) -o axiom ./$(CMD)

## build-linux: Build for Linux
build-linux:
	@echo "🔨 Building axiom for Linux..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o axiom-linux ./$(CMD)

# ============================================
# Development
# ============================================

## fmt: Format code
fmt:
	@echo "🎨 Formatting code..."
	$(GOCMD) fmt ./...

## vet: Run go vet
vet:
	@echo "🔍 Running go vet..."
	$(GOCMD) vet ./...

## lint: Run vet and static analysis
lint: vet
	@echo "✅ Linting complete..."

## tidy: Clean dependencies
tidy:
	@echo "🧹 Cleaning dependencies..."
	$(GOCMD) mod tidy

# ============================================
# Utilities
# ============================================

## clean: Clean build artifacts
clean:
	@echo "🧹 Cleaning..."
	rm -f axiom axiom-linux coverage.out coverage.html
	rm -rf $(GOCMD)-workspace

## help: Show this help
help:
	@echo "AXIOM - Development Commands"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
