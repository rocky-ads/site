package handler

import (
	"mime/multipart"
	"testing"

	"github.com/rocky-ads/site/internal/config"
)

func TestValidateImageFileOversize(t *testing.T) {
	fh := &multipart.FileHeader{
		Filename: "large.png",
		Size:     config.ServerUploadLimit + 1,
	}
	fh.Header = make(map[string][]string)
	fh.Header.Set("Content-Type", "image/png")

	if err := validateImageFile(fh); err == nil {
		t.Fatal("expected oversize error")
	}
}

func TestValidateImageFileTooSmall(t *testing.T) {
	fh := &multipart.FileHeader{
		Filename: "tiny.png",
		Size:     10,
	}
	fh.Header = make(map[string][]string)
	fh.Header.Set("Content-Type", "image/png")

	if err := validateImageFile(fh); err == nil {
		t.Fatal("expected too-small error")
	}
}
