// Package logger provides structured logging for the Verbalizer daemon.
package logger

import "strings"

// ErrorCategory represents the category of an error.
type ErrorCategory int

const (
	CategoryUnknown ErrorCategory = iota
	CategoryNetwork
	CategoryAuth
	CategoryAPI
	CategoryStorage
	CategoryConfig
	CategorySystem
)

func (ec ErrorCategory) String() string {
	switch ec {
	case CategoryNetwork:
		return "network"
	case CategoryAuth:
		return "auth"
	case CategoryAPI:
		return "api"
	case CategoryStorage:
		return "storage"
	case CategoryConfig:
		return "config"
	case CategorySystem:
		return "system"
	default:
		return "unknown"
	}
}

// IsRetryable returns true if errors of this category are typically retryable.
func (ec ErrorCategory) IsRetryable() bool {
	switch ec {
	case CategoryNetwork:
		return true
	case CategoryAPI:
		return true // 5xx and 429 are retryable
	default:
		return false
	}
}

// CategorizeError determines the category of an error.
func CategorizeError(err error) ErrorCategory {
	if err == nil {
		return CategoryUnknown
	}

	errStr := strings.ToLower(err.Error())

	// Network errors
	networkPatterns := []string{
		"connection refused",
		"connection reset",
		"connection timed out",
		"no such host",
		"network",
		"timeout",
		"i/o timeout",
		"temporary failure",
		"econnrefused",
		"econnreset",
		"etimedout",
		"network unreachable",
	}
	for _, p := range networkPatterns {
		if strings.Contains(errStr, p) {
			return CategoryNetwork
		}
	}

	// Auth errors
	authPatterns := []string{
		"unauthorized",
		"forbidden",
		"invalid_grant",
		"access_denied",
		"token",
		"credential",
		"authentication",
		"oauth",
		"revoked",
	}
	for _, p := range authPatterns {
		if strings.Contains(errStr, p) {
			return CategoryAuth
		}
	}

	// API errors (HTTP 4xx/5xx)
	apiPatterns := []string{
		"status 400",
		"status 401",
		"status 403",
		"status 404",
		"status 429",
		"status 500",
		"status 502",
		"status 503",
		"status 504",
	}
	for _, p := range apiPatterns {
		if strings.Contains(errStr, p) {
			return CategoryAPI
		}
	}

	// Storage errors
	storagePatterns := []string{
		"open",
		"read",
		"write",
		"file",
		"directory",
		"disk",
		"space",
		"permission",
	}
	for _, p := range storagePatterns {
		if strings.Contains(errStr, p) {
			return CategoryStorage
		}
	}

	// Config errors
	configPatterns := []string{
		"config",
		"yaml",
		"invalid",
		"missing",
	}
	for _, p := range configPatterns {
		if strings.Contains(errStr, p) {
			return CategoryConfig
		}
	}

	return CategorySystem
}

// UploadError wraps an error with upload context.
type UploadError struct {
	CallID     string
	Err        error
	Category   ErrorCategory
	Retryable  bool
	StatusCode int
}

func (e *UploadError) Error() string {
	return e.Err.Error()
}

func (e *UploadError) Unwrap() error {
	return e.Err
}

// NewUploadError creates a new upload error.
func NewUploadError(callID string, err error, statusCode int) *UploadError {
	cat := CategorizeError(err)
	retryable := cat.IsRetryable() && isRetryableHTTPStatus(statusCode)
	return &UploadError{
		CallID:     callID,
		Err:        err,
		Category:   cat,
		Retryable:  retryable,
		StatusCode: statusCode,
	}
}

func isRetryableHTTPStatus(code int) bool {
	// 5xx errors and 429 are retryable
	return code >= 500 || code == 429
}
