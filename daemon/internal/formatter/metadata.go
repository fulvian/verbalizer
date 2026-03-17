package formatter

import "time"

// Metadata represents the meeting metadata.
type Metadata struct {
	Title     string
	Date      time.Time
	Platform  string
	Duration  time.Duration
	AudioFile string
}

// TranscriptSegment represents a single segment of the transcript.
type TranscriptSegment struct {
	Start time.Duration
	End   time.Duration
	Text  string
}

// TranscriptData holds the full transcript data including metadata and segments.
type TranscriptData struct {
	Metadata Metadata
	Segments []TranscriptSegment
	FullText string
}
