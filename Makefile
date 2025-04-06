# note: call scripts from /scripts

.PHONY: build run clean watch test test-verbose test-coverage

# Build the mojito application
build:
	@echo "Building mojito..."
	@go build -o bin/mojito ./cmd/mojito/main.go

# Run the mojito application
run:
	@echo "Running mojito..."
	@go run ./cmd/mojito/main.go

# Build and run the mojito application
build-run: build
	@echo "Starting mojito..."
	@./bin/mojito

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/mojito

# Watch for changes and automatically rebuild and restart
watch:
	@echo "Watching for changes..."
	@if ! command -v air > /dev/null; then \
		echo "Installing air..."; \
		go install github.com/air-verse/air@latest; \
	fi
	@air -c .air.toml

# Run tests
test:
	@echo "Running tests..."
	@go test ./internal/... ./cmd/...

# Run tests with verbose output
test-verbose:
	@echo "Running tests with verbose output..."
	@go test -v ./internal/... ./cmd/...

# Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage report..."
	@go test -cover ./internal/... ./cmd/...
	@echo "For detailed coverage report, run: make test-coverage-html"

# Generate HTML coverage report
test-coverage-html:
	@echo "Generating HTML coverage report..."
	@go test -coverprofile=coverage.out ./internal/... ./cmd/...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@open coverage.html