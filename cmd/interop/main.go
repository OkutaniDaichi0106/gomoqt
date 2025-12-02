package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

func main() {
	addr := flag.String("addr", "localhost:9000", "server address")
	lang := flag.String("lang", "go", "client language: go or ts")
	flag.Parse()

	slog.Info(" === MOQ Interop Test ===")

	// Create context for server lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background (run server package under cmd/interop)
	// Server binds to :port (all interfaces), not hostname:port
	_, port, _ := strings.Cut(*addr, ":")
	serverCmd := exec.CommandContext(ctx, "go", "run", "./server", "-addr", ":"+port)

	// Pipe server output so we can reformat and unify logs
	serverStdout, err := serverCmd.StdoutPipe()
	if err != nil {
		slog.Error("Failed to capture server stdout: " + err.Error())
		return
	}
	serverStderr, err := serverCmd.StderrPipe()
	if err != nil {
		slog.Error("Failed to capture server stderr: " + err.Error())
		return
	}

	err = serverCmd.Start()
	if err != nil {
		slog.Error("Failed to start server: " + err.Error())
		return
	}

	// Ensure server is killed when we exit
	// Stream server output in background
	var wg sync.WaitGroup
	// Channels to detect when the child server or client declare completion
	// serverReady := make(chan struct{}, 1)
	wg.Go(func() {
		streamAndLog("Server", serverStdout)
	})
	wg.Go(func() {
		streamAndLog("Server", serverStderr)
	})

	defer func() {
		if serverCmd.Process != nil {
			slog.Debug(" Killing server process...")
			_ = serverCmd.Process.Kill()
			_ = serverCmd.Wait()
			slog.Debug(" Server process terminated")
		}
	}()

	// Wait for server to be ready
	time.Sleep(2 * time.Second)

	// Run client and wait for completion
	slog.Debug(" Starting client...")
	var clientCmd *exec.Cmd
	if *lang == "ts" {
		// TypeScript client - run from moq-web directory
		// Get the directory where this main.go is located, then find moq-web relative to project root
		wd, _ := os.Getwd()
		// We're in cmd/interop, so go up two levels to project root, then into moq-web
		moqWebDir := filepath.Join(wd, "moq-web")
		// If running from cmd/interop, adjust path
		if _, err := os.Stat(moqWebDir); os.IsNotExist(err) {
			moqWebDir = filepath.Join(wd, "..", "..", "moq-web")
		}
		clientCmd = exec.CommandContext(ctx, "deno", "run", "--unstable-net", "--allow-all",
			"cli/interop/run_secure.ts", "--addr", "https://"+*addr)
		clientCmd.Dir = moqWebDir
	} else {
		// Go client
		clientCmd = exec.CommandContext(ctx, "go", "run", "./client", "-addr", "https://"+*addr)
	}

	clientStdout, err := clientCmd.StdoutPipe()
	if err != nil {
		slog.Error("Failed to capture client stdout: " + err.Error())
		return
	}
	clientStderr, err := clientCmd.StderrPipe()
	if err != nil {
		slog.Error("Failed to capture client stderr: " + err.Error())
		return
	}

	if err = clientCmd.Start(); err != nil {
		slog.Error(" Failed to start client: " + err.Error())
		return
	}

	wg.Go(func() {
		streamAndLog("Client", clientStdout)
	})
	wg.Go(func() {
		streamAndLog("Client", clientStderr)
	})

	// Wait for the client process to finish running and let the server
	// continue until it declares its operation complete as well. Use a
	// timeout to avoid waiting forever in case things fail.
	clientErr := clientCmd.Wait()

	// Stop the server to unblock output streams
	cancel()

	// Kill server process immediately
	if serverCmd.Process != nil {
		_ = serverCmd.Process.Kill()
		_ = serverCmd.Wait()
	}

	// Wait for all streaming goroutines to finish
	wg.Wait()

	slog.Info(" === Interop Test Completed ===")

	if clientErr != nil {
		slog.Error("Client failed: " + clientErr.Error())
		os.Exit(1)
	}
}

// streamAndLog reads from reader line by line and logs each line prefixed with
// the source name. It also notifies channels when specific patterns are found.
func streamAndLog(source string, r io.Reader) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Printf("[%s] %s\n", source, line)
	}
	if err := scanner.Err(); err != nil {
		slog.Warn("Error reading output stream", "error", err)
	}
}
