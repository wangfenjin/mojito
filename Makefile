# note: call scripts from /scripts

.PHONY: build run clean watch

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