// Package logger provides a console-based logger implementation.
package logger

import (
	"fmt"
	"os"
	"time"

	"github.com/ideamans/go-page-visual-regression-tester/pkg/ports"
)

// ConsoleLogger is a simple console-based logger.
type ConsoleLogger struct {
	level ports.LogLevel
}

// New creates a new ConsoleLogger with the default info level.
func New() *ConsoleLogger {
	return &ConsoleLogger{
		level: ports.LogLevelInfo,
	}
}

// Debug logs a debug-level message.
func (l *ConsoleLogger) Debug(msg string, args ...interface{}) {
	if l.level <= ports.LogLevelDebug {
		l.log("DEBUG", msg, args...)
	}
}

// Info logs an info-level message.
func (l *ConsoleLogger) Info(msg string, args ...interface{}) {
	if l.level <= ports.LogLevelInfo {
		l.log("INFO", msg, args...)
	}
}

// Warn logs a warning-level message.
func (l *ConsoleLogger) Warn(msg string, args ...interface{}) {
	if l.level <= ports.LogLevelWarn {
		l.log("WARN", msg, args...)
	}
}

// Error logs an error-level message.
func (l *ConsoleLogger) Error(msg string, args ...interface{}) {
	if l.level <= ports.LogLevelError {
		l.log("ERROR", msg, args...)
	}
}

// SetLevel sets the minimum log level to output.
func (l *ConsoleLogger) SetLevel(level ports.LogLevel) {
	l.level = level
}

func (l *ConsoleLogger) log(level, msg string, args ...interface{}) {
	timestamp := time.Now().Format("15:04:05")
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	fmt.Fprintf(os.Stderr, "[%s] %s: %s\n", timestamp, level, msg)
}

// Ensure ConsoleLogger implements ports.Logger
var _ ports.Logger = (*ConsoleLogger)(nil)
