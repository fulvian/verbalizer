//go:build !darwin

package audio

import "fmt"

func NewDarwinCapture() (Capture, error) {
	return nil, fmt.Errorf("darwin capture not supported on this platform")
}
