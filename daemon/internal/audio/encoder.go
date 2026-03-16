package audio

import (
	"fmt"
	"os"
	"os/exec"
)

// Encoder handles audio encoding using ffmpeg.
type Encoder struct {
	ffmpegPath string
}

// NewEncoder creates a new Encoder.
func NewEncoder() (*Encoder, error) {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg not found: %w", err)
	}
	return &Encoder{ffmpegPath: path}, nil
}

// EncodePCMToMP3 encodes a PCM file to MP3.
func (e *Encoder) EncodePCMToMP3(inputPath, outputPath string) error {
	// ffmpeg -f s16le -ar 44100 -ac 2 -i input.pcm -ab 128k output.mp3
	cmd := exec.Command(e.ffmpegPath,
		"-y", // overwrite output
		"-f", "s16le",
		"-ar", "44100",
		"-ac", "2",
		"-i", inputPath,
		"-ab", "128k",
		outputPath,
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w, output: %s", err, string(output))
	}

	return nil
}

// EnsureDir ensures the recordings directory exists.
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
