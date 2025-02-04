# List all available recipes
default:
    @just --list

# Install required dependencies
setup:
    # Install certificate tools
    if ! command -v mkcert >/dev/null 2>&1; then \
        echo "Installing mkcert..."; \
        if command -v brew >/dev/null 2>&1; then \
            brew install mkcert; \
        elif command -v apt-get >/dev/null 2>&1; then \
            sudo apt-get install -y mkcert; \
        elif command -v choco >/dev/null 2>&1; then \
            choco install mkcert; \
        else \
            echo "Please install mkcert manually: https://github.com/FiloSottile/mkcert#installation"; \
        fi; \
    fi

    # Install development tools
    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

    # Install project dependencies
    go mod tidy
    go install github.com/quic-go/quic-go@v0.48.2
    go install github.com/quic-go/webtransport-go@v0.8.1

# Generate development certificates (using either mkcert or OpenSSL)
generate-cert:
    mkdir -p examples/cert
    if command -v mkcert >/dev/null 2>&1; then
        echo "Using mkcert to generate certificates..."
        mkcert -install
        mkcert -key-file cert/key.pem -cert-file cert/cert.pem localhost 127.0.0.1 ::1
    else
        echo "mkcert not found, falling back to OpenSSL..."
        openssl req -x509 -newkey rsa:2048 -keyout cert/key.pem -out cert/cert.pem -days 365 -nodes -subj "/CN=localhost"
    fi

# Build and run the echo server
run-echo-server: generate-cert
    go run examples/echo/server/main.go

# Build and run the echo client
run-echo-client: generate-cert
    go run examples/echo/client/main.go

# Run tests
test:
    go test -v ./...

# Build the code
build:
    go build -v ./...

# Clean up generated files
clean:
    rm -rf cert
    go clean

# Run complete development setup
dev-setup: setup generate-cert

# Format code
fmt:
    goimports -w .
    go fmt ./...

# Run linter
lint:
    golangci-lint run

# Check code quality (format and lint)
check: fmt lint