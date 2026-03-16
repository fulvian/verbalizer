// Package main provides the Verbalizer daemon.
// This daemon runs as a background service and handles:
// - Audio capture from system audio (macOS) or PipeWire (Linux)
// - Transcription using whisper.cpp
// - Storage of recordings and transcripts
// - IPC server for communication with the native host
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

const (
	// DefaultSocketPath is the default Unix socket path for IPC.
	DefaultSocketPath = "/tmp/verbalizer.sock"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Create context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		fmt.Println("Shutting down...")
		cancel()
	}()

	fmt.Println("Verbalizer daemon starting...")
	fmt.Printf("Socket path: %s\n", DefaultSocketPath)

	// TODO: Initialize components
	// - IPC server
	// - Audio capture
	// - Transcriber
	// - Storage

	// Wait for shutdown
	<-ctx.Done()

	fmt.Println("Verbalizer daemon stopped")
	return nil
}
