// Package config provides configuration management for the Verbalizer daemon.
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config holds the daemon configuration.
type Config struct {
	DataDir        string `yaml:"data_dir"`
	RecordingsDir  string `yaml:"recordings_dir"`
	TranscriptsDir string `yaml:"transcripts_dir"`
	DBPath         string `yaml:"db_path"`

	Audio         AudioConfig         `yaml:"audio"`
	Transcription TranscriptionConfig `yaml:"transcription"`
	Logging       LoggingConfig       `yaml:"logging"`
	Cloud         CloudConfig         `yaml:"cloud"`
}

// CloudConfig holds cloud synchronization settings.
type CloudConfig struct {
	Enabled           bool        `yaml:"enabled"`
	Provider          string      `yaml:"provider"` // "google_drive"
	OAuthClientID     string      `yaml:"oauth_client_id"`
	OAuthRedirectHost string      `yaml:"oauth_redirect_host"`
	OAuthRedirectPort string      `yaml:"oauth_redirect_port_range"`
	Scope             string      `yaml:"scope"`
	TargetFolderID    string      `yaml:"target_folder_id"`
	UploadMode        string      `yaml:"upload_mode"` // "multipart" or "resumable"
	Retry             RetryConfig `yaml:"retry"`
}

// RetryConfig holds retry policy settings.
type RetryConfig struct {
	MaxAttempts      int `yaml:"max_attempts"`
	BaseDelaySeconds int `yaml:"base_delay_seconds"`
	MaxDelaySeconds  int `yaml:"max_delay_seconds"`
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
		DBPath:         filepath.Join(dataDir, "verbalizer.db"),
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
		Cloud: CloudConfig{
			Enabled:           false,
			Provider:          "google_drive",
			OAuthRedirectHost: "127.0.0.1",
			OAuthRedirectPort: "49152-65535",
			Scope:             "https://www.googleapis.com/auth/drive.file",
			UploadMode:        "multipart",
			Retry: RetryConfig{
				MaxAttempts:      20,
				BaseDelaySeconds: 30,
				MaxDelaySeconds:  7200,
			},
		},
	}
}

// Load loads configuration from a YAML file.
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, return defaults
			return cfg, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Ensure directories use defaults if not specified
	if cfg.DataDir == "" {
		homeDir, _ := os.UserHomeDir()
		cfg.DataDir = filepath.Join(homeDir, "verbalizer")
	}
	if cfg.RecordingsDir == "" {
		cfg.RecordingsDir = filepath.Join(cfg.DataDir, "recordings")
	}
	if cfg.TranscriptsDir == "" {
		cfg.TranscriptsDir = filepath.Join(cfg.DataDir, "transcripts")
	}
	if cfg.DBPath == "" {
		cfg.DBPath = filepath.Join(cfg.DataDir, "verbalizer.db")
	}

	return cfg, nil
}

// LoadFromDataDir loads configuration from the default data directory.
func LoadFromDataDir(dataDir string) (*Config, error) {
	configPath := filepath.Join(dataDir, "config.yaml")
	return Load(configPath)
}

// Save saves configuration to a YAML file.
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
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

// IsCloudEnabled returns true if cloud sync is enabled and configured.
func (c *Config) IsCloudEnabled() bool {
	return c.Cloud.Enabled && c.Cloud.Provider == "google_drive"
}
