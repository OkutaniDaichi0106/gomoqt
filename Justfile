# Justfile - Development tasks for gomoqt
#
# Usage:
#   just [command]
#
# Available commands:
#   dev-setup       Set up the development environment
#   run-echo-server Start the echo server
#   run-echo-client Start the echo client
#   fmt             Format Go source code
#   lint            Run linter
#   test            Run tests
#   check           Overall quality checks (formatting and linting)
#   build           Build the project
#   clean           Clean up generated files
#   help            Show this help message
#
# By default, help is executed.
default: help

help:
	@echo "Available commands:"
	@echo "  just dev-setup       Set up the development environment"
	@echo "  just run-echo-server Start the echo server"
	@echo "  just run-echo-client Start the echo client"
	@echo "  just fmt             Format Go source code"
	@echo "  just lint            Run linter"
	@echo "  just test            Run tests"
	@echo "  just check           Overall quality checks (formatting and linting)"
	@echo "  just build           Build the project"
	@echo "  just clean           Clean up generated files"
	@echo "  just help            Show this help message"

# Build target example:
build:
	@echo "Building project..."
	go build ./...

# Test target example:
test:
	@echo "Running tests..."
	go test ./...

# Run target example:
run:
	@echo "Running project..."
	# Define generic run command if needed
	go run .

# New command: dev-setup
dev-setup:
	@echo "Setting up development environment..."
	@echo "Installing certificate tools (mkcert)..."
	mkcert -install || true
	@echo "Installing development tools (goimports, golangci-lint)..."
	go install golang.org/x/tools/cmd/goimports@latest
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Downloading project dependencies..."
	go mod tidy
	@echo "Generating development certificates..."
	# (Add commands for generating certificates if necessary)

# New command: run-push-server
run-push-server:
	@echo "Starting push server..."
	go run ./examples/push/server/main.go

# New command: run-push-client
run-push-client:
	@echo "Starting push client..."
	go run ./examples/push/client/main.go

# New command: fmt
fmt:
	@echo "Formatting code..."
	go fmt ./...

# New command: lint
lint:
	@echo "Running linter..."
	golangci-lint run

# New command: check (depends on fmt and lint)
check: fmt lint
	@echo "Quality checks complete."

# New command: clean
clean:
	@echo "Cleaning up generated files..."
	# Remove binaries or other generated files as necessary (e.g., the ./bin directory)
	rm -rf ./bin

set windows-shell := ["C:\\Program Files\\Git\\bin\\sh.exe","-c"]