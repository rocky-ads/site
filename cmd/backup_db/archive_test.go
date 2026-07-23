package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCreateExtractTarGzRoundTrip(t *testing.T) {
	src := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "images", "0"), 0755); err != nil {
		t.Fatal(err)
	}
	want := []byte("hello-image")
	if err := os.WriteFile(
		filepath.Join(src, "images", "0", "1-480w.webp"), want, 0644,
	); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(
		filepath.Join(src, "manifest.json"), []byte("{}\n"), 0644,
	); err != nil {
		t.Fatal(err)
	}

	archive := filepath.Join(t.TempDir(), "out.tar.gz")
	if err := createTarGz(archive, src); err != nil {
		t.Fatalf("create: %v", err)
	}

	dest := t.TempDir()
	if err := extractTarGz(archive, dest); err != nil {
		t.Fatalf("extract: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dest, "images", "0", "1-480w.webp"))
	if err != nil {
		t.Fatalf("read extracted: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	if _, err := os.Stat(filepath.Join(dest, "manifest.json")); err != nil {
		t.Fatalf("manifest: %v", err)
	}
}

func TestResolveArchivePath(t *testing.T) {
	if got := resolveArchivePath("prod"); got != "prod.tar.gz" {
		t.Fatalf("got %q", got)
	}
	if got := resolveArchivePath("prod.tar.gz"); got != "prod.tar.gz" {
		t.Fatalf("got %q", got)
	}
}

func TestDefaultBackupArchiveName(t *testing.T) {
	ts := time.Date(2026, 7, 22, 17, 20, 45, 0, time.UTC)
	got := defaultBackupArchiveName(ts)
	want := "backup-20260722-172045.tar.gz"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
