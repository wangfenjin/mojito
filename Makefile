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


# Build migrator binary
.PHONY: build-migrator
build-migrator:
	@mkdir -p bin
	@go build -o bin/migrator cmd/migrator/main.go

# Generate migration SQL files
.PHONY: gen-migration
gen-migration: build-migrator
	./bin/migrator

# Clean all build artifacts and generated files
.PHONY: clean
clean:
	@echo "Cleaning build artifacts and generated files..."
	@rm -rf bin/
	@rm -rf scripts/db/
	@rm -rf coverage.out coverage.html


# API Testing with Hurl
.PHONY: test-api
test-api:
	@echo "Running API tests with Hurl..."
	@hurl --test --variable host=http://localhost:8080 tests/login.hurl
	@hurl --test --variable host=http://localhost:8080 tests/users.hurl
	@hurl --test --variable host=http://localhost:8080 tests/items.hurl

# Run all tests (unit tests and API tests)
.PHONY: test-all
test-all: test test-api
	@echo "All tests completed!"