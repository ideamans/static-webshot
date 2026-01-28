package record

import "testing"

func TestGetPreset(t *testing.T) {
	tests := []struct {
		name       string
		preset     string
		wantWidth  int
		wantHeight int
		wantMobile bool
	}{
		{
			name:       "desktop preset",
			preset:     "desktop",
			wantWidth:  1920,
			wantHeight: 1080,
			wantMobile: false,
		},
		{
			name:       "mobile preset",
			preset:     "mobile",
			wantWidth:  390,
			wantHeight: 844,
			wantMobile: true,
		},
		{
			name:       "unknown preset falls back to desktop",
			preset:     "unknown",
			wantWidth:  1920,
			wantHeight: 1080,
			wantMobile: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			preset := GetPreset(tt.preset)

			if preset.ViewportWidth != tt.wantWidth {
				t.Errorf("ViewportWidth = %d, want %d", preset.ViewportWidth, tt.wantWidth)
			}
			if preset.ViewportHeight != tt.wantHeight {
				t.Errorf("ViewportHeight = %d, want %d", preset.ViewportHeight, tt.wantHeight)
			}
			if preset.IsMobile != tt.wantMobile {
				t.Errorf("IsMobile = %v, want %v", preset.IsMobile, tt.wantMobile)
			}
		})
	}
}
