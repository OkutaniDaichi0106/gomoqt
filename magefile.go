//go:build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// ======================================
// SETUP
// ======================================

type Setup mg.Namespace

func (Setup) All() error {
	fmt.Println("Setting up development environment...")
	// Install golangci-lint
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		fmt.Println("Installing golangci-lint...")
		if err := sh.RunV("go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"); err != nil {
			return err
		}
	}

	// Install Deno
	if _, err := exec.LookPath("deno"); err != nil {
		fmt.Println("Installing Deno...")
		if err := sh.RunV("sh", "-c", "curl -fsSL https://deno.land/x/install/install.sh | sh"); err != nil {
			return err
		}
	}

	return nil
}

func (Setup) Go() error {
	fmt.Println("Setting up Go environment...")

	// Check Go version
	fmt.Println("Checking Go version... (go version)")
	if err := goVersion(); err != nil {
		return err
	}

	// Install Go tools
	// Install golangci-lint
	if _, err := exec.LookPath("golangci-lint"); err != nil {
		fmt.Println("Installing golangci-lint...")
		if err := sh.RunV("go", "install", "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"); err != nil {
			return err
		}
	}

	fmt.Println("Go environment setup complete.")

	return nil
}

func goVersion() error {
	out, err := exec.Command("go", "version").Output()
	if err != nil {
		return err
	}
	version := string(out)
	remote := struct {
		major int
		minor int
	}{
		major: 1,
		minor: 25,
	}

	required := struct {
		major int
		minor int
	}{
		major: 1,
		minor: 22,
	}

	re := regexp.MustCompile(`go version go([0-9]+)\.([0-9]+)`)
	matches := re.FindStringSubmatch(version)
	if len(matches) > 2 {
		major, _ := strconv.Atoi(matches[1])
		minor, _ := strconv.Atoi(matches[2])

		fmt.Printf("go version: %d.%d (local) | %d.%d (repository)", major, minor, remote.major, remote.minor)
		if major < required.major || (major == required.major && minor < required.minor) {
			fmt.Printf("   └─ >= %d.%d required", required.major, required.minor)
		}
	} else {
		return fmt.Errorf("failed to parse Go version from: %s", version)
	}

	return nil
}

func (Setup) Deno() error {
	fmt.Println("Setting up Deno environment...")

	// Check Deno version
	fmt.Println("Checking Deno version... (deno --version)")
	if err := denoVersion(); err != nil {
		return err
	}

	// Install Deno
	if _, err := exec.LookPath("deno"); err != nil {
		fmt.Println("Installing Deno...")
		if err := sh.RunV("sh", "-c", "curl -fsSL https://deno.land/x/install/install.sh | sh"); err != nil {
			return err
		}
	}

	return nil
}

func denoVersion() error {
	out, err := exec.Command("deno", "--version").Output()
	if err != nil {
		return err
	}
	output := string(out)
	lines := strings.Split(strings.TrimSpace(output), "\n")

	// Extract versions using regex
	re := regexp.MustCompile(`^(deno|v8|typescript)\s+([^\s]+)`)
	versions := map[string]string{}

	for _, line := range lines {
		if match := re.FindStringSubmatch(line); match != nil {
			versions[match[1]] = match[2]
		}
	}
	remote := struct {
		major int
		minor int
	}{
		major: 2,
		minor: 5,
	}

	// Format and output versions
	fmt.Println(
		"deno: %s, v8: %s, typescript: %s (local) | %d.%d (repository)",
		versions["deno"],
		versions["v8"],
		versions["typescript"],
		remote.major,
		remote.minor,
	)

	required := struct {
		major int
		minor int
	}{
		major: 2,
		minor: 0,
	}

	major, _ := strconv.Atoi(strings.Split(versions["deno"], ".")[0])
	minor, _ := strconv.Atoi(strings.Split(versions["deno"], ".")[1])
	if major < required.major || (major == required.major && minor < required.minor) {
		fmt.Printf("   └─ >= %d.%d required", required.major, required.minor)
	}
	return nil
}

// ======================================
// TESTING
// ======================================

type Test mg.Namespace

// Test runs all tests in the project
func (t Test) All() error {
	fmt.Println("Running tests...")
	return sh.RunV("go", "test", "./...")
}

// Coverage runs tests with coverage reporting
func (t Test) Coverage() error {
	fmt.Println("Running tests with coverage...")
	return sh.RunV("deno", "test", "--coverage=coverage")
}

// ======================================
// INTEROPERABILITY
// ======================================

type Server mg.Namespace

func (s Server) Default() error {
	fmt.Println("Setting up default server environment...")
	s.Go()
	return nil
}

func (Server) Go() error {
	fmt.Println("Setting up Go environment...")
	return nil
}

type Client mg.Namespace

func (c Client) Default() error {
	fmt.Println("Setting up default client environment...")
	c.Go()
	return nil
}

func (c Client) All() error {
	fmt.Println("Setting up all client environments...")
	c.Go()
	c.Deno()
	c.Node()
	c.Bun()
	c.Chrome()
	c.Firefox()
	c.Safari()
	return nil
}

func (Client) Go() error {
	fmt.Println("Setting up Go environment...")
	return nil
}

func (Client) Deno() error {
	fmt.Println("Setting up Deno environment...")
	return nil
}

func (Client) Node() error {
	fmt.Println("Setting up Node.js environment...")
	return nil
}

func (Client) Bun() error {
	fmt.Println("Setting up Bun environment...")
	return nil
}

func (Client) Chrome() error {
	fmt.Println("Setting up Chrome environment...")
	return nil
}

func (Client) Firefox() error {
	fmt.Println("Setting up Firefox environment...")
	return nil
}

func (Client) Safari() error {
	fmt.Println("Setting up Safari environment...")
	return nil
}

// ======================================
// INTEROP
// ======================================

type Interop mg.Namespace

// Go runs the interop test
func (Interop) Go() error {
	fmt.Println("Running interop test...")

	// Save current working directory
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	// Restore working directory when done
	defer func() {
		_ = os.Chdir(wd)
	}()

	// Change to interop directory
	if err := os.Chdir("cmd/interop"); err != nil {
		return err
	}

	return sh.RunV("go", "run", ".")
}

// ======================================
// DEVELOPMENT UTILITIES
// ======================================

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
	fmt.Println("Formatting go code...")
	if err := sh.RunV("go", "fmt", "./..."); err != nil {
		return err
	}

	fmt.Println("Formatting TypeScript code...")
	if err := sh.RunV("deno", "fmt"); err != nil {
		return err
	}
	return nil
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
	fmt.Println("  mage build  - Build the project")
	fmt.Println("  mage clean  - Clean up generated files")
	fmt.Println("  mage help   - Show this help message")
	fmt.Println("")
	fmt.Println("You can also run 'mage -l' to list all available targets.")
}

// Default target - displays help when no target is specified
var Default = Help
