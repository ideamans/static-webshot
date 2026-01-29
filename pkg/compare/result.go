// Package compare provides result structures for the compare command.
package compare

import (
	"encoding/json"
	"fmt"
)

// Result holds the comparison result data.
type Result struct {
	PixelDiffCount int     `json:"pixelDiffCount"`
	PixelDiffRatio float64 `json:"pixelDiffRatio"`
	DiffPercent    float64 `json:"diffPercent"`
	TotalPixels    int     `json:"totalPixels"`
	BaselinePath   string  `json:"baselinePath"`
	CurrentPath    string  `json:"currentPath"`
	DiffPath       string  `json:"diffPath,omitempty"`
}

// ToJSON converts the result to JSON string.
func (r *Result) ToJSON() (string, error) {
	// Compute diff percent for output
	out := *r
	out.DiffPercent = r.PixelDiffRatio * 100
	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToText converts the result to human-readable text.
func (r *Result) ToText() string {
	return fmt.Sprintf(`Visual Regression Test Result
==============================
Pixel Diff: %d / %d
Diff Percent: %.4f%%
Baseline: %s
Current: %s
Diff: %s
`,
		r.PixelDiffCount,
		r.TotalPixels,
		r.PixelDiffRatio*100,
		r.BaselinePath,
		r.CurrentPath,
		r.DiffPath,
	)
}
