// Package main provides the Chrome Native Messaging Host for Verbalizer.
// This binary communicates with the Chrome extension via stdin/stdout
// using the Chrome Native Messaging protocol, and forwards messages
// to the Verbalizer daemon via Unix socket.
package main

import (
	"fmt"
	"os"

	"github.com/fulvian/verbalizer/native-host/internal/ipc"
	"github.com/fulvian/verbalizer/native-host/internal/messaging"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	client := ipc.NewClient("")

	// Read messages from stdin in Native Messaging format
	for {
		msg, err := messaging.ReadMessage(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read message: %w", err)
		}

		response := handleMessage(client, msg)
		if err := messaging.WriteMessage(os.Stdout, response); err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	}
}

// handleMessage processes incoming messages from the extension and forwards them to the daemon.
func handleMessage(client *ipc.Client, msg *messaging.Message) *messaging.Response {
	var daemonCmd string
	switch msg.Type {
	case "START_RECORDING":
		daemonCmd = "START_RECORDING"
	case "STOP_RECORDING":
		daemonCmd = "STOP_RECORDING"
	case "GET_STATUS":
		daemonCmd = "GET_STATUS"
	// Cloud commands
	case "GOOGLE_AUTH_START":
		daemonCmd = "GOOGLE_AUTH_START"
	case "GOOGLE_AUTH_STATUS":
		daemonCmd = "GOOGLE_AUTH_STATUS"
	case "GOOGLE_AUTH_DISCONNECT":
		daemonCmd = "GOOGLE_AUTH_DISCONNECT"
	case "GOOGLE_DRIVE_SET_FOLDER":
		daemonCmd = "GOOGLE_DRIVE_SET_FOLDER"
	case "GOOGLE_DRIVE_GET_FOLDER":
		daemonCmd = "GOOGLE_DRIVE_GET_FOLDER"
	case "GOOGLE_DRIVE_SYNC_STATUS":
		daemonCmd = "GOOGLE_DRIVE_SYNC_STATUS"
	case "GOOGLE_DRIVE_SYNC_RETRY":
		daemonCmd = "GOOGLE_DRIVE_SYNC_RETRY"
	case "ping":
		return &messaging.Response{Success: true, Data: "pong"}
	default:
		return &messaging.Response{Success: false, Error: fmt.Sprintf("unknown message type: %s", msg.Type)}
	}

	// Forward to daemon via Unix socket
	resp, err := client.Send(daemonCmd, msg.Payload)
	if err != nil {
		return &messaging.Response{Success: false, Error: fmt.Sprintf("ipc error: %v", err)}
	}

	return &messaging.Response{
		Success: resp.Success,
		Data:    resp.Data,
		Error:   resp.Error,
	}
}
