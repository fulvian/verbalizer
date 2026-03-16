package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/fulvian/verbalizer/daemon/internal/config"
	"github.com/fulvian/verbalizer/daemon/internal/ipc"
	"github.com/fulvian/verbalizer/daemon/internal/session"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := cfg.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to ensure directories: %w", err)
	}

	// Initialize components
	sessionMgr, err := session.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize session manager: %w", err)
	}
	handler := NewDaemonHandler(sessionMgr)
	server := ipc.NewServer(ipc.DefaultSocketPath, handler)

	// Start IPC server
	if err := server.Start(); err != nil {
		return fmt.Errorf("failed to start IPC server: %w", err)
	}
	defer server.Stop()

	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("\nShutting down...")
		cancel()
	}()

	fmt.Println("Verbalizer daemon starting...")
	fmt.Printf("Socket path: %s\n", ipc.DefaultSocketPath)
	fmt.Printf("Recordings dir: %s\n", cfg.RecordingsDir)

	// Wait for shutdown
	<-ctx.Done()

	fmt.Println("Verbalizer daemon stopped")
	return nil
}
