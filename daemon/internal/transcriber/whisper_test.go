package transcriber

import (
	"fmt"
	"os"
	"os/exec"
	"testing"
)

// mockExecCommand is a helper for mocking exec.Command.
func mockExecCommand(command string, args ...string) *exec.Cmd {
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestWhisper_Transcribe(t *testing.T) {
	// Save original execCommand and restore after test
	oldExec := execCommand
	defer func() { execCommand = oldExec }()
	execCommand = mockExecCommand

	// Create dummy binary and model files
	tmpDir := t.TempDir()
	binaryPath := tmpDir + "/whisper"
	modelPath := tmpDir + "/model.bin"
	audioPath := tmpDir + "/audio.wav"

	os.Create(binaryPath)
	os.Create(modelPath)
	os.Create(audioPath)
	os.Chmod(binaryPath, 0755)

	w := NewWhisper(binaryPath, modelPath)

	transcript, err := w.Transcribe(audioPath)
	if err != nil {
		t.Fatalf("Transcribe failed: %v", err)
	}

	if transcript.Text != " Hello world This is a test" {
		t.Errorf("Expected transcript text ' Hello world This is a test', got '%s'", transcript.Text)
	}

	if transcript.Language != "en" {
		t.Errorf("Expected language 'en', got '%s'", transcript.Language)
	}

	if len(transcript.Segments) != 2 {
		t.Errorf("Expected 2 segments, got %d", len(transcript.Segments))
	}
}

// TestHelperProcess is used to mock the execution of external commands.
func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	defer os.Exit(0)

	args := os.Args
	for i := 0; i < len(args); i++ {
		if args[i] == "--" {
			args = args[i+1:]
			break
		}
	}

	if len(args) == 0 {
		fmt.Fprintf(os.Stderr, "No command specified\n")
		os.Exit(1)
	}

	cmd := args[0]
	switch cmd {
	case "ffmpeg":
		// Create the output .wav file
		wavPath := ""
		for i, arg := range args {
			if arg == "-ac" && i+2 < len(args) {
				// The output file is usually the last argument or after some flags
			}
		}
		// In whisper.go: convCmd := execCommand("ffmpeg", "-y", "-i", audioPath, "-ar", "16000", "-ac", "1", "-c:a", "pcm_s16le", wavPath)
		wavPath = args[len(args)-1]
		os.Create(wavPath)
	default:
		// Check if it's the whisper binary (it will have the tmpDir path)
		// whisper.go: cmd := execCommand(w.binaryPath, "-m", w.modelPath, "-f", wavPath, "-oj")
		// It will create wavPath.json
		for i, arg := range args {
			if arg == "-f" && i+1 < len(args) {
				wavPath := args[i+1]
				jsonPath := wavPath + ".json"
				jsonContent := `{
					"result": {
						"language": "en"
					},
					"transcription": [
						{
							"offsets": { "from": 0, "to": 1000 },
							"text": " Hello world"
						},
						{
							"offsets": { "from": 1000, "to": 2000 },
							"text": " This is a test"
						}
					]
				}`
				os.WriteFile(jsonPath, []byte(jsonContent), 0644)
			}
		}
	}
}
