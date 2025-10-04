# Test Makefile for callgraph-mcp

.PHONY: test test-unit test-integration test-all clean build-for-test

# Default target
test: test-unit

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v ./unit/...

# Run integration tests
test-integration: build-for-test
	@echo "Running integration tests..."
	go test -v ./integration/...

# Run all Go tests
test-all: build-for-test
	@echo "Running all tests..."
	go test -v ./...

# Run Python integration tests
test-python: build-for-test
	@echo "Running Python integration tests..."
	cd integration && python mcp_server_test.py

# Build binary for integration tests
build-for-test:
	@echo "Building callgraph-mcp for testing..."
	cd .. && go build -o callgraph-mcp .

# Clean test artifacts
clean:
	@echo "Cleaning test artifacts..."
	rm -f ../callgraph-mcp

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
test-bench:
	@echo "Running benchmarks..."
	go test -bench=. ./...

# Lint tests
test-lint:
	@echo "Linting test files..."
	golangci-lint run ./...

# Help
help:
	@echo "Available targets:"
	@echo "  test           - Run unit tests (default)"
	@echo "  test-unit      - Run unit tests only"
	@echo "  test-integration - Run integration tests"
	@echo "  test-all       - Run all Go tests"
	@echo "  test-python    - Run Python integration tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  test-bench     - Run benchmarks"
	@echo "  test-lint      - Lint test files"
	@echo "  build-for-test - Build binary for testing"
	@echo "  clean          - Clean test artifacts"
	@echo "  help           - Show this help"