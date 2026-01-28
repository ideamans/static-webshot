package chromebrowser

import (
	"strings"
	"testing"
)

func TestDisableAnimationsCSS(t *testing.T) {
	// Verify it contains key CSS rules
	if !strings.Contains(DisableAnimationsCSS, "animation: none") {
		t.Error("DisableAnimationsCSS should contain 'animation: none'")
	}
	if !strings.Contains(DisableAnimationsCSS, "transition: none") {
		t.Error("DisableAnimationsCSS should contain 'transition: none'")
	}
}

func TestGenerateMockTimeScript(t *testing.T) {
	tests := []struct {
		name         string
		fixedTime    string
		wantContains []string
	}{
		{
			name:      "generates script with fixed time and random",
			fixedTime: "2024-01-01T00:00:00Z",
			wantContains: []string{
				"2024-01-01T00:00:00Z",
				"Date",
				"Math.random",
				"return 0.5",
				"performance.now",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			script := GenerateMockTimeScript(tt.fixedTime)

			for _, want := range tt.wantContains {
				if !strings.Contains(script, want) {
					t.Errorf("GenerateMockTimeScript() does not contain %q", want)
				}
			}
		})
	}
}

func TestGetAllDeterministicScripts(t *testing.T) {
	tests := []struct {
		name         string
		mockTime     string
		wantContains []string
		wantMissing  []string
	}{
		{
			name:     "without mock time",
			mockTime: "",
			wantContains: []string{
				"IntersectionObserver",
				"scrollTo",
				"animate",
			},
			wantMissing: []string{
				"fixedTimestamp",
			},
		},
		{
			name:     "with mock time",
			mockTime: "2024-01-01T00:00:00Z",
			wantContains: []string{
				"fixedTimestamp",
				"IntersectionObserver",
				"scrollTo",
				"animate",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scripts := GetAllDeterministicScripts(tt.mockTime)

			for _, want := range tt.wantContains {
				if !strings.Contains(scripts, want) {
					t.Errorf("GetAllDeterministicScripts() does not contain %q", want)
				}
			}

			for _, notWant := range tt.wantMissing {
				if strings.Contains(scripts, notWant) {
					t.Errorf("GetAllDeterministicScripts() should not contain %q", notWant)
				}
			}
		})
	}
}
