// Package ipc provides client for daemon communication.
package ipc

import (
	"encoding/json"
	"fmt"
	"net"
	"time"
)

const (
	// DefaultSocketPath is the default Unix socket path.
	DefaultSocketPath = "/tmp/verbalizer.sock"

	// Timeout for IPC operations.
	Timeout = 5 * time.Second
)

// Client communicates with the daemon via Unix socket.
type Client struct {
	socketPath string
}

// NewClient creates a new IPC client.
func NewClient(socketPath string) *Client {
	if socketPath == "" {
		socketPath = DefaultSocketPath
	}
	return &Client{socketPath: socketPath}
}

// Send sends a command to the daemon and returns the response.
func (c *Client) Send(cmdType string, payload interface{}) (*Response, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, Timeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Set deadline
	conn.SetDeadline(time.Now().Add(Timeout))

	// Encode command
	var payloadBytes json.RawMessage
	if payload != nil {
		bytes, err := json.Marshal(payload)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal payload: %w", err)
		}
		payloadBytes = bytes
	}

	cmd := &Command{
		Type:    cmdType,
		Payload: payloadBytes,
	}

	// Send command
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(cmd); err != nil {
		return nil, fmt.Errorf("failed to encode command: %w", err)
	}

	// Read response
	var resp Response
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resp, nil
}

// Command represents a command sent to the daemon.
type Command struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// Response represents a response from the daemon.
type Response struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   string          `json:"error,omitempty"`
}
