// Package ports defines interfaces for external dependencies.
package ports

import (
	"image"
)

// CompareOptions configures image comparison behavior.
type CompareOptions struct {
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

	// LabelFontPath is the path to a TrueType font file for labels (optional)
	// If not specified, a basic built-in font will be used
	LabelFontPath string

	// LabelFontSize is the font size for labels in points (default: 14)
	LabelFontSize float64

	// BaselineLabel is the label text for the baseline panel (default: "baseline")
	BaselineLabel string

	// DiffLabel is the label text for the diff panel (default: "diff")
	DiffLabel string

	// CurrentLabel is the label text for the current panel (default: "current")
	CurrentLabel string
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
