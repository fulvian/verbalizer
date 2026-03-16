package ipc

import (
	"encoding/json"
	"net"
	"os"
	"testing"
)

func TestClient_Send(t *testing.T) {
	socketPath := "/tmp/verbalizer_test.sock"
	defer os.Remove(socketPath)

	// Mock server
	l, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatal(err)
	}
	defer l.Close()

	go func() {
		conn, err := l.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		var cmd Command
		decoder := json.NewDecoder(conn)
		if err := decoder.Decode(&cmd); err != nil {
			return
		}

		resp := Response{
			Success: true,
			Data:    json.RawMessage(`{"status": "ok"}`),
		}
		encoder := json.NewEncoder(conn)
		encoder.Encode(resp)
	}()

	client := NewClient(socketPath)
	resp, err := client.Send("START_RECORDING", map[string]string{"callId": "123"})
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	if !resp.Success {
		t.Error("Expected success true")
	}

	var data map[string]string
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatal(err)
	}

	if data["status"] != "ok" {
		t.Errorf("Expected status ok, got %s", data["status"])
	}
}
