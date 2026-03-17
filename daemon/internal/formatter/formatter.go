package formatter

// Formatter is the interface for transcript formatters.
type Formatter interface {
	Format(data TranscriptData) (string, error)
}
