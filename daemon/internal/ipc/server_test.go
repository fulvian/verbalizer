package ipc

import (
	"encoding/json"
	"net"
	"os"
	"testing"
	"time"

	"github.com/fulvian/verbalizer/daemon/pkg/api"
)

type mockHandler struct {
	status api.StatusData
}

func (h *mockHandler) HandleCommand(cmd *api.Command) (*api.Response, error) {
	switch cmd.Type {
	case api.CmdGetStatus:
		return &api.Response{Success: true, Data: h.status}, nil
	case api.CmdStartRecording:
		h.status.IsRecording = true
		return &api.Response{Success: true}, nil
	case api.CmdStopRecording:
		h.status.IsRecording = false
		return &api.Response{Success: true}, nil
	default:
		return &api.Response{Success: false, Error: "Unknown command"}, nil
	}
}

func TestServer(t *testing.T) {
	socketPath := "/tmp/verbalizer_test.sock"
	defer os.Remove(socketPath)

	handler := &mockHandler{}
	server := NewServer(socketPath, handler)

	err := server.Start()
	if err != nil {
		t.Fatalf("Failed to start server: %v", err)
	}
	defer server.Stop()

	// Connect to server
	conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
	if err != nil {
		t.Fatalf("Failed to connect to server: %v", err)
	}
	defer conn.Close()

	// Send GET_STATUS command
	cmd := api.Command{Type: api.CmdGetStatus}
	if err := json.NewEncoder(conn).Encode(cmd); err != nil {
		t.Fatalf("Failed to encode command: %v", err)
	}

	// Read response
	var resp api.Response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success response, got error: %s", resp.Error)
	}

	// Verify data
	statusBytes, _ := json.Marshal(resp.Data)
	var status api.StatusData
	json.Unmarshal(statusBytes, &status)
	if status.IsRecording {
		t.Errorf("Expected IsRecording to be false")
	}
}
