// Package messaging implements Chrome Native Messaging protocol.
package messaging

import (
	"encoding/binary"
	"encoding/json"
	"io"
)

// Message represents a Native Messaging message from the extension.
type Message struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Response represents a Native Messaging response sent back to the extension.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ReadMessage reads a Native Messaging message from the reader.
func ReadMessage(r io.Reader) (*Message, error) {
	// Read 4-byte length prefix (little-endian)
	var length uint32
	if err := binary.Read(r, binary.LittleEndian, &length); err != nil {
		return nil, err
	}

	// Read message body
	data := make([]byte, length)
	if _, err := io.ReadFull(r, data); err != nil {
		return nil, err
	}

	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

// WriteMessage writes a Native Messaging response to the writer.
func WriteMessage(w io.Writer, resp *Response) error {
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}

	// Write 4-byte length prefix (little-endian)
	length := uint32(len(data))
	if err := binary.Write(w, binary.LittleEndian, &length); err != nil {
		return err
	}

	// Write message body
	_, err = w.Write(data)
	return err
}
