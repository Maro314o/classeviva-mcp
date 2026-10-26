package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

func TestSaveDownloadedFileRejectsEmptyBody(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "material.json")

	if _, err := saveDownloadedFile(destination, "application/json", nil); err == nil {
		t.Fatal("saveDownloadedFile() error = nil, want empty-file error")
	}
	if _, err := os.Stat(destination); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("destination stat error = %v, want file not to exist", err)
	}
}

func TestSaveDownloadedFileReturnsSavedFileMetadata(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "material.pdf")
	body := []byte("sample PDF payload")

	got, err := saveDownloadedFile(destination, "application/pdf", body)
	if err != nil {
		t.Fatalf("saveDownloadedFile() error = %v", err)
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

func TestSaveDownloadedFileWritesRawJSONBytes(t *testing.T) {
	destination := filepath.Join(t.TempDir(), "material.json")
	body := []byte(`{"title":"Materiale didattico"}`)

	got, err := saveDownloadedFile(destination, "application/json", body)
	if err != nil {
		t.Fatalf("saveDownloadedFile() error = %v", err)
	}
	if got.Path != destination || got.MIMEType != "application/json" || got.SizeBytes != int64(len(body)) {
		t.Fatalf("saved metadata = %#v, want path/MIME/size for raw JSON file", got)
	}

	saved, err := os.ReadFile(destination)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", destination, err)
	}
	if !bytes.Equal(saved, body) {
		t.Fatalf("saved bytes = %q, want %q", saved, body)
	}
}

func TestValidateDownloadDestinationRequiresAbsolutePath(t *testing.T) {
	if err := validateDownloadDestination(filepath.Join("relative", "material.pdf")); err == nil {
		t.Fatal("validateDownloadDestination() error = nil, want relative path error")
	}
}

func TestValidateDownloadDestinationRequiresPath(t *testing.T) {
	if err := validateDownloadDestination(""); err == nil {
		t.Fatal("validateDownloadDestination() error = nil, want missing path error")
	}
}

func TestDidacticsDownloadSchemaRequiresDestination(t *testing.T) {
	schema, err := jsonschema.For[DidacticsDownloadInput](nil)
	if err != nil {
		t.Fatalf("jsonschema.For() error = %v", err)
	}
	var contentIDRequired, destinationRequired bool
	for _, required := range schema.Required {
		if required == "content_id" {
			contentIDRequired = true
		}
		if required == "destination_path" {
			destinationRequired = true
		}
	}
	if !contentIDRequired || !destinationRequired {
		t.Fatalf("required fields = %v, want content_id and destination_path", schema.Required)
	}
}

func TestNoticeAttachmentSchemaRequiresDestination(t *testing.T) {
	schema, err := jsonschema.For[NoticeAttachmentInput](nil)
	if err != nil {
		t.Fatalf("jsonschema.For() error = %v", err)
	}
	for _, required := range schema.Required {
		if required == "destination_path" {
			return
		}
	}
	t.Fatalf("required fields = %v, want destination_path", schema.Required)
}
