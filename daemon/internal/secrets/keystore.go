// Package secrets provides secure storage for credentials using OS keychain.
package secrets

import (
	"fmt"
	"os"
	"path/filepath"
)

// SecretStore defines the interface for secure credential storage.
type SecretStore interface {
	// Save saves a secret value.
	Save(key string, value []byte) error

	// Get retrieves a secret value.
	Get(key string) ([]byte, error)

	// Delete removes a secret.
	Delete(key string) error

	// Exists checks if a secret exists.
	Exists(key string) (bool, error)
}

// FileSecretStore implements SecretStore using encrypted files.
// This is a fallback for systems without keychain support.
type FileSecretStore struct {
	baseDir string
}

// NewFileSecretStore creates a new file-based secret store.
func NewFileSecretStore(baseDir string) (*FileSecretStore, error) {
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create secrets directory: %w", err)
	}
	return &FileSecretStore{baseDir: baseDir}, nil
}

func (s *FileSecretStore) secretPath(key string) string {
	return filepath.Join(s.baseDir, fmt.Sprintf("%s.secret", key))
}

// Save saves a secret value to a file.
func (s *FileSecretStore) Save(key string, value []byte) error {
	// Simple XOR encryption with machine-specific key
	// In production, use proper encryption like AES-GCM
	path := s.secretPath(key)

	// Encrypt using simple cipher (placeholder - use real encryption in production)
	encrypted := encryptSimple(value, getMachineKey())

	if err := os.WriteFile(path, encrypted, 0600); err != nil {
		return fmt.Errorf("failed to write secret: %w", err)
	}
	return nil
}

// Get retrieves a secret value from a file.
func (s *FileSecretStore) Get(key string) ([]byte, error) {
	path := s.secretPath(key)

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &NotFoundError{Key: key}
		}
		return nil, fmt.Errorf("failed to read secret: %w", err)
	}

	// Decrypt
	decrypted := decryptSimple(data, getMachineKey())
	return decrypted, nil
}

// Delete removes a secret.
func (s *FileSecretStore) Delete(key string) error {
	path := s.secretPath(key)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete secret: %w", err)
	}
	return nil
}

// Exists checks if a secret exists.
func (s *FileSecretStore) Exists(key string) (bool, error) {
	path := s.secretPath(key)
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// NotFoundError indicates a secret was not found.
type NotFoundError struct {
	Key string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("secret not found: %s", e.Key)
}

// simple XOR cipher for basic obfuscation
// In production, use crypto/aes or similar
func encryptSimple(data, key []byte) []byte {
	result := make([]byte, len(data))
	for i := range data {
		result[i] = data[i] ^ key[i%len(key)]
	}
	return result
}

func decryptSimple(data, key []byte) []byte {
	// XOR is symmetric
	return encryptSimple(data, key)
}

// getMachineKey returns a key derived from machine-specific data.
func getMachineKey() []byte {
	// Use user-specific and machine-specific data to derive key
	homeDir, _ := os.UserHomeDir()
	hostname, _ := os.Hostname()

	keySource := fmt.Sprintf("%s-%s-verbalizer", homeDir, hostname)
	return []byte(keySource)
}
