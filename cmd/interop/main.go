package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func main() {
	slog.Info("=== MOQ Interop Test ===")

	// Wait for port to be available before starting
	const serverAddr = "127.0.0.1:9000"
	if !openPort(serverAddr, 10*time.Second) {
		slog.Error("✗ Port is still in use, cannot start test")
		return
	}

	// Check and generate certificates if needed
	if err := mkcert(); err != nil {
		slog.Error("Failed to setup certificates: " + err.Error())
		return
	}

	// Create context for server lifecycle
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start server in background
	serverCmd := exec.CommandContext(ctx, "go", "run", "./server")
	serverCmd.Stdout = os.Stdout
	serverCmd.Stderr = os.Stderr

	err := serverCmd.Start()
	if err != nil {
		slog.Error("Failed to start server: " + err.Error())
		return
	}

	// Ensure server is killed when we exit
	defer func() {
		if serverCmd.Process != nil {
			slog.Debug("Killing server process...")
			serverCmd.Process.Kill()
			serverCmd.Wait()
			slog.Debug("Server process terminated")

			// Wait for port to be released
			if !openPort(serverAddr, 15*time.Second) {
				slog.Warn("Port may still be in use after cleanup")
			}
		}
	}()

	// Wait for server to start
	slog.Debug("Waiting for server to start...")
	time.Sleep(2 * time.Second)

	// Run client and wait for completion
	slog.Debug("Starting client...")
	clientCmd := exec.Command("go", "run", "./client")
	clientCmd.Stdout = os.Stdout
	clientCmd.Stderr = os.Stderr

	err = clientCmd.Run()
	if err != nil {
		slog.Error("✗ Client failed: " + err.Error())
	} else {
		slog.Info("✓ Client completed successfully")
	}

	slog.Info("=== Interop Test Completed ===")
}

func mkcert() error {
	serverCertPath := filepath.Join("server", "moqt.example.com.pem")

	// Check if server certificates exist
	if _, err := os.Stat(serverCertPath); os.IsNotExist(err) {
		slog.Debug("Server certificates not found, generating with mkcert...")
		cmd := exec.Command("mkcert", "moqt.example.com")
		cmd.Dir = "server"
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			slog.Error("Failed to generate certificates: " + err.Error())
			return err
		}
		slog.Debug("Server certificates generated successfully")
	}
	return nil
}

// openPort waits until the specified port becomes available
// Returns true if port becomes available within timeout, false otherwise
func openPort(addr string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	attempt := 0
	for time.Now().Before(deadline) {
		attempt++
		// Try to bind to the UDP port
		conn, err := net.ListenPacket("udp", addr)
		if err == nil {
			// Port is available
			conn.Close()
			slog.Debug("Port is available", "addr", addr, "attempts", attempt)
			return true
		}

		// Port still in use, wait a bit
		if attempt == 1 || attempt%4 == 0 { // Log every 2 seconds
			slog.Debug("Port still in use, waiting...", "addr", addr, "attempt", attempt)
		}
		time.Sleep(500 * time.Millisecond)
	}

	slog.Warn("Timeout waiting for port to become available", "addr", addr, "attempts", attempt)
	return false
}
