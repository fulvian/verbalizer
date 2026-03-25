package logger

import (
	"errors"
	"testing"
)

func TestCategorizeError_Network(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected ErrorCategory
	}{
		{"connection_refused", "connection refused", CategoryNetwork},
		{"connection_reset", "connection reset", CategoryNetwork},
		{"connection_timed_out", "connection timed out", CategoryNetwork},
		{"no_such_host", "no such host", CategoryNetwork},
		{"network_unreachable", "network unreachable", CategoryNetwork},
		{"timeout", "timeout occurred", CategoryNetwork},
		{"io_timeout", "i/o timeout", CategoryNetwork},
		{"econnrefused", "dial tcp: connection refused (econnrefused)", CategoryNetwork},
		{"econnreset", "read tcp: connection reset by peer (econnreset)", CategoryNetwork},
		{"etimedout", "dial tcp: connection timed out (etimedout)", CategoryNetwork},
		{"temporary_failure", "temporary failure in name resolution", CategoryNetwork},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			cat := CategorizeError(err)
			if cat != tt.expected {
				t.Errorf("CategorizeError(%q) = %v, want %v", tt.errMsg, cat, tt.expected)
			}
		})
	}
}

func TestCategorizeError_Auth(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected ErrorCategory
	}{
		{"unauthorized", "unauthorized: invalid credentials", CategoryAuth},
		{"forbidden", "forbidden: access denied", CategoryAuth},
		{"invalid_grant", "invalid_grant: token expired", CategoryAuth},
		{"access_denied", "access_denied", CategoryAuth},
		{"token_missing", "token is missing", CategoryAuth},
		{"credential_invalid", "invalid credential", CategoryAuth},
		{"authentication_failed", "authentication failed", CategoryAuth},
		{"oauth_error", "oauth2: cannot fetch token", CategoryAuth},
		{"token_revoked", "token has been revoked", CategoryAuth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			cat := CategorizeError(err)
			if cat != tt.expected {
				t.Errorf("CategorizeError(%q) = %v, want %v", tt.errMsg, cat, tt.expected)
			}
		})
	}
}

func TestCategorizeError_API(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected ErrorCategory
	}{
		{"status_400", "API error: status 400", CategoryAPI},
		// "HTTP 401" doesn't contain "unauthorized" or "status 401" so it falls to system
		{"status_401_with_unauthorized", "unauthorized: status 401", CategoryAuth},
		{"status_403", "drive: status 403", CategoryAPI},
		{"status_404", "file not found: status 404", CategoryAPI},
		{"status_429", "rate limit exceeded: status 429", CategoryAPI},
		{"status_500", "server error: status 500", CategoryAPI},
		{"status_502", "bad gateway: status 502", CategoryAPI},
		{"status_503", "service unavailable: status 503", CategoryAPI},
		// Note: "gateway timeout" matches network before API due to "timeout" pattern
		{"status_504_gateway_timeout", "gateway timeout: status 504", CategoryNetwork},
		// Pure API status code without trigger words
		{"status_504_standalone", "status 504", CategoryAPI},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			cat := CategorizeError(err)
			if cat != tt.expected {
				t.Errorf("CategorizeError(%q) = %v, want %v", tt.errMsg, cat, tt.expected)
			}
		})
	}
}

func TestCategorizeError_Storage(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected ErrorCategory
	}{
		{"file_open_error", "failed to open file", CategoryStorage},
		{"read_error", "read error: no such file", CategoryStorage},
		{"write_error", "write error: disk full", CategoryStorage},
		{"file_not_found", "file does not exist", CategoryStorage},
		{"directory_missing", "directory not found", CategoryStorage},
		{"no_space_left", "no space left on device", CategoryStorage},
		{"permission_denied", "permission denied reading file", CategoryStorage},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			cat := CategorizeError(err)
			if cat != tt.expected {
				t.Errorf("CategorizeError(%q) = %v, want %v", tt.errMsg, cat, tt.expected)
			}
		})
	}
}

func TestCategorizeError_Config(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected ErrorCategory
	}{
		{"yaml_parse_error", "failed to parse YAML", CategoryConfig},
		{"invalid_value", "invalid configuration value", CategoryConfig},
		{"missing_required", "required field is missing", CategoryConfig},
		{"config_not_found", "config not found", CategoryConfig},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			cat := CategorizeError(err)
			if cat != tt.expected {
				t.Errorf("CategorizeError(%q) = %v, want %v", tt.errMsg, cat, tt.expected)
			}
		})
	}
}

func TestCategorizeError_System(t *testing.T) {
	tests := []struct {
		name     string
		errMsg   string
		expected ErrorCategory
	}{
		{"generic_error", "something went wrong", CategorySystem},
		{"unknown_error", "internal error occurred", CategorySystem},
		{"nil_error", "", CategoryUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.errMsg != "" {
				err = errors.New(tt.errMsg)
			}
			cat := CategorizeError(err)
			if cat != tt.expected {
				t.Errorf("CategorizeError(%q) = %v, want %v", tt.errMsg, cat, tt.expected)
			}
		})
	}
}

func TestErrorCategory_IsRetryable(t *testing.T) {
	tests := []struct {
		category ErrorCategory
		expected bool
	}{
		{CategoryUnknown, false},
		{CategoryNetwork, true},
		{CategoryAuth, false},
		{CategoryAPI, true}, // 5xx and 429 are retryable
		{CategoryStorage, false},
		{CategoryConfig, false},
		{CategorySystem, false},
	}

	for _, tt := range tests {
		t.Run(tt.category.String(), func(t *testing.T) {
			retryable := tt.category.IsRetryable()
			if retryable != tt.expected {
				t.Errorf("Category(%v).IsRetryable() = %v, want %v", tt.category, retryable, tt.expected)
			}
		})
	}
}

func TestErrorCategory_String(t *testing.T) {
	tests := []struct {
		category ErrorCategory
		expected string
	}{
		{CategoryUnknown, "unknown"},
		{CategoryNetwork, "network"},
		{CategoryAuth, "auth"},
		{CategoryAPI, "api"},
		{CategoryStorage, "storage"},
		{CategoryConfig, "config"},
		{CategorySystem, "system"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			s := tt.category.String()
			if s != tt.expected {
				t.Errorf("Category(%v).String() = %q, want %q", tt.category, s, tt.expected)
			}
		})
	}
}

func TestIsRetryableHTTPStatus(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
		{504, true},
	}

	for _, tt := range tests {
		t.Run(httpStatusString(tt.code), func(t *testing.T) {
			result := isRetryableHTTPStatus(tt.code)
			if result != tt.expected {
				t.Errorf("isRetryableHTTPStatus(%d) = %v, want %v", tt.code, result, tt.expected)
			}
		})
	}
}

func httpStatusString(code int) string {
	switch code {
	case 200:
		return "200_OK"
	case 400:
		return "400_BadRequest"
	case 401:
		return "401_Unauthorized"
	case 403:
		return "403_Forbidden"
	case 404:
		return "404_NotFound"
	case 429:
		return "429_RateLimit"
	case 500:
		return "500_InternalServerError"
	case 502:
		return "502_BadGateway"
	case 503:
		return "503_ServiceUnavailable"
	case 504:
		return "504_GatewayTimeout"
	default:
		return "unknown"
	}
}

func TestNewUploadError(t *testing.T) {
	tests := []struct {
		name       string
		callID     string
		errMsg     string
		statusCode int
		wantRetry  bool
		wantCat    ErrorCategory
	}{
		{
			name:       "network_error_timeout",
			callID:     "call123",
			errMsg:     "connection timed out",
			statusCode: 504, // Gateway timeout is retryable
			wantRetry:  true,
			wantCat:    CategoryNetwork, // "timed out" matches network first
		},
		{
			name:       "api_error_500",
			callID:     "call456",
			errMsg:     "status 500", // explicit status code in error
			statusCode: 500,
			wantRetry:  true,
			wantCat:    CategoryAPI,
		},
		{
			name:       "api_error_401",
			callID:     "call789",
			errMsg:     "status 401",
			statusCode: 401,
			wantRetry:  false, // 401 is not in 5xx or 429 range
			wantCat:    CategoryAPI,
		},
		{
			name:       "auth_error",
			callID:     "callabc",
			errMsg:     "invalid_grant",
			statusCode: 401,
			wantRetry:  false,
			wantCat:    CategoryAuth,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			uploadErr := NewUploadError(tt.callID, err, tt.statusCode)

			if uploadErr.CallID != tt.callID {
				t.Errorf("CallID = %q, want %q", uploadErr.CallID, tt.callID)
			}
			if uploadErr.Retryable != tt.wantRetry {
				t.Errorf("Retryable = %v, want %v", uploadErr.Retryable, tt.wantRetry)
			}
			if uploadErr.Category != tt.wantCat {
				t.Errorf("Category = %v, want %v", uploadErr.Category, tt.wantCat)
			}
			if uploadErr.StatusCode != tt.statusCode {
				t.Errorf("StatusCode = %d, want %d", uploadErr.StatusCode, tt.statusCode)
			}
		})
	}
}

func TestUploadError_Error(t *testing.T) {
	innerErr := errors.New("inner error")
	uploadErr := &UploadError{
		CallID:    "call123",
		Err:       innerErr,
		Category:  CategoryNetwork,
		Retryable: true,
	}

	if uploadErr.Error() != innerErr.Error() {
		t.Errorf("Error() = %q, want %q", uploadErr.Error(), innerErr.Error())
	}
}

func TestUploadError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	uploadErr := &UploadError{
		CallID:    "call123",
		Err:       innerErr,
		Category:  CategoryNetwork,
		Retryable: true,
	}

	if uploadErr.Unwrap() != innerErr {
		t.Errorf("Unwrap() = %v, want %v", uploadErr.Unwrap(), innerErr)
	}
}
