package formatter

import (
	"strings"
	"testing"
	"time"
)

func TestMarkdownFormatter_Format(t *testing.T) {
	formatter := NewMarkdownFormatter()

	data := TranscriptData{
		Metadata: Metadata{
			Title:     "Daily Standup",
			Date:      time.Date(2023, 10, 15, 10, 0, 0, 0, time.UTC),
			Platform:  "Google Meet",
			Duration:  15 * time.Minute,
			AudioFile: "audio.wav",
		},
		FullText: "This is a full transcript text.",
	}

	markdown, err := formatter.Format(data)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	expected := `---
title: "Daily Standup"
date: 2023-10-15 10:00:00
platform: Google Meet
duration: 15:00
audio_file: audio.wav
---

# Daily Standup

This is a full transcript text.
`

	if strings.TrimSpace(markdown) != strings.TrimSpace(expected) {
		t.Errorf("Format output mismatch.\nExpected:\n%s\nGot:\n%s", expected, markdown)
	}
}

func TestMarkdownFormatter_FormatWithSegments(t *testing.T) {
	formatter := NewMarkdownFormatter()

	data := TranscriptData{
		Metadata: Metadata{
			Title:     "Daily Standup",
			Date:      time.Date(2023, 10, 15, 10, 0, 0, 0, time.UTC),
			Platform:  "Google Meet",
			Duration:  15 * time.Minute,
			AudioFile: "audio.wav",
		},
		Segments: []TranscriptSegment{
			{Start: 0, End: 10 * time.Second, Text: "Hello world."},
			{Start: 10 * time.Second, End: 20 * time.Second, Text: "Goodbye world."},
		},
	}

	markdown, err := formatter.Format(data)
	if err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	expected := `---
title: "Daily Standup"
date: 2023-10-15 10:00:00
platform: Google Meet
duration: 15:00
audio_file: audio.wav
---

# Daily Standup

[00:00 - 00:10] Hello world.

[00:10 - 00:20] Goodbye world.
`

	if strings.TrimSpace(markdown) != strings.TrimSpace(expected) {
		t.Errorf("Format output mismatch.\nExpected:\n%s\nGot:\n%s", expected, markdown)
	}
}
