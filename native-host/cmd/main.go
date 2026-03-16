// Package main provides the Chrome Native Messaging Host for Verbalizer.
// This binary communicates with the Chrome extension via stdin/stdout
// using the Chrome Native Messaging protocol, and forwards messages
// to the Verbalizer daemon via Unix socket.
package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Message represents a message from the Chrome extension.
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Response represents a response sent back to the extension.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Read messages from stdin in Native Messaging format
	// Each message is prefixed with a 4-byte length (little-endian)
	for {
		msg, err := readMessage(os.Stdin)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("failed to read message: %w", err)
		}

		response := handleMessage(msg)
		if err := writeMessage(os.Stdout, response); err != nil {
			return fmt.Errorf("failed to write response: %w", err)
		}
	}
}

// readMessage reads a Native Messaging protocol message from the reader.
func readMessage(r io.Reader) (*Message, error) {
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to parse message: %w", err)
	}

	return &msg, nil
}

// writeMessage writes a response in Native Messaging protocol format.
func writeMessage(w io.Writer, resp *Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	length := uint32(len(data))
	if err := binary.Write(w, binary.LittleEndian, length); err != nil {
		return err
	}

	_, err = w.Write(data)
	return err
}

// handleMessage processes incoming messages from the extension.
// TODO: Implement IPC communication with the daemon.
func handleMessage(msg *Message) *Response {
	switch msg.Type {
	case "ping":
		return &Response{Success: true, Data: "pong"}
	case "startRecording":
		// TODO: Forward to daemon via Unix socket
		return &Response{Success: true, Data: "recording_started"}
	case "stopRecording":
		// TODO: Forward to daemon via Unix socket
		return &Response{Success: true, Data: "recording_stopped"}
	default:
		return &Response{Success: false, Error: fmt.Sprintf("unknown message type: %s", msg.Type)}
	}
}
