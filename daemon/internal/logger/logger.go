// Package logger provides structured logging for the Verbalizer daemon.
package logger

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

// Level represents the log level.
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Logger provides structured logging.
type Logger struct {
	mu     sync.Mutex
	output io.Writer
	level  Level
}

// Global default logger
var defaultLogger = &Logger{
	output: os.Stdout,
	level:  LevelInfo,
}

// Field represents a log field.
type Field struct {
	Key   string
	Value interface{}
}

// String creates a string field.
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int creates an integer field.
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 creates an int64 field.
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Bool creates a boolean field.
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Err creates an error field.
func Err(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// WithFields creates a log entry with fields.
func WithFields(fields ...Field) *Entry {
	return &Entry{logger: defaultLogger, fields: fields}
}

// WithField creates a log entry with a single field.
func WithField(key string, value interface{}) *Entry {
	return WithFields(Field{Key: key, Value: value})
}

// Debug logs a debug message.
func Debug(msg string, fields ...Field) {
	WithFields(fields...).debug(msg)
}

// Info logs an info message.
func Info(msg string, fields ...Field) {
	WithFields(fields...).info(msg)
}

// Warn logs a warning message.
func Warn(msg string, fields ...Field) {
	WithFields(fields...).warn(msg)
}

// Error logs an error message.
func Error(msg string, fields ...Field) {
	WithFields(fields...).error(msg)
}

// Debugf logs a formatted debug message.
func Debugf(format string, args ...interface{}) {
	Debug(fmt.Sprintf(format, args...))
}

// Infof logs a formatted info message.
func Infof(format string, args ...interface{}) {
	Info(fmt.Sprintf(format, args...))
}

// Warnf logs a formatted warning message.
func Warnf(format string, args ...interface{}) {
	Warn(fmt.Sprintf(format, args...))
}

// Errorf logs a formatted error message with format info.
func Errorf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	Error(msg, String("format", format))
}

// SetLevel sets the global log level.
func SetLevel(level Level) {
	defaultLogger.setLevel(level)
}

// SetOutput sets the global log output.
func SetOutput(w io.Writer) {
	defaultLogger.setOutput(w)
}

func (l *Logger) setLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) setOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// Entry represents a log entry being built.
type Entry struct {
	logger *Logger
	fields []Field
}

func (e *Entry) debug(msg string) {
	e.log(LevelDebug, msg)
}

func (e *Entry) info(msg string) {
	e.log(LevelInfo, msg)
}

func (e *Entry) warn(msg string) {
	e.log(LevelWarn, msg)
}

func (e *Entry) error(msg string) {
	e.log(LevelError, msg)
}

func (e *Entry) log(level Level, msg string) {
	e.logger.mu.Lock()
	defer e.logger.mu.Unlock()

	if level < e.logger.level {
		return
	}

	// Format: LEVEL RFC3339 message field1=value1 field2=value2
	timestamp := time.Now().Format(time.RFC3339)
	fmt.Fprintf(e.logger.output, "%s %s %s", level.String(), timestamp, msg)
	for _, f := range e.fields {
		fmt.Fprintf(e.logger.output, " %s=%v", f.Key, f.Value)
	}
	fmt.Fprintln(e.logger.output)
}
