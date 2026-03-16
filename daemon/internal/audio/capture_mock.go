package audio

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// MockCapture is a mock implementation of the Capture interface for testing.
type MockCapture struct {
	mu          sync.Mutex
	isCapturing bool
	mp3Path     string
}

func NewMockCapture() *MockCapture {
	return &MockCapture{}
}

func (c *MockCapture) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isCapturing {
		return fmt.Errorf("already capturing")
	}

	c.isCapturing = true
	recordingsDir := "recordings"
	_ = EnsureDir(recordingsDir)
	c.mp3Path = filepath.Join(recordingsDir, "mock_recording.mp3")
	
	// Create a dummy mp3 file
	_ = os.WriteFile(c.mp3Path, []byte("mock audio data"), 0644)
	
	return nil
}

func (c *MockCapture) Stop() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isCapturing {
		return "", fmt.Errorf("not capturing")
	}

	c.isCapturing = false
	return c.mp3Path, nil
}

func (c *MockCapture) IsCapturing() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isCapturing
}
