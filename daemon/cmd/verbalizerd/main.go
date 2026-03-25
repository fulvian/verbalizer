package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/fulvian/verbalizer/daemon/internal/cloud/driveclient"
	cloudmgr "github.com/fulvian/verbalizer/daemon/internal/cloud/manager"
	"github.com/fulvian/verbalizer/daemon/internal/cloud/syncqueue"
	"github.com/fulvian/verbalizer/daemon/internal/config"
	"github.com/fulvian/verbalizer/daemon/internal/ipc"
	"github.com/fulvian/verbalizer/daemon/internal/secrets"
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

	// Initialize secrets store
	secretsDir := filepath.Join(cfg.DataDir, ".secrets")
	secretsStore, err := secrets.NewFileSecretStore(secretsDir)
	if err != nil {
		return fmt.Errorf("failed to initialize secrets store: %w", err)
	}

	// Initialize components
	sessionMgr, err := session.NewManager(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize session manager: %w", err)
	}

	// Initialize cloud manager
	cloudMgr := cloudmgr.NewManager(cfg, sessionMgr.GetDatabase(), secretsStore)

	// Initialize sync queue worker if cloud is enabled
	var syncWorker *syncqueue.Worker
	if cfg.IsCloudEnabled() {
		driveClient := driveclient.NewDriveClient(&cfg.Cloud, secretsStore)
		syncWorker = syncqueue.NewWorker(sessionMgr.GetDatabase(), driveClient, &cfg.Cloud.Retry)
		syncWorker.Start()
		sessionMgr.SetSyncQueue(syncWorker)
		fmt.Println("Cloud sync worker started")
	}

	handler := NewDaemonHandler(sessionMgr, cloudMgr)
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
		if syncWorker != nil {
			syncWorker.Stop()
		}
		cancel()
	}()

	fmt.Println("Verbalizer daemon starting...")
	fmt.Printf("Socket path: %s\n", ipc.DefaultSocketPath)
	fmt.Printf("Recordings dir: %s\n", cfg.RecordingsDir)
	fmt.Printf("Cloud sync enabled: %v\n", cfg.IsCloudEnabled())

	// Wait for shutdown
	<-ctx.Done()

	fmt.Println("Verbalizer daemon stopped")
	return nil
}
