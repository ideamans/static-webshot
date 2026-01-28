// Package ports defines interfaces for external dependencies.
package ports

import "os"

// FileSystem abstracts file system operations.
type FileSystem interface {
	// ReadFile reads the entire file at the given path.
	ReadFile(path string) ([]byte, error)

	// WriteFile writes data to the file at the given path.
	WriteFile(path string, data []byte, perm os.FileMode) error

	// Exists checks if a file or directory exists at the given path.
	Exists(path string) bool

	// MkdirAll creates a directory along with any necessary parents.
	MkdirAll(path string, perm os.FileMode) error
}
