//go:build linux

package audio

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"
)

type LinuxCapture struct {
	mu          sync.Mutex
	isCapturing bool
	cmd         *exec.Cmd
	pcmPath     string
	mp3Path     string
	encoder     *Encoder
}

func NewLinuxCapture() (*LinuxCapture, error) {
	encoder, err := NewEncoder()
	if err != nil {
		return nil, err
	}
	return &LinuxCapture{
		encoder: encoder,
	}, nil
}

func (c *LinuxCapture) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isCapturing {
		return fmt.Errorf("already capturing")
	}

	recordingsDir := "recordings"
	if err := EnsureDir(recordingsDir); err != nil {
		return err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	c.mp3Path = filepath.Join(recordingsDir, fmt.Sprintf("%s_linux.mp3", timestamp))
	c.pcmPath = filepath.Join(recordingsDir, fmt.Sprintf("%s_linux.pcm", timestamp))

	// Using pw-record to capture audio. 
	// Note: Capturing from Chrome specifically might require identifying the correct monitor source.
	// For now, we capture default monitor which usually contains Chrome audio.
	// ffmpeg can also be used: ffmpeg -f pulse -i default out.pcm
	c.cmd = exec.Command("pw-record", "--format=s16", "--rate=44100", "--channels=2", c.pcmPath)

	if err := c.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start pw-record: %w", err)
	}

	c.isCapturing = true
	return nil
}

func (c *LinuxCapture) Stop() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isCapturing {
		return "", fmt.Errorf("not capturing")
	}

	if c.cmd != nil && c.cmd.Process != nil {
		// Use Interrupt instead of Kill to allow graceful exit if possible, 
		// but pw-record might need Kill.
		_ = c.cmd.Process.Signal(os.Interrupt)
		
		// Wait for process to exit
		_ = c.cmd.Wait()
	}

	c.isCapturing = false

	// Encode PCM to MP3
	if err := c.encoder.EncodePCMToMP3(c.pcmPath, c.mp3Path); err != nil {
		return "", fmt.Errorf("encoding failed: %w", err)
	}

	// Cleanup PCM file
	_ = os.Remove(c.pcmPath)

	return c.mp3Path, nil
}

func (c *LinuxCapture) IsCapturing() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isCapturing
}
