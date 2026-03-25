package logger

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	field := String("key", "value")
	if field.Key != "key" {
		t.Errorf("String().Key = %q, want %q", field.Key, "key")
	}
	if field.Value != "value" {
		t.Errorf("String().Value = %v, want %v", field.Value, "value")
	}
}

func TestInt(t *testing.T) {
	field := Int("count", 42)
	if field.Key != "count" {
		t.Errorf("Int().Key = %q, want %q", field.Key, "count")
	}
	if field.Value != 42 {
		t.Errorf("Int().Value = %v, want %v", field.Value, 42)
	}
}

func TestInt64(t *testing.T) {
	field := Int64("big", 1<<33)
	if field.Key != "big" {
		t.Errorf("Int64().Key = %q, want %q", field.Key, "big")
	}
	if field.Value != int64(1<<33) {
		t.Errorf("Int64().Value = %v, want %v", field.Value, int64(1<<33))
	}
}

func TestBool(t *testing.T) {
	field := Bool("flag", true)
	if field.Key != "flag" {
		t.Errorf("Bool().Key = %q, want %q", field.Key, "flag")
	}
	if field.Value != true {
		t.Errorf("Bool().Value = %v, want %v", field.Value, true)
	}
}

func TestErr(t *testing.T) {
	testErr := errors.New("test error")
	field := Err(testErr)
	if field.Key != "error" {
		t.Errorf("Err().Key = %q, want %q", field.Key, "error")
	}
	if field.Value != "test error" {
		t.Errorf("Err().Value = %v, want %v", field.Value, "test error")
	}
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level    Level
		expected string
	}{
		{LevelDebug, "DEBUG"},
		{LevelInfo, "INFO"},
		{LevelWarn, "WARN"},
		{LevelError, "ERROR"},
		{Level(100), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("Level.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestSetLevel(t *testing.T) {
	// Save original level
	origLevel := defaultLogger.level

	// Test setting each level
	SetLevel(LevelDebug)
	if defaultLogger.level != LevelDebug {
		t.Errorf("SetLevel(LevelDebug) failed, got %v", defaultLogger.level)
	}

	SetLevel(LevelInfo)
	if defaultLogger.level != LevelInfo {
		t.Errorf("SetLevel(LevelInfo) failed, got %v", defaultLogger.level)
	}

	SetLevel(LevelWarn)
	if defaultLogger.level != LevelWarn {
		t.Errorf("SetLevel(LevelWarn) failed, got %v", defaultLogger.level)
	}

	SetLevel(LevelError)
	if defaultLogger.level != LevelError {
		t.Errorf("SetLevel(LevelError) failed, got %v", defaultLogger.level)
	}

	// Restore original level
	defaultLogger.level = origLevel
}

func TestSetOutput(t *testing.T) {
	// Save original output
	origOutput := defaultLogger.output

	// Create a buffer to test
	buf := &bytes.Buffer{}
	SetOutput(buf)

	if defaultLogger.output != buf {
		t.Errorf("SetOutput() did not set the output correctly")
	}

	// Restore original output
	defaultLogger.output = origOutput
}

func TestDebug(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelDebug)

	Debug("test message", String("key", "value"))

	output := buf.String()
	if !strings.Contains(output, "DEBUG") {
		t.Errorf("Debug output should contain 'DEBUG', got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Debug output should contain 'test message', got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Debug output should contain 'key=value', got: %s", output)
	}
}

func TestInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelInfo)

	Info("info message", Int("count", 42))

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("Info output should contain 'INFO', got: %s", output)
	}
	if !strings.Contains(output, "info message") {
		t.Errorf("Info output should contain 'info message', got: %s", output)
	}
	if !strings.Contains(output, "count=42") {
		t.Errorf("Info output should contain 'count=42', got: %s", output)
	}
}

func TestWarn(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelWarn)

	Warn("warning message", Bool("flag", true))

	output := buf.String()
	if !strings.Contains(output, "WARN") {
		t.Errorf("Warn output should contain 'WARN', got: %s", output)
	}
	if !strings.Contains(output, "warning message") {
		t.Errorf("Warn output should contain 'warning message', got: %s", output)
	}
}

func TestError(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelError)

	testErr := errors.New("test error")
	Error("error message", Err(testErr))

	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Errorf("Error output should contain 'ERROR', got: %s", output)
	}
	if !strings.Contains(output, "error message") {
		t.Errorf("Error output should contain 'error message', got: %s", output)
	}
	if !strings.Contains(output, "error=test error") {
		t.Errorf("Error output should contain 'error=test error', got: %s", output)
	}
}

func TestDebugf(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelDebug)

	Debugf("number: %d, string: %s", 42, "test")

	output := buf.String()
	if !strings.Contains(output, "number: 42, string: test") {
		t.Errorf("Debugf output mismatch, got: %s", output)
	}
}

func TestInfof(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelInfo)

	Infof("value: %v", float64(3.14))

	output := buf.String()
	if !strings.Contains(output, "value: 3.14") {
		t.Errorf("Infof output mismatch, got: %s", output)
	}
}

func TestLogLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelWarn) // Only warnings and above should be logged

	Info("should not appear")
	Debug("also should not appear")

	output := buf.String()
	if output != "" {
		t.Errorf("Expected empty output due to log level filtering, got: %s", output)
	}

	// Now set to Info and verify Info appears
	SetLevel(LevelInfo)
	buf.Reset()
	Info("should appear")
	output = buf.String()
	if !strings.Contains(output, "should appear") {
		t.Errorf("Expected 'should appear' in output, got: %s", output)
	}
}

func TestCloudLogger_OAuthStart(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelDebug)

	Cloud().OAuthStart("test@example.com")

	output := buf.String()
	if !strings.Contains(output, "OAuth flow started") {
		t.Errorf("Expected 'OAuth flow started' in output, got: %s", output)
	}
	if !strings.Contains(output, "provider=google") {
		t.Errorf("Expected 'provider=google' in output, got: %s", output)
	}
	if !strings.Contains(output, "email=test@example.com") {
		t.Errorf("Expected 'email=test@example.com' in output, got: %s", output)
	}
}

func TestCloudLogger_OAuthComplete_Success(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelDebug)

	Cloud().OAuthComplete("test@example.com", true)

	output := buf.String()
	if !strings.Contains(output, "OAuth flow completed") {
		t.Errorf("Expected 'OAuth flow completed' in output, got: %s", output)
	}
	if !strings.Contains(output, "success=true") {
		t.Errorf("Expected 'success=true' in output, got: %s", output)
	}
}

func TestCloudLogger_OAuthComplete_Failure(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelError)

	Cloud().OAuthComplete("test@example.com", false)

	output := buf.String()
	if !strings.Contains(output, "ERROR") {
		t.Errorf("Expected 'ERROR' level in output for failure, got: %s", output)
	}
}

func TestCloudLogger_OAuthError(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelError)

	err := errors.New("authentication failed")
	Cloud().OAuthError(err, "token_exchange")

	output := buf.String()
	if !strings.Contains(output, "OAuth error") {
		t.Errorf("Expected 'OAuth error' in output, got: %s", output)
	}
	if !strings.Contains(output, "stage=token_exchange") {
		t.Errorf("Expected 'stage=token_exchange' in output, got: %s", output)
	}
}

func TestCloudLogger_UploadStart(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelDebug)

	Cloud().UploadStart("call123", "/path/to/file.md")

	output := buf.String()
	if !strings.Contains(output, "Upload started") {
		t.Errorf("Expected 'Upload started' in output, got: %s", output)
	}
	if !strings.Contains(output, "call_id=call123") {
		t.Errorf("Expected 'call_id=call123' in output, got: %s", output)
	}
}

func TestCloudLogger_UploadComplete(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelDebug)

	Cloud().UploadComplete("call123", "fileid456", 1024)

	output := buf.String()
	if !strings.Contains(output, "Upload completed") {
		t.Errorf("Expected 'Upload completed' in output, got: %s", output)
	}
	if !strings.Contains(output, "remote_id=fileid456") {
		t.Errorf("Expected 'remote_id=fileid456' in output, got: %s", output)
	}
}

func TestCloudLogger_UploadError(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelError)

	err := errors.New("connection reset")
	Cloud().UploadError("call123", err, true)

	output := buf.String()
	if !strings.Contains(output, "Upload failed") {
		t.Errorf("Expected 'Upload failed' in output, got: %s", output)
	}
	if !strings.Contains(output, "retryable=true") {
		t.Errorf("Expected 'retryable=true' in output, got: %s", output)
	}
}

func TestCloudLogger_UploadRetry(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelWarn)

	Cloud().UploadRetry("call123", 2, 5, 60)

	output := buf.String()
	if !strings.Contains(output, "Upload retry scheduled") {
		t.Errorf("Expected 'Upload retry scheduled' in output, got: %s", output)
	}
	if !strings.Contains(output, "WARN") {
		t.Errorf("Expected 'WARN' level in output, got: %s", output)
	}
	if !strings.Contains(output, "attempt=2") {
		t.Errorf("Expected 'attempt=2' in output, got: %s", output)
	}
}

func TestCloudLogger_SyncEnqueue(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelDebug)

	Cloud().SyncEnqueue("call123", "folder456")

	output := buf.String()
	if !strings.Contains(output, "Sync job enqueued") {
		t.Errorf("Expected 'Sync job enqueued' in output, got: %s", output)
	}
	if !strings.Contains(output, "folder_id=folder456") {
		t.Errorf("Expected 'folder_id=folder456' in output, got: %s", output)
	}
}

func TestCloudLogger_SyncDequeue(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelDebug)

	Cloud().SyncDequeue("call123")

	output := buf.String()
	if !strings.Contains(output, "Sync job dequeued") {
		t.Errorf("Expected 'Sync job dequeued' in output, got: %s", output)
	}
}

func TestCloudLogger_SyncComplete(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelDebug)

	Cloud().SyncComplete("call123")

	output := buf.String()
	if !strings.Contains(output, "Sync completed") {
		t.Errorf("Expected 'Sync completed' in output, got: %s", output)
	}
}

func TestCloudLogger_SyncPermanentFail(t *testing.T) {
	buf := &bytes.Buffer{}
	SetOutput(buf)
	SetLevel(LevelError)

	err := errors.New("folder not found")
	Cloud().SyncPermanentFail("call123", err)

	output := buf.String()
	if !strings.Contains(output, "Sync permanently failed") {
		t.Errorf("Expected 'Sync permanently failed' in output, got: %s", output)
	}
	if !strings.Contains(output, "ERROR") {
		t.Errorf("Expected 'ERROR' level in output, got: %s", output)
	}
}
