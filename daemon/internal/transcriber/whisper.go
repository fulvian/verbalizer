// Package transcriber provides speech-to-text transcription.
package transcriber

// Transcriber defines the interface for transcription implementations.
type Transcriber interface {
	// Transcribe transcribes an audio file and returns the transcript.
	Transcribe(audioPath string) (*Transcript, error)
}

// Transcript represents a transcription result.
type Transcript struct {
	Text     string    `json:"text"`
	Segments []Segment `json:"segments"`
	Language string    `json:"language"`
	Duration float64   `json:"duration"`
}

// Segment represents a segment of transcribed text.
type Segment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}
