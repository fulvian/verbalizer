//go:build !linux

package audio

import "fmt"

func NewLinuxCapture() (Capture, error) {
	return nil, fmt.Errorf("linux capture not supported on this platform")
}
