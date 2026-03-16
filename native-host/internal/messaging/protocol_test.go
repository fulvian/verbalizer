package messaging

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"testing"
)

func TestReadWriteMessage(t *testing.T) {
	msg := &Response{
		Success: true,
		Data:    "test",
	}

	buf := new(bytes.Buffer)
	if err := WriteMessage(buf, msg); err != nil {
		t.Fatalf("WriteMessage failed: %v", err)
	}

	// Verify length prefix
	var length uint32
	if err := binary.Read(bytes.NewReader(buf.Bytes()), binary.LittleEndian, &length); err != nil {
		t.Fatal(err)
	}

	if int(length) != buf.Len()-4 {
		t.Errorf("Expected length %d, got %d", buf.Len()-4, length)
	}

	// Test ReadMessage
	_, err := ReadMessage(buf)
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	inputMsg := &Message{
		Type:    "TEST",
		Payload: json.RawMessage(`{"foo":"bar"}`),
	}

	buf2 := new(bytes.Buffer)
	data, _ := json.Marshal(inputMsg)
	binary.Write(buf2, binary.LittleEndian, uint32(len(data)))
	buf2.Write(data)

	readMsg2, err := ReadMessage(buf2)
	if err != nil {
		t.Fatalf("ReadMessage failed: %v", err)
	}

	if readMsg2.Type != "TEST" {
		t.Errorf("Expected type TEST, got %s", readMsg2.Type)
	}
}
