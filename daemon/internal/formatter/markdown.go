// Package formatter provides transcript formatting.
package formatter

import (
	"fmt"
	"strings"
	"text/template"
	"time"
)

const markdownTemplate = `---
title: "{{.Metadata.Title}}"
date: {{formatDate .Metadata.Date}}
platform: {{.Metadata.Platform}}
duration: {{formatDuration .Metadata.Duration}}
audio_file: {{.Metadata.AudioFile}}
---

# {{.Metadata.Title}}
{{if .Segments}}{{range .Segments}}
[{{formatDuration .Start}} - {{formatDuration .End}}] {{.Text}}
{{end}}{{else}}
{{.FullText}}
{{end}}
`

// MarkdownFormatter generates Markdown transcripts.
type MarkdownFormatter struct {
	tmpl *template.Template
}

// NewMarkdownFormatter creates a new Markdown formatter.
func NewMarkdownFormatter() *MarkdownFormatter {
	funcMap := template.FuncMap{
		"formatDate": func(t time.Time) string {
			return t.Format("2006-01-02 15:04:05")
		},
		"formatDuration": func(d time.Duration) string {
			h := int(d.Hours())
			m := int(d.Minutes()) % 60
			s := int(d.Seconds()) % 60
			if h > 0 {
				return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
			}
			return fmt.Sprintf("%02d:%02d", m, s)
		},
	}

	tmpl := template.Must(template.New("markdown").Funcs(funcMap).Parse(markdownTemplate))
	return &MarkdownFormatter{tmpl: tmpl}
}

// Format generates a Markdown string from transcript data.
func (f *MarkdownFormatter) Format(data TranscriptData) (string, error) {
	var buf strings.Builder
	if err := f.tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.String(), nil
}
