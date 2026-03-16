// Package config provides configuration management for the Verbalizer daemon.
package config

import (
	"os"
	"path/filepath"
)

// Config holds the daemon configuration.
type Config struct {
	DataDir        string `yaml:"data_dir"`
	RecordingsDir  string `yaml:"recordings_dir"`
	TranscriptsDir string `yaml:"transcripts_dir"`

	Audio         AudioConfig         `yaml:"audio"`
	Transcription TranscriptionConfig `yaml:"transcription"`
	Logging       LoggingConfig       `yaml:"logging"`
}

// AudioConfig holds audio-related settings.
type AudioConfig struct {
	Format     string `yaml:"format"`
	Bitrate    string `yaml:"bitrate"`
	SampleRate int    `yaml:"sample_rate"`
}

// TranscriptionConfig holds transcription-related settings.
type TranscriptionConfig struct {
	Model    string `yaml:"model"`
	Language string `yaml:"language"`
	Threads  int    `yaml:"threads"`
}

// LoggingConfig holds logging-related settings.
type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// DefaultConfig returns the default configuration.
func DefaultConfig() *Config {
	homeDir, _ := os.UserHomeDir()
	dataDir := filepath.Join(homeDir, "verbalizer")

	return &Config{
		DataDir:        dataDir,
		RecordingsDir:  filepath.Join(dataDir, "recordings"),
		TranscriptsDir: filepath.Join(dataDir, "transcripts"),
		Audio: AudioConfig{
			Format:     "mp3",
			Bitrate:    "128k",
			SampleRate: 16000,
		},
		Transcription: TranscriptionConfig{
			Model:    "small",
			Language: "auto",
			Threads:  4,
		},
		Logging: LoggingConfig{
			Level: "info",
		},
	}
}

// Load loads configuration from a YAML file.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	// TODO: Implement YAML file loading

	return cfg, nil
}

// EnsureDirs creates necessary directories if they don't exist.
func (c *Config) EnsureDirs() error {
	dirs := []string{
		c.DataDir,
		c.RecordingsDir,
		c.TranscriptsDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}
