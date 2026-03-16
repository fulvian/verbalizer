// Package audio provides audio capture functionality.
package audio

// Capture defines the interface for audio capture implementations.
type Capture interface {
	// Start begins audio capture.
	Start() error

	// Stop ends audio capture and returns the captured audio file path.
	Stop() (string, error)

	// IsCapturing returns whether audio is currently being captured.
	IsCapturing() bool
}
