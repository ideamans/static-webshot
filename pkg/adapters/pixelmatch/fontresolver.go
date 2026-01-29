package pixelmatch

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// fontDirs returns OS-specific font directories to search.
func fontDirs() []string {
	home, _ := os.UserHomeDir()

	switch runtime.GOOS {
	case "darwin":
		dirs := []string{
			"/System/Library/Fonts",
			"/Library/Fonts",
		}
		if home != "" {
			dirs = append(dirs, filepath.Join(home, "Library", "Fonts"))
		}
		return dirs

	case "windows":
		windir := os.Getenv("WINDIR")
		if windir == "" {
			windir = `C:\Windows`
		}
		dirs := []string{
			filepath.Join(windir, "Fonts"),
		}
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			dirs = append(dirs, filepath.Join(localAppData, "Microsoft", "Windows", "Fonts"))
		}
		return dirs

	case "linux":
		dirs := []string{
			"/usr/share/fonts",
			"/usr/local/share/fonts",
		}
		if home != "" {
			dirs = append(dirs, filepath.Join(home, ".fonts"))
			dirs = append(dirs, filepath.Join(home, ".local", "share", "fonts"))
		}
		return dirs

	default:
		return nil
	}
}

// fontFaceFileNames maps normalized font face names (lowercase) to possible file names.
var fontFaceFileNames = map[string][]string{
	// Japanese fonts - macOS
	"hiragino sans":             {"ヒラギノ角ゴシック W3.ttc", "HiraginoSans-W3.ttc"},
	"hiragino kaku gothic pron": {"ヒラギノ角ゴ ProN W3.otf", "HiraKakuProN-W3.otf"},
	"hiragino kaku gothic pro":  {"ヒラギノ角ゴ Pro W3.otf", "HiraKakuPro-W3.otf"},

	// Japanese fonts - Windows
	"yu gothic":    {"YuGothR.ttc", "YuGothM.ttc", "YuGothB.ttc"},
	"yu mincho":    {"YuMincho.ttc"},
	"meiryo":       {"meiryo.ttc", "Meiryo.ttc"},
	"ms gothic":    {"msgothic.ttc"},
	"ms mincho":    {"msmincho.ttc"},
	"biz udgothic": {"BIZ-UDGothicR.ttc", "BIZUDGothic-Regular.ttf"},
	"biz udmincho": {"BIZ-UDMinchoM.ttc", "BIZUDMincho-Regular.ttf"},

	// Japanese fonts - Linux
	"noto sans cjk jp":  {"NotoSansCJK-Regular.ttc", "NotoSansCJKjp-Regular.otf", "NotoSansCJKjp-Regular.ttf"},
	"noto sans jp":      {"NotoSansJP-Regular.otf", "NotoSansJP-Regular.ttf", "NotoSansJP[wght].ttf"},
	"noto serif cjk jp": {"NotoSerifCJK-Regular.ttc", "NotoSerifCJKjp-Regular.otf"},
	"ipa gothic":        {"ipag.ttf", "IPAGothic.ttf"},
	"ipaex gothic":      {"ipaexg.ttf", "IPAexGothic.ttf"},
	"ipa mincho":        {"ipam.ttf", "IPAMincho.ttf"},
	"ipaex mincho":      {"ipaexm.ttf", "IPAexMincho.ttf"},
	"vl gothic":         {"VL-Gothic-Regular.ttf"},
	"takao gothic":      {"TakaoGothic.ttf"},

	// Cross-platform fonts
	"arial":           {"Arial.ttf", "arial.ttf"},
	"helvetica":       {"Helvetica.ttc", "Helvetica.ttf"},
	"dejavu sans":     {"DejaVuSans.ttf"},
	"liberation sans": {"LiberationSans-Regular.ttf"},
	"roboto":          {"Roboto-Regular.ttf", "Roboto[wdth,wght].ttf"},
}

// DefaultFontFaces returns the default font face candidates for the current OS.
// These are ordered by preference and include Japanese-capable fonts.
func DefaultFontFaces() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"Hiragino Sans",
			"Hiragino Kaku Gothic ProN",
			"Arial",
			"Helvetica",
		}
	case "windows":
		return []string{
			"Yu Gothic",
			"BIZ UDGothic",
			"Meiryo",
			"MS Gothic",
			"Arial",
		}
	case "linux":
		return []string{
			"Noto Sans CJK JP",
			"Noto Sans JP",
			"IPAex Gothic",
			"IPA Gothic",
			"VL Gothic",
			"Takao Gothic",
			"DejaVu Sans",
			"Liberation Sans",
		}
	default:
		return []string{
			"Arial",
			"DejaVu Sans",
			"Liberation Sans",
		}
	}
}

// errFontFound is a sentinel error to stop filepath.WalkDir early.
var errFontFound = errors.New("font found")

// ResolveFontPath searches for a font matching one of the given face names
// in OS-specific font directories. Returns the path to the first found font,
// or an empty string if none is found.
func ResolveFontPath(faces []string) string {
	dirs := fontDirs()
	if len(dirs) == 0 {
		return ""
	}

	for _, face := range faces {
		key := strings.ToLower(strings.TrimSpace(face))
		fileNames, ok := fontFaceFileNames[key]
		if !ok {
			continue
		}

		for _, dir := range dirs {
			for _, fileName := range fileNames {
				// Try direct path first (fast path)
				direct := filepath.Join(dir, fileName)
				if _, err := os.Stat(direct); err == nil {
					return direct
				}

				// Search subdirectories
				found := searchFontFile(dir, fileName)
				if found != "" {
					return found
				}
			}
		}
	}

	return ""
}

// searchFontFile recursively searches for a font file by name in the given directory.
func searchFontFile(dir, targetName string) string {
	var found string
	_ = filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.IsDir() && strings.EqualFold(d.Name(), targetName) {
			found = p
			return errFontFound
		}
		return nil
	})
	return found
}
