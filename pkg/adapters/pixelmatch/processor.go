// Package pixelmatch provides an image processor using pixel comparison.
package pixelmatch

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/orisano/pixelmatch"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"github.com/ideamans/go-page-visual-regression-tester/pkg/ports"
)

// Processor implements ports.ImageProcessor using pixel comparison.
type Processor struct{}

// New creates a new Processor.
func New() *Processor {
	return &Processor{}
}

// LoadImage loads an image from the given file path.
func (p *Processor) LoadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open image %s: %w", path, err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("decode image %s: %w", path, err)
	}

	return img, nil
}

// SaveImage saves an image to the given file path.
func (p *Processor) SaveImage(path string, img image.Image) error {
	// Ensure the parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create directory %s: %w", dir, err)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create image %s: %w", path, err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encode image %s: %w", path, err)
	}

	return nil
}

// Compare compares two images and returns the comparison result.
func (p *Processor) Compare(baseline, current image.Image, opts ports.CompareOptions) (*ports.CompareResult, error) {
	baselineBounds := baseline.Bounds()
	currentBounds := current.Bounds()

	// Calculate comparison dimensions (use max of both)
	width := max(baselineBounds.Dx(), currentBounds.Dx())
	height := max(baselineBounds.Dy(), currentBounds.Dy())

	// Apply max height limit if specified
	if opts.MaxHeight > 0 && height > opts.MaxHeight {
		height = opts.MaxHeight
	}

	totalPixels := width * height

	// Crop/normalize both images to the same size
	baseline = p.normalizeImage(baseline, width, height)
	current = p.normalizeImage(current, width, height)

	// Apply ignore regions by masking them in both images
	maskedBaseline := baseline
	maskedCurrent := current

	if len(opts.IgnoreRegions) > 0 {
		maskedBaseline = p.applyMask(baseline, opts.IgnoreRegions)
		maskedCurrent = p.applyMask(current, opts.IgnoreRegions)
	}

	// Color threshold (normalized to 0-1 range for pixelmatch)
	colorThreshold := float64(opts.ColorThreshold)
	if colorThreshold == 0 {
		colorThreshold = 10
	}
	colorThreshold = colorThreshold / 255.0

	// Build comparison options with diff image output
	var diffImgPtr image.Image
	matchOpts := []pixelmatch.MatchOption{
		pixelmatch.Threshold(colorThreshold),
		pixelmatch.Alpha(0.1),
		pixelmatch.DiffColor(color.RGBA{R: 255, G: 0, B: 0, A: 255}),
		pixelmatch.WriteTo(&diffImgPtr),
		pixelmatch.EnableDiffMask,
	}

	// Include antialiasing detection unless explicitly disabled
	if !opts.IgnoreAntialiasing {
		matchOpts = append(matchOpts, pixelmatch.IncludeAntiAlias)
	}

	// Perform comparison
	diffCount, err := pixelmatch.MatchPixel(maskedBaseline, maskedCurrent, matchOpts...)
	if err != nil {
		return nil, fmt.Errorf("pixel comparison: %w", err)
	}

	// Generate diff image based on mode
	var diffImg image.Image
	if opts.DiffOverlay {
		// Create side-by-side composite: before | diff | after
		diffPanel := p.createOverlayDiffImage(maskedBaseline, maskedCurrent, colorThreshold)
		labels := []string{opts.BaselineLabel, opts.DiffLabel, opts.CurrentLabel}
		// Apply defaults if empty
		if labels[0] == "" {
			labels[0] = "baseline"
		}
		if labels[1] == "" {
			labels[1] = "diff"
		}
		if labels[2] == "" {
			labels[2] = "current"
		}
		diffImg = p.createCompositeImage(maskedBaseline, maskedCurrent, diffPanel, opts.LabelFontPath, opts.LabelFontSize, labels)
	} else if diffImgPtr != nil {
		diffImg = diffImgPtr
	} else {
		// Fallback: create standard diff image
		diffImg = p.createDiffImage(maskedBaseline, maskedCurrent, colorThreshold)
	}

	diffRatio := float64(diffCount) / float64(totalPixels)

	return &ports.CompareResult{
		PixelDiffCount: diffCount,
		PixelDiffRatio: diffRatio,
		TotalPixels:    totalPixels,
		DiffImage:      diffImg,
	}, nil
}

// createOverlayDiffImage creates the center diff panel.
// Shows before image faded (c' = c * 0.5 + 0.5) with red overlay on difference pixels.
func (p *Processor) createOverlayDiffImage(baseline, current image.Image, threshold float64) image.Image {
	bounds := baseline.Bounds()
	overlayImg := image.NewRGBA(bounds)

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			br, bg, bb, _ := baseline.At(x, y).RGBA()
			cr, cg, cb, _ := current.At(x, y).RGBA()

			// Calculate color difference (normalized to 0-1)
			dr := float64(int32(br)-int32(cr)) / 65535.0
			dg := float64(int32(bg)-int32(cg)) / 65535.0
			db := float64(int32(bb)-int32(cb)) / 65535.0

			diff := (abs(dr) + abs(dg) + abs(db)) / 3.0

			// Base: before image faded (c' = c * 0.5 + 0.5)
			// In 8-bit: c' = c / 2 + 128
			baseR := uint8(br>>9) + 128
			baseG := uint8(bg>>9) + 128
			baseB := uint8(bb>>9) + 128

			if diff > threshold {
				// Overlay red on difference pixels
				overlayImg.Set(x, y, color.RGBA{
					R: 255,
					G: baseG / 2,
					B: baseB / 2,
					A: 255,
				})
			} else {
				// Faded before image
				overlayImg.Set(x, y, color.RGBA{
					R: baseR,
					G: baseG,
					B: baseB,
					A: 255,
				})
			}
		}
	}

	return overlayImg
}

// createCompositeImage creates a side-by-side image: before | diff | after
func (p *Processor) createCompositeImage(baseline, current, diffPanel image.Image, fontPath string, fontSize float64, labels []string) image.Image {
	bounds := baseline.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Label bar height
	labelHeight := 24
	if fontSize > 0 && fontSize > 14 {
		labelHeight = int(fontSize) + 10
	}

	// Create composite image with 3 panels side by side + label bar
	composite := image.NewRGBA(image.Rect(0, 0, width*3, height+labelHeight))

	// Fill label bar with light gray
	labelBg := color.RGBA{R: 240, G: 240, B: 240, A: 255}
	for y := 0; y < labelHeight; y++ {
		for x := 0; x < width*3; x++ {
			composite.Set(x, y, labelBg)
		}
	}

	// Draw labels
	face := p.loadFont(fontPath, fontSize)
	for i, label := range labels {
		p.drawCenteredText(composite, label, i*width, 0, width, labelHeight, face)
	}

	// Draw 1px border line at bottom of label area
	borderColor := color.RGBA{R: 200, G: 200, B: 200, A: 255}
	for x := 0; x < width*3; x++ {
		composite.Set(x, labelHeight-1, borderColor)
	}

	// Panel 1: Before (left)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			composite.Set(x, y+labelHeight, baseline.At(x+bounds.Min.X, y+bounds.Min.Y))
		}
	}

	// Panel 2: Diff (center)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			composite.Set(x+width, y+labelHeight, diffPanel.At(x+bounds.Min.X, y+bounds.Min.Y))
		}
	}

	// Panel 3: After (right)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			composite.Set(x+width*2, y+labelHeight, current.At(x+bounds.Min.X, y+bounds.Min.Y))
		}
	}

	return composite
}

// loadFont loads a TrueType font from the given path, or returns a basic font if path is empty.
func (p *Processor) loadFont(fontPath string, fontSize float64) font.Face {
	if fontSize <= 0 {
		fontSize = 14
	}

	if fontPath != "" {
		data, err := os.ReadFile(fontPath)
		if err == nil {
			// Try parsing as single font first
			f, err := opentype.Parse(data)
			if err == nil {
				face, err := opentype.NewFace(f, &opentype.FaceOptions{
					Size:    fontSize,
					DPI:     72,
					Hinting: font.HintingFull,
				})
				if err == nil {
					return face
				}
			}

			// Try parsing as TrueType Collection (.ttc)
			collection, err := opentype.ParseCollection(data)
			if err == nil && collection.NumFonts() > 0 {
				f, err := collection.Font(0) // Use first font in collection
				if err == nil {
					face, err := opentype.NewFace(f, &opentype.FaceOptions{
						Size:    fontSize,
						DPI:     72,
						Hinting: font.HintingFull,
					})
					if err == nil {
						return face
					}
				}
			}
		}
	}

	// Fallback to basic font
	return basicfont.Face7x13
}

// drawCenteredText draws text centered within the given rectangle.
func (p *Processor) drawCenteredText(img *image.RGBA, text string, x, y, width, height int, face font.Face) {
	// Measure text width
	textWidth := font.MeasureString(face, text).Ceil()
	metrics := face.Metrics()
	textHeight := (metrics.Ascent + metrics.Descent).Ceil()

	// Calculate centered position
	posX := x + (width-textWidth)/2
	posY := y + (height+textHeight)/2 - metrics.Descent.Ceil()

	// Draw text
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(color.Black),
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.I(posX), Y: fixed.I(posY)},
	}
	d.DrawString(text)
}

// createDiffImage creates a diff image by comparing two images pixel by pixel.
func (p *Processor) createDiffImage(baseline, current image.Image, threshold float64) image.Image {
	bounds := baseline.Bounds()
	diffImg := image.NewRGBA(bounds)

	// Copy baseline as gray and mark differences in red
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			br, bg, bb, ba := baseline.At(x, y).RGBA()
			cr, cg, cb, ca := current.At(x, y).RGBA()

			// Calculate color difference (normalized to 0-1)
			dr := float64(int32(br)-int32(cr)) / 65535.0
			dg := float64(int32(bg)-int32(cg)) / 65535.0
			db := float64(int32(bb)-int32(cb)) / 65535.0
			da := float64(int32(ba)-int32(ca)) / 65535.0

			diff := (abs(dr) + abs(dg) + abs(db) + abs(da)) / 4.0

			if diff > threshold {
				// Mark as red for differences
				diffImg.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
			} else {
				// Gray for unchanged pixels
				gray := uint8((br + bg + bb) / 3 / 257)
				diffImg.Set(x, y, color.RGBA{R: gray, G: gray, B: gray, A: 255})
			}
		}
	}

	return diffImg
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// normalizeImage creates a new image of the specified size, copying the original
// and filling any extra space with a distinct color (magenta) to highlight size differences.
func (p *Processor) normalizeImage(img image.Image, width, height int) image.Image {
	bounds := img.Bounds()
	normalized := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill background with magenta (to highlight areas that don't exist in one image)
	fillColor := color.RGBA{R: 255, G: 0, B: 255, A: 255}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			normalized.Set(x, y, fillColor)
		}
	}

	// Copy original image content
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			normalized.Set(x-bounds.Min.X, y-bounds.Min.Y, img.At(x, y))
		}
	}

	return normalized
}

// applyMask creates a copy of the image with ignored regions filled with black.
func (p *Processor) applyMask(img image.Image, regions []ports.IgnoreRegion) image.Image {
	bounds := img.Bounds()
	masked := image.NewRGBA(bounds)

	// Copy original image
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			masked.Set(x, y, img.At(x, y))
		}
	}

	// Fill ignored regions with black
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	for _, region := range regions {
		for y := region.Y; y < region.Y+region.Height && y < bounds.Max.Y; y++ {
			for x := region.X; x < region.X+region.Width && x < bounds.Max.X; x++ {
				if x >= bounds.Min.X && y >= bounds.Min.Y {
					masked.Set(x, y, black)
				}
			}
		}
	}

	return masked
}

// Ensure Processor implements ports.ImageProcessor
var _ ports.ImageProcessor = (*Processor)(nil)
