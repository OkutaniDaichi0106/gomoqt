//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/magefile/mage/sh"
)

// Test runs all tests in the project
func Test() error {
	fmt.Println("Running tests...")
	return sh.RunV("go", "test", "./...")
}

// Lint runs the linter (golangci-lint)
func Lint() error {
	fmt.Println("Running linter...")
	// Check if golangci-lint is available
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		return fmt.Errorf("golangci-lint not found. Please install it first:\n  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest")
	}
	return sh.RunV("golangci-lint", "run")
}

// Fmt formats Go source code
func Fmt() error {
	fmt.Println("Formatting code...")
	return sh.RunV("go", "fmt", "./...")
}

// Build builds the project
func Build() error {
	fmt.Println("Building project...")
	return sh.RunV("go", "build", "./...")
}

// Clean removes generated files
func Clean() error {
	fmt.Println("Cleaning up generated files...")
	// Remove binaries directory if it exists
	if err := sh.Rm("./bin"); err != nil {
		// Ignore errors if directory doesn't exist
		if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

// Help displays available commands (default target)
func Help() {
	fmt.Println("Available Mage commands:")
	fmt.Println("  mage test   - Run all tests")
	fmt.Println("  mage lint   - Run golangci-lint")
	fmt.Println("  mage fmt    - Format Go source code")
	fmt.Println("  mage build  - Build the project")
	fmt.Println("  mage clean  - Clean up generated files")
	fmt.Println("  mage help   - Show this help message")
	fmt.Println("")
	fmt.Println("You can also run 'mage -l' to list all available targets.")
}

// Default target - displays help when no target is specified
var Default = Help
