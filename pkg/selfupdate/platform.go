package selfupdate

import "fmt"

// ArchiveName returns the GoReleaser archive filename for the given OS, architecture, and version.
// Windows uses .zip, all others use .tar.gz (matching GoReleaser defaults).
func ArchiveName(goos, goarch, version string) string {
	ext := "tar.gz"
	if goos == "windows" {
		ext = "zip"
	}
	return fmt.Sprintf("autopus-adk_%s_%s_%s.%s", version, goos, goarch, ext)
}
