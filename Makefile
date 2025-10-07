.PHONY: test test-verbose test-coverage test-race bench clean verify fmt lint

# Default target
all: test

# Run tests
test:
	go test . ./config/...

# Run tests with verbose output
test-verbose:
	go test -v . ./config/...

# Run tests with coverage
test-coverage:
	go test -cover . ./config/... -coverprofile=coverage.out

# Run tests with race detection
test-race:
	go test -race . ./config/...

# Run benchmarks
bench:
	go test -bench=.

test-all: test-verbose test-coverage test-race bench

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint to be installed)
lint:
	golangci-lint run

# Clean test cache
clean:
	go clean -testcache

# Simulate CI tests
verify:
	@echo "Running CI tests..."
	@echo "Checking Linting:"
	make lint
	@echo "Checking Tests:"
	make test-all
	@echo "Checking Build:"
	go build -v ./...
	go clean -testcache
	@echo "✅ CI Test will pass, you are ready to commit / open the PR! Thank you for your contribution :)"
# Show help
help:
	@echo "Available targets:"
	@echo "  test          - Run basic tests"
	@echo "  test-verbose  - Run tests with verbose output"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  test-race     - Run tests with race detection"
	@echo "  bench         - Run benchmarks"
	@echo "  test-all      - Run all tests"
	@echo "  fmt           - Format code"
	@echo "  lint          - Lint code"
	@echo "  clean         - Clean test cache"
	@echo "  verify        - Simulate CI Checks before opening a PR"
	@echo "  help          - Show this help"
