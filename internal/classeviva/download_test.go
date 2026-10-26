package classeviva

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

type failingDownloadedFile struct {
	err error
}

func (f failingDownloadedFile) Write([]byte) (int, error) { return 1, f.err }
func (f failingDownloadedFile) Close() error              { return nil }

func TestSaveDownloadedFileWritesExactBytesToAbsolutePath(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "nested", "material.pdf")
	want := []byte("sample PDF payload")

	gotPath, err := SaveDownloadedFile(destination, want)
	if err != nil {
		t.Fatalf("SaveDownloadedFile() error = %v", err)
	}
	if gotPath != filepath.Clean(destination) {
		t.Fatalf("SaveDownloadedFile() path = %q, want %q", gotPath, filepath.Clean(destination))
	}

	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", destination, err)
	}
	if string(got) != string(want) {
		t.Fatalf("saved bytes = %q, want %q", got, want)
	}

	if runtime.GOOS != "windows" {
		dirInfo, err := os.Stat(filepath.Dir(destination))
		if err != nil {
			t.Fatalf("Stat(%q) error = %v", filepath.Dir(destination), err)
		}
		if gotMode := dirInfo.Mode().Perm(); gotMode != 0o700 {
			t.Errorf("directory permissions = %04o, want 0700", gotMode)
		}

		info, err := os.Stat(destination)
		if err != nil {
			t.Fatalf("Stat(%q) error = %v", destination, err)
		}
		if gotMode := info.Mode().Perm(); gotMode != 0o600 {
			t.Errorf("file permissions = %04o, want 0600", gotMode)
		}
	}
}

func TestSaveDownloadedFileRejectsRelativePath(t *testing.T) {
	if _, err := SaveDownloadedFile(filepath.Join("relative", "material.pdf"), []byte("payload")); err == nil {
		t.Fatal("SaveDownloadedFile() error = nil, want relative path error")
	}
}

func TestSaveDownloadedFileDoesNotOverwriteExistingFile(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "material.pdf")
	want := []byte("keep this file")
	if err := os.WriteFile(destination, want, 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", destination, err)
	}

	if _, err := SaveDownloadedFile(destination, []byte("replacement")); err == nil {
		t.Fatal("SaveDownloadedFile() error = nil, want existing destination error")
	}

	got, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", destination, err)
	}
	if string(got) != string(want) {
		t.Fatalf("existing file bytes = %q, want unchanged %q", got, want)
	}
}

func TestSaveDownloadedFileRejectsEmptyPath(t *testing.T) {
	if _, err := SaveDownloadedFile("", []byte("payload")); err == nil {
		t.Fatal("SaveDownloadedFile() error = nil, want empty path error")
	}
}

func TestWriteDownloadedFileRemovesPartialFileOnWriteFailure(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "partial.pdf")
	if err := os.WriteFile(destination, []byte("partial"), 0o600); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", destination, err)
	}

	err := writeDownloadedFile(failingDownloadedFile{err: errors.New("disk full")}, destination, []byte("payload"))
	if err == nil {
		t.Fatal("writeDownloadedFile() error = nil, want write failure")
	}
	if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Stat(%q) error = %v, want partial file to be removed", destination, err)
	}
}
