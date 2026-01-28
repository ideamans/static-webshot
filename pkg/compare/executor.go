// Package compare provides the compare command execution logic.
package compare

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ideamans/go-page-visual-regression-tester/pkg/ports"
)

// Executor executes the compare command.
type Executor struct {
	processor  ports.ImageProcessor
	filesystem ports.FileSystem
	logger     ports.Logger
}

// NewExecutor creates a new Executor with the given dependencies.
func NewExecutor(processor ports.ImageProcessor, filesystem ports.FileSystem, logger ports.Logger) *Executor {
	return &Executor{
		processor:  processor,
		filesystem: filesystem,
		logger:     logger,
	}
}

// Execute runs the compare command with the given configuration.
func (e *Executor) Execute(ctx context.Context, cfg Config) (*Result, error) {
	baseline, err := e.processor.LoadImage(cfg.BaselinePath)
	if err != nil {
		return nil, fmt.Errorf("load baseline: %w", err)
	}

	current, err := e.processor.LoadImage(cfg.CurrentPath)
	if err != nil {
		return nil, fmt.Errorf("load current: %w", err)
	}

	e.logger.Debug("Comparing %s vs %s", cfg.BaselinePath, cfg.CurrentPath)
	compareOpts := ports.CompareOptions{
		Threshold:          cfg.Threshold,
		ColorThreshold:     cfg.ColorThreshold,
		IgnoreAntialiasing: cfg.IgnoreAntialiasing,
		MaxHeight:          cfg.MaxHeight,
		DiffOverlay:        cfg.DiffOverlay,
	}

	compareResult, err := e.processor.Compare(baseline, current, compareOpts)
	if err != nil {
		return nil, fmt.Errorf("compare images: %w", err)
	}

	// Save diff image
	if err := e.processor.SaveImage(cfg.OutputPath, compareResult.DiffImage); err != nil {
		return nil, fmt.Errorf("save diff image: %w", err)
	}

	result := &Result{
		Pass:           compareResult.Pass,
		PixelDiffCount: compareResult.PixelDiffCount,
		PixelDiffRatio: compareResult.PixelDiffRatio,
		TotalPixels:    compareResult.TotalPixels,
		Threshold:      cfg.Threshold,
		BaselinePath:   cfg.BaselinePath,
		CurrentPath:    cfg.CurrentPath,
		DiffPath:       cfg.OutputPath,
	}

	// Generate digest text
	digest := e.generateDigest(result)

	// Output digest to stdout
	fmt.Println(digest)

	// Save digest to file if path is specified
	if cfg.DigestPath != "" {
		if err := e.saveDigest(cfg.DigestPath, digest); err != nil {
			return nil, fmt.Errorf("save digest: %w", err)
		}
	}

	return result, nil
}

// generateDigest creates a digest text summary of the comparison result.
func (e *Executor) generateDigest(result *Result) string {
	status := "PASS"
	if !result.Pass {
		status = "FAIL"
	}

	return fmt.Sprintf(`[Compare Result]
Status: %s
Baseline: %s
Current: %s
Output: %s
Diff Pixels: %d / %d (%.2f%%)
Threshold: %.2f%%`,
		status,
		result.BaselinePath,
		result.CurrentPath,
		result.DiffPath,
		result.PixelDiffCount,
		result.TotalPixels,
		result.PixelDiffRatio*100,
		result.Threshold*100,
	)
}

// saveDigest saves the digest text to a file.
func (e *Executor) saveDigest(path, digest string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(digest+"\n"), 0644); err != nil {
		return fmt.Errorf("write digest file %s: %w", path, err)
	}

	return nil
}
