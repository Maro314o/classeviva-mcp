package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"classeviva/internal/classeviva"
)

type DidacticsDownloadInput struct {
	ContentID       string `json:"content_id" jsonschema:"Content ID returned by list_didactics."`
	DestinationPath string `json:"destination_path" jsonschema:"Absolute local file path including the filename. Existing files are not overwritten."`
}

type DownloadedFileOutput struct {
	Path      string `json:"path" jsonschema:"Absolute path where the downloaded material was saved."`
	MIMEType  string `json:"mime_type" jsonschema:"MIME type returned by ClasseViva."`
	SizeBytes int64  `json:"size_bytes" jsonschema:"Size of the saved file in bytes."`
}

func validateDownloadDestination(destinationPath string) error {
	if strings.TrimSpace(destinationPath) == "" {
		return fmt.Errorf("destination_path is required")
	}
	if !filepath.IsAbs(destinationPath) {
		return fmt.Errorf("destination_path must be absolute and include the filename")
	}
	return nil
}

func saveDownloadedFile(destinationPath, mimeType string, body []byte) (*DownloadedFileOutput, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("ClasseViva returned an empty file")
	}

	savedPath, err := classeviva.SaveDownloadedFile(destinationPath, body)
	if err != nil {
		return nil, err
	}
	return &DownloadedFileOutput{
		Path:      savedPath,
		MIMEType:  mimeType,
		SizeBytes: int64(len(body)),
	}, nil
}
