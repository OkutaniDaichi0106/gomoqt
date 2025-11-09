package main

import (
	"context"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

func main() {
	slog.Info("Starting moq interop test")

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

	// Run client and wait for completion
	clientCmd := exec.Command("go", "run", "./client")
	clientCmd.Stdout = os.Stdout
	clientCmd.Stderr = os.Stderr

	slog.Info("Starting client...")
	err = clientCmd.Run()
	if err != nil {
		slog.Error("Client failed: " + err.Error())
		// Continue to cleanup
	} else {
		slog.Info("Client completed successfully")
	}

	// Stop server and wait for it to exit
	slog.Info("Stopping server...")
	cancel() // Signal context cancellation

	// Wait for server process to completely exit
	waitErr := serverCmd.Wait()
	if waitErr != nil {
		slog.Warn("Server exited with error: " + waitErr.Error())
	} else {
		slog.Info("Server stopped successfully")
	}

	if err != nil {
		slog.Error("Interop test failed")
		return
	}

	slog.Info("Interop test completed successfully")
}

func mkcert() error {
	serverCertPath := filepath.Join("server", "moqt.example.com.pem")

	// Check if server certificates exist
	if _, err := os.Stat(serverCertPath); os.IsNotExist(err) {
		slog.Info("Server certificates not found, generating with mkcert...")
		cmd := exec.Command("mkcert", "moqt.example.com")
		cmd.Dir = "server"
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			slog.Error("Failed to generate certificates: " + err.Error())
			return err
		}
		slog.Info("Server certificates generated successfully")
	}
	return nil
}
