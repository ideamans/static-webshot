package compare

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestResult_ToJSON(t *testing.T) {
	result := &Result{
		Pass:           true,
		PixelDiffCount: 100,
		PixelDiffRatio: 0.001,
		TotalPixels:    100000,
		Threshold:      0.15,
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
	if parsed.Pass != result.Pass {
		t.Errorf("Pass = %v, want %v", parsed.Pass, result.Pass)
	}
	if parsed.PixelDiffCount != result.PixelDiffCount {
		t.Errorf("PixelDiffCount = %d, want %d", parsed.PixelDiffCount, result.PixelDiffCount)
	}
	if parsed.PixelDiffRatio != result.PixelDiffRatio {
		t.Errorf("PixelDiffRatio = %f, want %f", parsed.PixelDiffRatio, result.PixelDiffRatio)
	}
}

func TestResult_ToText(t *testing.T) {
	tests := []struct {
		name           string
		pass           bool
		wantContains   []string
	}{
		{
			name:         "pass result",
			pass:         true,
			wantContains: []string{"PASS", "Pixel Diff", "Threshold"},
		},
		{
			name:         "fail result",
			pass:         false,
			wantContains: []string{"FAIL", "Pixel Diff", "Threshold"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := &Result{
				Pass:           tt.pass,
				PixelDiffCount: 100,
				PixelDiffRatio: 0.001,
				TotalPixels:    100000,
				Threshold:      0.15,
				BaselinePath:   "/path/to/baseline.png",
				CurrentPath:    "/path/to/current.png",
				DiffPath:       "/path/to/diff.png",
			}

			text := result.ToText()

			for _, want := range tt.wantContains {
				if !strings.Contains(text, want) {
					t.Errorf("ToText() does not contain %q", want)
				}
			}
		})
	}
}
