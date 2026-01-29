package pixelmatch

import (
	"image"
	"image/color"
	"testing"

	"github.com/ideamans/go-page-visual-regression-tester/pkg/ports"
)

func createTestImage(width, height int, c color.Color) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
	return img
}

func TestProcessor_Compare_IdenticalImages(t *testing.T) {
	processor := New()

	baseline := createTestImage(100, 100, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	current := createTestImage(100, 100, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	result, err := processor.Compare(baseline, current, ports.CompareOptions{})
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}

	if result.PixelDiffCount != 0 {
		t.Errorf("Compare() PixelDiffCount = %d, want 0", result.PixelDiffCount)
	}
	if result.TotalPixels != 10000 {
		t.Errorf("Compare() TotalPixels = %d, want 10000", result.TotalPixels)
	}
	if result.PixelDiffRatio != 0 {
		t.Errorf("Compare() PixelDiffRatio = %f, want 0", result.PixelDiffRatio)
	}
}

func TestProcessor_Compare_DifferentImages(t *testing.T) {
	processor := New()

	baseline := createTestImage(100, 100, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	current := createTestImage(100, 100, color.RGBA{R: 0, G: 255, B: 0, A: 255})

	result, err := processor.Compare(baseline, current, ports.CompareOptions{})
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}

	if result.PixelDiffCount == 0 {
		t.Error("Compare() expected PixelDiffCount > 0 for different images")
	}
	if result.PixelDiffRatio == 0 {
		t.Error("Compare() expected PixelDiffRatio > 0 for different images")
	}
}

func TestProcessor_Compare_DimensionMismatch(t *testing.T) {
	processor := New()

	baseline := createTestImage(100, 100, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	current := createTestImage(200, 100, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// Dimension mismatch is handled by normalizing images to the same size
	result, err := processor.Compare(baseline, current, ports.CompareOptions{})
	if err != nil {
		t.Fatalf("Compare() error = %v, want no error (dimensions normalized)", err)
	}

	// The normalized comparison should use the max dimensions (200x100)
	if result.TotalPixels != 200*100 {
		t.Errorf("Compare() TotalPixels = %d, want %d", result.TotalPixels, 200*100)
	}
}

func TestProcessor_Compare_WithIgnoreRegions(t *testing.T) {
	processor := New()

	// Create images that differ only in the center
	baseline := createTestImage(100, 100, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	current := createTestImage(100, 100, color.RGBA{R: 255, G: 0, B: 0, A: 255})

	// Make the current image different in the center
	currentRGBA := current.(*image.RGBA)
	for y := 40; y < 60; y++ {
		for x := 40; x < 60; x++ {
			currentRGBA.Set(x, y, color.RGBA{R: 0, G: 255, B: 0, A: 255})
		}
	}

	// Compare without ignore regions - should have diffs
	result1, err := processor.Compare(baseline, current, ports.CompareOptions{})
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
	diffWithoutIgnore := result1.PixelDiffCount

	// Compare with ignore region covering the difference - should have fewer diffs
	result2, err := processor.Compare(baseline, current, ports.CompareOptions{
		IgnoreRegions: []ports.IgnoreRegion{
			{X: 35, Y: 35, Width: 30, Height: 30},
		},
	})
	if err != nil {
		t.Fatalf("Compare() error = %v", err)
	}

	// The ignore region should reduce the diff count
	if result2.PixelDiffCount > diffWithoutIgnore {
		t.Errorf("Compare() with ignore region should have fewer diffs: without=%d, with=%d",
			diffWithoutIgnore, result2.PixelDiffCount)
	}
}
