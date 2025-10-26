go install github.com/magefile/mage@latest
  ```

### Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/OkutaniDaichi0106/gomoqt.git
   cd gomoqt
   ```

2. Install the package:
   ```bash
   go get github.com/OkutaniDaichi0106/gomoqt
   ```

3. Set up the development environment:
   ```bash
   mage dev-setup
   ```

This command will perform the following:
- Install the required certificate tools (mkcert).
- Install development tools (goimports, golangci-lint).
- Download project dependencies.
- Generate development certificates.

### Development Commands

Mage tasks are defined in `magefile.go`. You can list available tasks with: