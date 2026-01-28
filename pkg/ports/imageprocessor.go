// Package ports defines interfaces for external dependencies.
package ports

import (
	"image"
)

// CompareOptions configures image comparison behavior.
type CompareOptions struct {
	// Threshold is the acceptable pixel difference ratio (0.0 - 1.0)
	// Default: 0.15 (15% of pixels can differ)
	Threshold float64

	// ColorThreshold is the per-pixel color difference threshold (0-255)
	// Default: 10
	ColorThreshold int

	// IgnoreAntialiasing ignores antialiased pixels when comparing
	IgnoreAntialiasing bool

	// IgnoreRegions specifies rectangular areas to exclude from comparison
	IgnoreRegions []IgnoreRegion

	// MaxHeight limits comparison to the top N pixels (0 = no limit)
	MaxHeight int

	// DiffOverlay overlays diff markers on the current image instead of creating a separate diff image
	DiffOverlay bool
}

// IgnoreRegion defines a rectangular area to exclude from comparison.
type IgnoreRegion struct {
	X      int
	Y      int
	Width  int
	Height int
}

// CompareResult contains the results of image comparison.
type CompareResult struct {
	// Pass indicates whether the comparison passed (diff ratio <= threshold)
	Pass bool

	// PixelDiffCount is the number of differing pixels
	PixelDiffCount int

	// PixelDiffRatio is the ratio of differing pixels to total pixels
	PixelDiffRatio float64

	// TotalPixels is the total number of pixels compared
	TotalPixels int

	// DiffImage is the generated difference visualization image
	DiffImage image.Image
}

// ImageProcessor handles image loading, comparison, and diff generation.
type ImageProcessor interface {
	// LoadImage loads an image from the given file path.
	LoadImage(path string) (image.Image, error)

	// SaveImage saves an image to the given file path.
	SaveImage(path string, img image.Image) error

	// Compare compares two images and returns the comparison result.
	Compare(baseline, current image.Image, opts CompareOptions) (*CompareResult, error)
}
