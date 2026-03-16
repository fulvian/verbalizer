package audio

import (
	"os"
	"testing"
)

func TestEncoder(t *testing.T) {
	_, err := NewEncoder()
	if err != nil {
		t.Skip("ffmpeg not found, skipping encoder test")
	}

	t.Run("EnsureDir", func(t *testing.T) {
		dir := "test_recordings"
		err := EnsureDir(dir)
		if err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}
		defer os.RemoveAll(dir)

		if _, err := os.Stat(dir); os.IsNotExist(err) {
			t.Errorf("Directory was not created")
		}
	})
}
