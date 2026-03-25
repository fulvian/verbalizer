package transcriber

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
)

// Whisper implements the Transcriber interface using whisper.cpp.
type Whisper struct {
	binaryPath string
	modelPath  string
}

// execCommand is used for mocking in tests.
var execCommand = exec.Command

// NewWhisper creates a new Whisper transcriber.
func NewWhisper(binaryPath, modelPath string) *Whisper {
	return &Whisper{
		binaryPath: binaryPath,
		modelPath:  modelPath,
	}
}

// Transcribe transcribes an audio file using whisper.cpp.
func (w *Whisper) Transcribe(audioPath string) (*Transcript, error) {
	if _, err := os.Stat(w.binaryPath); err != nil {
		return nil, fmt.Errorf("whisper.cpp binary not found at %s: %w", w.binaryPath, err)
	}

	if _, err := os.Stat(w.modelPath); err != nil {
		return nil, fmt.Errorf("whisper model not found at %s: %w", w.modelPath, err)
	}

	// 1. Convert to 16kHz WAV (required by whisper.cpp)
	wavPath := audioPath + ".16kHz.wav"
	// Use ffmpeg for conversion
	convCmd := execCommand("ffmpeg", "-y", "-i", audioPath, "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", wavPath)
	if output, err := convCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to convert audio to 16kHz WAV: %w, output: %s", err, string(output))
	}
	defer os.Remove(wavPath)

	// 2. Run whisper.cpp
	// -f: input file
	// -m: model path
	// -oj: output JSON
	// -l it: force Italian language
	// whisper.cpp will create wavPath.json
	cmd := execCommand(w.binaryPath, "-m", w.modelPath, "-f", wavPath, "-l", "it", "-oj")
	if output, err := cmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("whisper.cpp failed: %w, output: %s", err, string(output))
	}

	jsonPath := wavPath + ".json"
	defer os.Remove(jsonPath)

	// 3. Read and parse JSON
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read whisper.cpp output: %w", err)
	}

	var whisperOutput struct {
		Result struct {
			Language string `json:"language"`
		} `json:"result"`
		Transcription []struct {
			Offsets struct {
				From int64 `json:"from"`
				To   int64 `json:"to"`
			} `json:"offsets"`
			Text string `json:"text"`
		} `json:"transcription"`
	}

	if err := json.Unmarshal(data, &whisperOutput); err != nil {
		return nil, fmt.Errorf("failed to parse whisper.cpp JSON: %w", err)
	}

	// 4. Map to Transcript
	transcript := &Transcript{
		Language: whisperOutput.Result.Language,
		Text:     "",
	}

	for i, seg := range whisperOutput.Transcription {
		transcript.Segments = append(transcript.Segments, Segment{
			Start: float64(seg.Offsets.From) / 1000.0,
			End:   float64(seg.Offsets.To) / 1000.0,
			Text:  seg.Text,
		})
		if i > 0 && len(transcript.Text) > 0 && transcript.Text[len(transcript.Text)-1] != ' ' && len(seg.Text) > 0 && seg.Text[0] != ' ' {
			transcript.Text += " "
		}
		transcript.Text += seg.Text
	}

	return transcript, nil
}
