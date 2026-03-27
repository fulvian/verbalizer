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

// LinuxCapture captures system audio on Linux using PulseAudio/PipeWire
type LinuxCapture struct {
	mu          sync.Mutex
	isCapturing bool
	cmd         *exec.Cmd
	pcmPath     string
	mp3Path     string
	encoder     *Encoder
	sourceName  string
}

// NewLinuxCapture creates a new Linux audio capture instance
func NewLinuxCapture() (*LinuxCapture, error) {
	encoder, err := NewEncoder()
	if err != nil {
		return nil, err
	}
	return &LinuxCapture{
		encoder: encoder,
	}, nil
}

// Start begins capturing audio from the specified source
func (c *LinuxCapture) Start() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isCapturing {
		return fmt.Errorf("already capturing")
	}

	// Discover and select appropriate audio source
	sourceName, err := c.discoverAndSelectSource()
	if err != nil {
		return fmt.Errorf("source discovery failed: %w", err)
	}
	c.sourceName = sourceName

	return c.startCapture(sourceName)
}

// startCapture starts the actual capture with the given source
func (c *LinuxCapture) startCapture(sourceName string) error {
	recordingsDir := "recordings"
	if err := EnsureDir(recordingsDir); err != nil {
		return fmt.Errorf("failed to create recordings directory: %w", err)
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	c.mp3Path = filepath.Join(recordingsDir, fmt.Sprintf("%s_linux.mp3", timestamp))
	c.pcmPath = filepath.Join(recordingsDir, fmt.Sprintf("%s_linux.pcm", timestamp))

	// Get the ffmpeg input string for the source
	ffmpegSource := GetSourceForFFmpeg(sourceName)

	// Capture system audio from the selected source
	// -f pulse -i <source> captures from PulseAudio
	// For PipeWire, we can also use -f pipewire
	c.cmd = exec.Command("ffmpeg",
		"-f", "pulse",
		"-i", ffmpegSource,
		"-f", "s16le",
		"-ar", "44100",
		"-ac", "2",
		c.pcmPath,
	)

	// Capture stderr for debugging
	stderr, err := c.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to capture stderr: %w", err)
	}

	if err := c.cmd.Start(); err != nil {
		// Read stderr for better error message
		buf := make([]byte, 1024)
		n, _ := stderr.Read(buf)
		stderrMsg := string(buf[:n])
		if stderrMsg != "" {
			return fmt.Errorf("failed to start capture from source %q: %s (stderr: %s)", sourceName, err, stderrMsg)
		}
		return fmt.Errorf("failed to start capture from source %q: %w", sourceName, err)
	}

	c.isCapturing = true
	return nil
}

// discoverAndSelectSource finds the best audio source for capturing call audio
func (c *LinuxCapture) discoverAndSelectSource() (string, error) {
	sd := NewSourceDiscovery()

	// First, try to find a monitor source (system audio)
	monitorSource, err := sd.FindMonitorSource()
	fmt.Printf("[DEBUG] discoverAndSelectSource: FindMonitorSource returned: source=%q, err=%v\n", monitorSource, err)
	if err == nil && monitorSource != "" {
		// Validate the source exists
		valid, valErr := sd.ValidateSource(monitorSource)
		fmt.Printf("[DEBUG] discoverAndSelectSource: ValidateSource(%q) returned: valid=%v, err=%v\n", monitorSource, valid, valErr)
		if valErr == nil && valid {
			return monitorSource, nil
		}
	} else {
		fmt.Printf("[DEBUG] discoverAndSelectSource: FindMonitorSource failed: %v\n", err)
	}

	// Fallback: try to get default source
	defaultSource, err := sd.GetDefaultSource()
	fmt.Printf("[DEBUG] discoverAndSelectSource: GetDefaultSource returned: source=%q, err=%v\n", defaultSource, err)
	if err == nil && defaultSource != "" {
		valid, valErr := sd.ValidateSource(defaultSource)
		fmt.Printf("[DEBUG] discoverAndSelectSource: ValidateSource(%q) returned: valid=%v, err=%v\n", defaultSource, valid, valErr)
		if valErr == nil && valid {
			fmt.Printf("[DEBUG] discoverAndSelectSource: falling back to defaultSource=%q\n", defaultSource)
			return defaultSource, nil
		}
	}

	// Last resort: use "default" as the source name
	// ffmpeg will use the system's default source
	fmt.Printf("[DEBUG] discoverAndSelectSource: all fallbacks failed, using 'default'\n")
	return "default", nil
}

// Stop stops the recording and returns the path to the output file
func (c *LinuxCapture) Stop() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.isCapturing {
		return "", fmt.Errorf("not capturing")
	}

	if c.cmd != nil && c.cmd.Process != nil {
		// First try graceful termination
		err := c.cmd.Process.Signal(os.Interrupt)
		if err != nil {
			// If graceful fails, force kill after a brief wait
			time.Sleep(100 * time.Millisecond)
			_ = c.cmd.Process.Kill()
		}

		// Wait for process to exit
		_ = c.cmd.Wait()
	}

	c.isCapturing = false

	// Check if PCM file was created and has content
	if c.pcmPath != "" {
		info, err := os.Stat(c.pcmPath)
		if err != nil {
			return "", fmt.Errorf("PCM file not created: %w", err)
		}
		if info.Size() == 0 {
			return "", fmt.Errorf("recording produced empty file (no audio captured from source %q)", c.sourceName)
		}
	}

	// Encode PCM to MP3
	if err := c.encoder.EncodePCMToMP3(c.pcmPath, c.mp3Path); err != nil {
		return "", fmt.Errorf("encoding failed: %w", err)
	}

	// Verify the output file exists and has content
	if c.mp3Path != "" {
		info, err := os.Stat(c.mp3Path)
		if err != nil {
			return "", fmt.Errorf("MP3 file not created: %w", err)
		}
		if info.Size() == 0 {
			return "", fmt.Errorf("encoding produced empty file")
		}
	}

	// Cleanup PCM file
	_ = os.Remove(c.pcmPath)

	return c.mp3Path, nil
}

// IsCapturing returns true if currently capturing audio
func (c *LinuxCapture) IsCapturing() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.isCapturing
}

// GetSourceName returns the name of the audio source being used
func (c *LinuxCapture) GetSourceName() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.sourceName
}

// ValidateSource checks if the specified source is available
func ValidateSource(sourceName string) (bool, error) {
	sd := NewSourceDiscovery()
	return sd.ValidateSource(sourceName)
}

// ListAvailableSources returns all available audio sources
func ListAvailableSources() ([]AudioSource, error) {
	sd := NewSourceDiscovery()
	return sd.DiscoverSources()
}

// PreflightCheck performs validation before starting capture
func PreflightCheck() (bool, string, error) {
	// Check ffmpeg is available
	cmd := exec.Command("ffmpeg", "-version")
	if err := cmd.Run(); err != nil {
		return false, "", fmt.Errorf("ffmpeg not available: %w", err)
	}

	// Check we have audio sources
	sd := NewSourceDiscovery()
	sources, err := sd.DiscoverSources()
	if err != nil {
		return false, fmt.Sprintf("source discovery failed: %v", err), nil
	}

	if len(sources) == 0 {
		return false, "no audio sources found", nil
	}

	// Find a valid source
	monitorSource, _ := sd.FindMonitorSource()
	if monitorSource == "" {
		// Use default source
		monitorSource, _ = sd.GetDefaultSource()
	}

	return true, fmt.Sprintf("using source: %s (found %d sources)", monitorSource, len(sources)), nil
}
