// Package ipc provides client for daemon communication.
package ipc

import (
	"encoding/json"
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
		return nil, err
	}
	defer conn.Close()

	// Set deadline
	conn.SetDeadline(time.Now().Add(Timeout))

	// Encode command
	payloadBytes, _ := json.Marshal(payload)
	cmd := &Command{
		Type:    cmdType,
		Payload: payloadBytes,
	}

	// Send command
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(cmd); err != nil {
		return nil, err
	}

	// Read response
	var resp Response
	decoder := json.NewDecoder(conn)
	if err := decoder.Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// Command represents a command sent to the daemon.
type Command struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Response represents a response from the daemon.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}
