package selfupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	maxChecksumSize = 1 << 20       // 1 MB
	maxArchiveSize  = 100 << 20     // 100 MB
	maxExtractSize  = 100 << 20     // 100 MB per file
	downloadRetries = 3             // retry count for transient HTTP errors
)

// Downloader downloads and verifies release archives.
type Downloader struct{}

// NewDownloader creates a new Downloader.
func NewDownloader() *Downloader {
	return &Downloader{}
}

// DownloadAndVerify downloads the archive and checksums, verifies integrity, and extracts the binary.
// Retries transient HTTP errors (non-200) up to downloadRetries times with exponential backoff.
func (d *Downloader) DownloadAndVerify(archiveURL, checksumURL, archiveName, destDir string) (string, error) {
	checksumData, err := httpGetWithRetry(checksumURL, maxChecksumSize)
	if err != nil {
		return "", fmt.Errorf("checksums download: %w", err)
	}

	checksums, err := ParseChecksums(checksumData)
	if err != nil {
		return "", err
	}

	expectedChecksum := checksums[archiveName]
	if expectedChecksum == "" {
		return "", fmt.Errorf("checksum not found for %s in checksums.txt", archiveName)
	}

	archiveData, err := httpGetWithRetry(archiveURL, maxArchiveSize)
	if err != nil {
		return "", fmt.Errorf("archive download: %w", err)
	}

	// @AX:NOTE: [AUTO] security-critical — SHA256 integrity verification guards against tampered binaries
	actualChecksum := fmt.Sprintf("%x", sha256.Sum256(archiveData))
	if actualChecksum != expectedChecksum {
		return "", fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
	}

	return extractBinary(archiveData, destDir)
}

// httpGetWithRetry downloads a URL with retry on non-200 responses.
// Handles CDN propagation delays after new releases.
func httpGetWithRetry(url string, maxSize int64) ([]byte, error) {
	var lastErr error
	for attempt := range downloadRetries {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt*2) * time.Second)
		}

		resp, err := http.Get(url)
		if err != nil {
			lastErr = err
			continue
		}

		data, err := io.ReadAll(io.LimitReader(resp.Body, maxSize))
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
			continue
		}

		return data, nil
	}
	return nil, fmt.Errorf("failed after %d attempts: %w", downloadRetries, lastErr)
}

// ParseChecksums parses checksums.txt format into a map.
func ParseChecksums(data []byte) (map[string]string, error) {
	result := make(map[string]string)

	for line := range strings.SplitSeq(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) >= 2 {
			result[parts[1]] = parts[0]
		}
	}

	return result, nil
}

// binaryName is the expected binary filename inside the archive.
const binaryName = "auto"

// extractBinary extracts the "auto" binary from tar.gz archive.
func extractBinary(data []byte, destDir string) (string, error) {
	gzr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		baseName := filepath.Base(header.Name)
		if header.Typeflag != tar.TypeReg || baseName != binaryName {
			continue
		}

		path := filepath.Join(destDir, baseName)
		f, err := os.Create(path)
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(f, io.LimitReader(tr, maxExtractSize)); err != nil {
			f.Close()
			return "", err
		}
		f.Close()

		if err := os.Chmod(path, os.FileMode(header.Mode)); err != nil {
			return "", err
		}

		return path, nil
	}

	return "", fmt.Errorf("binary %q not found in archive", binaryName)
}
