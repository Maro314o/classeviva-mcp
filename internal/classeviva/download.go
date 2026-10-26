package classeviva

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// SaveDownloadedFile writes a downloaded file to an explicit absolute path.
// Existing files are never replaced.
func SaveDownloadedFile(destinationPath string, body []byte) (string, error) {
	if destinationPath == "" {
		return "", fmt.Errorf("destination path is required")
	}
	if !filepath.IsAbs(destinationPath) {
		return "", fmt.Errorf("destination path must be absolute")
	}

	cleanPath := filepath.Clean(destinationPath)
	if err := os.MkdirAll(filepath.Dir(cleanPath), 0o700); err != nil {
		return "", fmt.Errorf("cannot create destination directory: %w", err)
	}

	file, err := os.OpenFile(cleanPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return "", fmt.Errorf("cannot create destination file without overwriting: %w", err)
	}

	if err := writeDownloadedFile(file, cleanPath, body); err != nil {
		return "", err
	}

	return cleanPath, nil
}

func writeDownloadedFile(file io.WriteCloser, destinationPath string, body []byte) error {
	cleanup := func() {
		_ = file.Close()
		_ = os.Remove(destinationPath)
	}

	written, err := file.Write(body)
	if err == nil && written != len(body) {
		err = io.ErrShortWrite
	}
	if err != nil {
		cleanup()
		return fmt.Errorf("cannot write downloaded material: %w", err)
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(destinationPath)
		return fmt.Errorf("cannot close downloaded material: %w", err)
	}
	return nil
}
