// Package osfilesystem provides a standard os-based filesystem implementation.
package osfilesystem

import (
	"os"
	"path/filepath"

	"github.com/ideamans/go-page-visual-regression-tester/pkg/ports"
)

// OSFileSystem implements ports.FileSystem using the standard os package.
type OSFileSystem struct{}

// New creates a new OSFileSystem.
func New() *OSFileSystem {
	return &OSFileSystem{}
}

// ReadFile reads the entire file at the given path.
func (fs *OSFileSystem) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

// WriteFile writes data to the file at the given path.
func (fs *OSFileSystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	// Ensure the parent directory exists
	dir := filepath.Dir(path)
	if err := fs.MkdirAll(dir, 0755); err != nil {
		return err
	}
	return os.WriteFile(path, data, perm)
}

// Exists checks if a file or directory exists at the given path.
func (fs *OSFileSystem) Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// MkdirAll creates a directory along with any necessary parents.
func (fs *OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Ensure OSFileSystem implements ports.FileSystem
var _ ports.FileSystem = (*OSFileSystem)(nil)
