package classeviva

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// CredentialsConfigured reports whether the configured credentials file exists
// and can be parsed. It never returns credential values.
func CredentialsConfigured() (bool, string, error) {
	path, err := CredentialsPath()
	if err != nil {
		return false, "", err
	}
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false, path, nil
	} else if err != nil {
		return false, path, fmt.Errorf("cannot inspect credentials file: %w", err)
	}
	if _, err := NewFromFile(path); err != nil {
		return false, path, nil
	}
	return true, path, nil
}

// SaveCredentials writes credentials atomically with private file permissions.
func SaveCredentials(uid, password string) (string, error) {
	uid = strings.TrimSpace(uid)
	if len(uid) < 2 || strings.ContainsAny(uid, "\r\n\x00") {
		return "", errors.New("inserisci un UID ClasseViva valido")
	}
	if strings.TrimSpace(password) == "" || strings.ContainsAny(password, "\r\n\x00") {
		return "", errors.New("inserisci una password valida")
	}

	path, err := CredentialsPath()
	if err != nil {
		return "", err
	}
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return "", fmt.Errorf("cannot create private configuration directory: %w", err)
	}
	if strings.TrimSpace(os.Getenv("CLASSEVIVA_CREDENTIALS_FILE")) == "" && runtime.GOOS != "windows" {
		if err := os.Chmod(directory, 0o700); err != nil {
			return "", fmt.Errorf("cannot secure configuration directory: %w", err)
		}
	}

	temp, err := os.CreateTemp(directory, ".credentials-*")
	if err != nil {
		return "", fmt.Errorf("cannot create credentials file: %w", err)
	}
	tempPath := temp.Name()
	defer os.Remove(tempPath)
	if err := temp.Chmod(0o600); err != nil {
		temp.Close()
		return "", fmt.Errorf("cannot secure credentials file: %w", err)
	}
	if _, err := fmt.Fprintf(temp, "%s\n%s\n", uid, password); err != nil {
		temp.Close()
		return "", fmt.Errorf("cannot write credentials file: %w", err)
	}
	if err := temp.Sync(); err != nil {
		temp.Close()
		return "", fmt.Errorf("cannot sync credentials file: %w", err)
	}
	if err := temp.Close(); err != nil {
		return "", fmt.Errorf("cannot close credentials file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		return "", fmt.Errorf("cannot install credentials file: %w", err)
	}
	return path, nil
}
