package compare

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestResult_ToJSON(t *testing.T) {
	result := &Result{
		PixelDiffCount: 100,
		PixelDiffRatio: 0.001,
		TotalPixels:    100000,
		BaselinePath:   "/path/to/baseline.png",
		CurrentPath:    "/path/to/current.png",
		DiffPath:       "/path/to/diff.png",
	}

	jsonStr, err := result.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON() error = %v", err)
	}

	// Verify it's valid JSON
	var parsed Result
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("Invalid JSON: %v", err)
	}

	// Verify values
	if parsed.PixelDiffCount != result.PixelDiffCount {
		t.Errorf("PixelDiffCount = %d, want %d", parsed.PixelDiffCount, result.PixelDiffCount)
	}
	if parsed.PixelDiffRatio != result.PixelDiffRatio {
		t.Errorf("PixelDiffRatio = %f, want %f", parsed.PixelDiffRatio, result.PixelDiffRatio)
	}
	if parsed.DiffPercent != result.PixelDiffRatio*100 {
		t.Errorf("DiffPercent = %f, want %f", parsed.DiffPercent, result.PixelDiffRatio*100)
	}
}

func TestResult_ToText(t *testing.T) {
	result := &Result{
		PixelDiffCount: 100,
		PixelDiffRatio: 0.001,
		TotalPixels:    100000,
		BaselinePath:   "/path/to/baseline.png",
		CurrentPath:    "/path/to/current.png",
		DiffPath:       "/path/to/diff.png",
	}

	text := result.ToText()

	wantContains := []string{"Pixel Diff", "Diff Percent", "baseline.png", "current.png"}
	for _, want := range wantContains {
		if !strings.Contains(text, want) {
			t.Errorf("ToText() does not contain %q", want)
		}
	}
}
