package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestSaveDidacticsItemRejectsStructuredData(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "material.json")

	if _, err := saveDidacticsItem(destination, "application/json", map[string]any{"title": "metadata"}, nil); err == nil {
		t.Fatal("saveDidacticsItem() error = nil, want structured-data error")
	}
	if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("destination stat error = %v, want file not to exist", err)
	}
}

func TestSaveDidacticsItemReturnsSavedFileMetadata(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "material.pdf")
	body := []byte("sample PDF payload")

	got, err := saveDidacticsItem(destination, "application/pdf", nil, body)
	if err != nil {
		t.Fatalf("saveDidacticsItem() error = %v", err)
	}
	if got.Path != destination {
		t.Errorf("Path = %q, want %q", got.Path, destination)
	}
	if got.MIMEType != "application/pdf" {
		t.Errorf("MIMEType = %q, want application/pdf", got.MIMEType)
	}
	if got.SizeBytes != int64(len(body)) {
		t.Errorf("SizeBytes = %d, want %d", got.SizeBytes, len(body))
	}

	saved, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", destination, err)
	}
	if string(saved) != string(body) {
		t.Fatalf("saved bytes = %q, want %q", saved, body)
	}
}

func TestValidateDidacticsDestinationRequiresAbsolutePath(t *testing.T) {
	if err := validateDidacticsDestination(filepath.Join("relative", "material.pdf")); err == nil {
		t.Fatal("validateDidacticsDestination() error = nil, want relative path error")
	}
}

func TestValidateDidacticsDestinationRequiresFilePath(t *testing.T) {
	if err := validateDidacticsDestination(""); err == nil {
		t.Fatal("validateDidacticsDestination() error = nil, want missing path error")
	}
}
