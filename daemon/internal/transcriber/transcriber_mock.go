package transcriber

import "sync"

// MockTranscriber is a mock implementation of the Transcriber interface for testing.
type MockTranscriber struct {
	mu              sync.Mutex
	TranscribeCount int
	LastAudioPath   string
	Result          *Transcript
	Err             error
}

func NewMockTranscriber() *MockTranscriber {
	return &MockTranscriber{
		Result: &Transcript{
			Text:     "Mock transcription",
			Language: "en",
		},
	}
}

func (m *MockTranscriber) Transcribe(audioPath string) (*Transcript, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TranscribeCount++
	m.LastAudioPath = audioPath

	if m.Err != nil {
		return nil, m.Err
	}
	return m.Result, nil
}
