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
		ColorThreshold:     cfg.ColorThreshold,
		IgnoreAntialiasing: cfg.IgnoreAntialiasing,
		MaxHeight:          cfg.MaxHeight,
		DiffOverlay:        cfg.DiffOverlay,
		LabelFontPath:      cfg.LabelFontPath,
		LabelFontSize:      cfg.LabelFontSize,
		BaselineLabel:      cfg.BaselineLabel,
		DiffLabel:          cfg.DiffLabel,
		CurrentLabel:       cfg.CurrentLabel,
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
		PixelDiffCount: compareResult.PixelDiffCount,
		PixelDiffRatio: compareResult.PixelDiffRatio,
		TotalPixels:    compareResult.TotalPixels,
		BaselinePath:   cfg.BaselinePath,
		CurrentPath:    cfg.CurrentPath,
		DiffPath:       cfg.OutputPath,
	}

	// Generate digest text
	digest := e.generateDigest(result)

	// Output digest to stdout
	fmt.Println(digest)

	// Save text digest to file if path is specified
	if cfg.DigestTxtPath != "" {
		if err := e.saveFile(cfg.DigestTxtPath, digest+"\n"); err != nil {
			return nil, fmt.Errorf("save text digest: %w", err)
		}
	}

	// Save JSON digest to file if path is specified
	if cfg.DigestJSONPath != "" {
		jsonStr, err := result.ToJSON()
		if err != nil {
			return nil, fmt.Errorf("marshal JSON digest: %w", err)
		}
		if err := e.saveFile(cfg.DigestJSONPath, jsonStr+"\n"); err != nil {
			return nil, fmt.Errorf("save JSON digest: %w", err)
		}
	}

	return result, nil
}

// generateDigest creates a digest text summary of the comparison result.
func (e *Executor) generateDigest(result *Result) string {
	return fmt.Sprintf(`[Compare Result]
Baseline: %s
Current: %s
Output: %s
Diff Pixels: %d / %d
Diff Percent: %.4f%%`,
		result.BaselinePath,
		result.CurrentPath,
		result.DiffPath,
		result.PixelDiffCount,
		result.TotalPixels,
		result.PixelDiffRatio*100,
	)
}

// saveFile saves content to a file, creating directories as needed.
func (e *Executor) saveFile(path, content string) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}

	return nil
}
