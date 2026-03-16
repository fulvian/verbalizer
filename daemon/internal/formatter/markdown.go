// Package formatter provides transcript formatting.
package formatter

import (
	"os"
	"path/filepath"
	"text/template"
	"time"
)

// MarkdownFormatter generates Markdown transcripts.
type MarkdownFormatter struct {
	template *template.Template
}

// TranscriptData holds data for the transcript template.
type TranscriptData struct {
	Title        string
	Date         string
	Platform     string
	Duration     string
	AudioFile    string
	Participants int
	Transcript   string
}

// NewMarkdownFormatter creates a new Markdown formatter.
func NewMarkdownFormatter() *MarkdownFormatter {
	tmpl := template.Must(template.New("transcript").Parse(`---
title: "{{.Title}}"
date: {{.Date}}
platform: {{.Platform}}
duration: {{.Duration}}
audio_file: {{.AudioFile}}
participants: {{.Participants}}
---

# Transcript

{{.Transcript}}
`))

	return &MarkdownFormatter{template: tmpl}
}

// Format generates a Markdown file from transcript data.
func (f *MarkdownFormatter) Format(data TranscriptData, outputPath string) error {
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()

	return f.template.Execute(file, data)
}

// FormatDuration converts seconds to MM:SS format.
func FormatDuration(seconds int) string {
	mins := seconds / 60
	secs := seconds % 60
	return string(rune(mins)) + ":" + string(rune(secs))
}

// FormatTimestamp generates an ISO timestamp.
func FormatTimestamp(t time.Time) string {
	return t.Format(time.RFC3339)
}
