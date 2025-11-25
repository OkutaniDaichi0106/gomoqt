package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"sync"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:9000", "server address")
	flag.Parse()

	slog.Info(" === MOQ Interop Test ===")

	// Create context for server lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background (run server package under cmd/interop)
	serverCmd := exec.CommandContext(ctx, "go", "run", "./server", "-addr", *addr)

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
			serverCmd.Process.Kill()
			serverCmd.Wait()
			slog.Debug(" Server process terminated")
		}
	}()

	// Run client and wait for completion (run client package under cmd/interop)
	slog.Debug(" Starting client...")
	clientCmd := exec.CommandContext(ctx, "go", "run", "./client", "-addr", "https://"+*addr)

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
	err = clientCmd.Wait()
	if err != nil {
		slog.Error("Client failed: " + err.Error())
	}

	// Stop the server to unblock output streams
	cancel()

	// Wait for all streaming goroutines to finish
	wg.Wait()

	slog.Info(" === Interop Test Completed ===")
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
