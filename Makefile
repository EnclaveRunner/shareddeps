.PHONY: test test-verbose test-coverage test-race bench clean verify fmt lint oapi update-version

# Default target
all: test

# Run tests
test:
	go test ./...

# Run tests with verbose output
test-verbose:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -cover ./... -coverprofile=coverage.out

# Run tests with race detection
test-race:
	go test -race ./...

# Run benchmarks
bench:
	go test -bench=.

test-all: test-verbose test-coverage test-race bench

# Format code
fmt:
	golangci-lint fmt

# Lint code (requires golangci-lint to be installed)
lint:
	golangci-lint run --fix

# Clean test cache
clean:
	go clean -testcache

# Generate OpenAPI server code from spec
oapi:
	go generate tools.go

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
# Update version references in files
update-version:
	@if [ ! -f "Version" ]; then \
		echo "❌ Version file not found"; \
		exit 1; \
	fi
	@VERSION=$$(cat Version | tr -d '\n'); \
	if [ -z "$$VERSION" ]; then \
		echo "❌ Version file is empty"; \
		exit 1; \
	fi; \
	echo " Current version: $$VERSION"; \
	echo " Searching for version patterns to update..."; \
	find . -type f \( -name "*.go" -o -name "*.md" -o -name "*.yml" -o -name "*.yaml" -o -name "*.json" \) \
		-not -path "./.git/*" \
		-not -path "./vendor/*" \
		-not -path "./.direnv/*" \
		-exec grep -l "v[0-9]\+\.[0-9]\+\.[0-9]\+" {} \; | \
	while read -r file; do \
		echo "   Updating $$file"; \
		sed -i.bak "s/v[0-9]\+\.[0-9]\+\.[0-9]\+/$$VERSION/g" "$$file" && rm -f "$$file.bak"; \
	done; \
	echo " Version update complete!"

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
	@echo "  lint          - Lint and fix code"
	@echo "  clean         - Clean test cache"
	@echo "  oapi          - Create gin server from OpenAPI spec"
	@echo "  verify        - Simulate CI Checks before opening a PR"
	@echo "  update-version - Update all version references to match Version file"
	@echo "  help          - Show this help"
