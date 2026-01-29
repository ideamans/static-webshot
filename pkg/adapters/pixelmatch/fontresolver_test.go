package pixelmatch

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFontDirs_ReturnsNonEmpty(t *testing.T) {
	dirs := fontDirs()
	if len(dirs) == 0 {
		t.Error("fontDirs() returned empty slice")
	}
}

func TestFontDirs_Platform(t *testing.T) {
	dirs := fontDirs()

	switch runtime.GOOS {
	case "darwin":
		found := false
		for _, d := range dirs {
			if d == "/System/Library/Fonts" {
				found = true
				break
			}
		}
		if !found {
			t.Error("fontDirs() on darwin should include /System/Library/Fonts")
		}
	case "linux":
		found := false
		for _, d := range dirs {
			if d == "/usr/share/fonts" {
				found = true
				break
			}
		}
		if !found {
			t.Error("fontDirs() on linux should include /usr/share/fonts")
		}
	case "windows":
		found := false
		for _, d := range dirs {
			if filepath.Base(d) == "Fonts" {
				found = true
				break
			}
		}
		if !found {
			t.Error("fontDirs() on windows should include a Fonts directory")
		}
	}
}

func TestDefaultFontFaces_ReturnsNonEmpty(t *testing.T) {
	faces := DefaultFontFaces()
	if len(faces) == 0 {
		t.Error("DefaultFontFaces() returned empty slice")
	}
}

func TestDefaultFontFaces_Platform(t *testing.T) {
	faces := DefaultFontFaces()

	switch runtime.GOOS {
	case "darwin":
		if faces[0] != "Hiragino Sans" {
			t.Errorf("DefaultFontFaces() on darwin: first face = %q, want %q", faces[0], "Hiragino Sans")
		}
	case "windows":
		if faces[0] != "Yu Gothic" {
			t.Errorf("DefaultFontFaces() on windows: first face = %q, want %q", faces[0], "Yu Gothic")
		}
	case "linux":
		if faces[0] != "Noto Sans CJK JP" {
			t.Errorf("DefaultFontFaces() on linux: first face = %q, want %q", faces[0], "Noto Sans CJK JP")
		}
	}
}

func TestResolveFontPath_NonExistentFace(t *testing.T) {
	result := ResolveFontPath([]string{"NonExistent Font 12345"})
	if result != "" {
		t.Errorf("ResolveFontPath() for non-existent face = %q, want empty", result)
	}
}

func TestResolveFontPath_EmptyFaces(t *testing.T) {
	result := ResolveFontPath(nil)
	if result != "" {
		t.Errorf("ResolveFontPath(nil) = %q, want empty", result)
	}

	result = ResolveFontPath([]string{})
	if result != "" {
		t.Errorf("ResolveFontPath([]) = %q, want empty", result)
	}
}

func TestResolveFontPath_UnknownFaceName(t *testing.T) {
	result := ResolveFontPath([]string{"Totally Unknown Font"})
	if result != "" {
		t.Errorf("ResolveFontPath() for unknown face = %q, want empty", result)
	}
}

func TestResolveFontPath_CaseInsensitiveLookup(t *testing.T) {
	r1 := ResolveFontPath([]string{"Arial"})
	r2 := ResolveFontPath([]string{"arial"})
	r3 := ResolveFontPath([]string{"ARIAL"})

	if r1 != r2 || r2 != r3 {
		t.Errorf("ResolveFontPath() case sensitivity: %q vs %q vs %q", r1, r2, r3)
	}
}

func TestResolveFontPath_PriorityOrder(t *testing.T) {
	result := ResolveFontPath([]string{"NonExistent Font 99999", "Arial"})
	resultDirect := ResolveFontPath([]string{"Arial"})

	if result != resultDirect {
		t.Errorf("ResolveFontPath() priority: got %q, want %q (same as direct Arial lookup)", result, resultDirect)
	}
}

func TestResolveFontPath_DefaultFaces(t *testing.T) {
	defaults := DefaultFontFaces()
	result := ResolveFontPath(defaults)

	if result == "" {
		t.Skip("No default fonts found on this system")
	}

	if _, err := os.Stat(result); err != nil {
		t.Errorf("ResolveFontPath() returned non-existent path: %q", result)
	}
}

func TestResolveFontPath_ResolvedFileIsReadable(t *testing.T) {
	defaults := DefaultFontFaces()
	result := ResolveFontPath(defaults)

	if result == "" {
		t.Skip("No default fonts found on this system")
	}

	data, err := os.ReadFile(result)
	if err != nil {
		t.Errorf("Could not read resolved font file %q: %v", result, err)
	}
	if len(data) == 0 {
		t.Errorf("Resolved font file %q is empty", result)
	}
}

func TestSearchFontFile_DirectPath(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "TestFont.ttf")
	if err := os.WriteFile(testFile, []byte("fake font data"), 0644); err != nil {
		t.Fatal(err)
	}

	result := searchFontFile(dir, "TestFont.ttf")
	if result != testFile {
		t.Errorf("searchFontFile() = %q, want %q", result, testFile)
	}
}

func TestSearchFontFile_Subdirectory(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "truetype", "noto")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	testFile := filepath.Join(subDir, "NotoSansCJK.ttc")
	if err := os.WriteFile(testFile, []byte("fake font data"), 0644); err != nil {
		t.Fatal(err)
	}

	result := searchFontFile(dir, "NotoSansCJK.ttc")
	if result != testFile {
		t.Errorf("searchFontFile() = %q, want %q", result, testFile)
	}
}

func TestSearchFontFile_NotFound(t *testing.T) {
	dir := t.TempDir()

	result := searchFontFile(dir, "NonExistent.ttf")
	if result != "" {
		t.Errorf("searchFontFile() = %q, want empty", result)
	}
}

func TestSearchFontFile_CaseInsensitive(t *testing.T) {
	dir := t.TempDir()
	testFile := filepath.Join(dir, "Arial.ttf")
	if err := os.WriteFile(testFile, []byte("fake font data"), 0644); err != nil {
		t.Fatal(err)
	}

	// searchFontFile uses strings.EqualFold, so case-insensitive match
	result := searchFontFile(dir, "arial.ttf")
	if result != testFile {
		t.Errorf("searchFontFile() = %q, want %q", result, testFile)
	}
}

func TestFontFaceFileNames_MappingExists(t *testing.T) {
	defaults := DefaultFontFaces()
	for _, face := range defaults {
		key := strings.ToLower(strings.TrimSpace(face))
		if _, ok := fontFaceFileNames[key]; !ok {
			t.Errorf("DefaultFontFaces() includes %q but fontFaceFileNames has no mapping for %q", face, key)
		}
	}
}

func TestLoadFont_DefaultResolution(t *testing.T) {
	p := New()

	// With no font path, should still return a face (either system font or basicfont)
	face := p.loadFont("", 14)
	if face == nil {
		t.Error("loadFont() returned nil")
	}
}

func TestLoadFont_ExplicitPathFallback(t *testing.T) {
	p := New()

	// With a non-existent explicit path, should fall through to system defaults
	face := p.loadFont("/nonexistent/path/font.ttf", 14)
	if face == nil {
		t.Error("loadFont() returned nil even with fallbacks")
	}
}

func TestLoadFont_SystemFont(t *testing.T) {
	defaults := DefaultFontFaces()
	resolved := ResolveFontPath(defaults)
	if resolved == "" {
		t.Skip("No default fonts found on this system")
	}

	p := New()
	// loadFont with no path should resolve to a system font (not basicfont)
	face := p.loadFont("", 14)
	if face == nil {
		t.Error("loadFont() returned nil with available system fonts")
	}
}

func TestLoadFontFile_ValidFont(t *testing.T) {
	defaults := DefaultFontFaces()
	resolved := ResolveFontPath(defaults)
	if resolved == "" {
		t.Skip("No default fonts found on this system")
	}

	p := New()
	face := p.loadFontFile(resolved, 14)
	if face == nil {
		t.Errorf("loadFontFile(%q) returned nil", resolved)
	}
}

func TestLoadFontFile_InvalidPath(t *testing.T) {
	p := New()
	face := p.loadFontFile("/nonexistent/font.ttf", 14)
	if face != nil {
		t.Error("loadFontFile() should return nil for non-existent path")
	}
}

func TestLoadFontFile_InvalidData(t *testing.T) {
	// Create a temp file with invalid font data
	dir := t.TempDir()
	badFont := filepath.Join(dir, "bad.ttf")
	if err := os.WriteFile(badFont, []byte("not a font"), 0644); err != nil {
		t.Fatal(err)
	}

	p := New()
	face := p.loadFontFile(badFont, 14)
	if face != nil {
		t.Error("loadFontFile() should return nil for invalid font data")
	}
}
