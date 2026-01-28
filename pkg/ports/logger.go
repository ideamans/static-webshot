// Package ports defines interfaces for external dependencies.
package ports

// LogLevel represents the severity of a log message.
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
)

// Logger abstracts logging operations.
type Logger interface {
	// Debug logs a debug-level message.
	Debug(msg string, args ...interface{})

	// Info logs an info-level message.
	Info(msg string, args ...interface{})

	// Warn logs a warning-level message.
	Warn(msg string, args ...interface{})

	// Error logs an error-level message.
	Error(msg string, args ...interface{})

	// SetLevel sets the minimum log level to output.
	SetLevel(level LogLevel)
}
