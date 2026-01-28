// Package compare provides result structures for the compare command.
package compare

import (
	"encoding/json"
	"fmt"
)

// Result holds the comparison result data.
type Result struct {
	Pass           bool    `json:"pass"`
	PixelDiffCount int     `json:"pixelDiffCount"`
	PixelDiffRatio float64 `json:"pixelDiffRatio"`
	TotalPixels    int     `json:"totalPixels"`
	Threshold      float64 `json:"threshold"`
	BaselinePath   string  `json:"baselinePath"`
	CurrentPath    string  `json:"currentPath"`
	DiffPath       string  `json:"diffPath,omitempty"`
}

// ToJSON converts the result to JSON string.
func (r *Result) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToText converts the result to human-readable text.
func (r *Result) ToText() string {
	status := "PASS"
	if !r.Pass {
		status = "FAIL"
	}

	return fmt.Sprintf(`Visual Regression Test Result
==============================
Status: %s
Pixel Diff: %d / %d (%.4f%%)
Threshold: %.4f%%
Baseline: %s
Current: %s
Diff: %s
`,
		status,
		r.PixelDiffCount,
		r.TotalPixels,
		r.PixelDiffRatio*100,
		r.Threshold*100,
		r.BaselinePath,
		r.CurrentPath,
		r.DiffPath,
	)
}
