package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func resolveArchivePath(path string) string {
	if strings.HasSuffix(path, ".tar.gz") {
		return path
	}
	return path + ".tar.gz"
}

func createTarGz(archivePath, srcDir string) error {
	if err := os.MkdirAll(filepath.Dir(archivePath), 0755); err != nil {
		return fmt.Errorf("create archive dir: %w", err)
	}
	f, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}
	defer f.Close()

	gz := gzip.NewWriter(f)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if path == srcDir {
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = rel
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rf, err := os.Open(path)
		if err != nil {
			return err
		}
		defer rf.Close()
		_, err = io.Copy(tw, rf)
		return err
	})
	if err != nil {
		return fmt.Errorf("write archive: %w", err)
	}
	return nil
}

func extractTarGz(archivePath, destDir string) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("open archive: %w", err)
	}
	defer f.Close()

	gz, err := gzip.NewReader(f)
	if err != nil {
		return fmt.Errorf("gzip reader: %w", err)
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	destAbs, err := filepath.Abs(destDir)
	if err != nil {
		return fmt.Errorf("abs dest: %w", err)
	}
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read archive: %w", err)
		}
		target := filepath.Join(destDir, filepath.FromSlash(hdr.Name))
		targetAbs, err := filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("abs target: %w", err)
		}
		if !strings.HasPrefix(targetAbs, destAbs+string(os.PathSeparator)) &&
			targetAbs != destAbs {
			return fmt.Errorf("invalid archive path: %s", hdr.Name)
		}
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("mkdir %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return fmt.Errorf("mkdir parent %s: %w", target, err)
			}
			wf, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
			if err != nil {
				return fmt.Errorf("create %s: %w", target, err)
			}
			if _, err := io.Copy(wf, tr); err != nil {
				wf.Close()
				return fmt.Errorf("extract %s: %w", target, err)
			}
			if err := wf.Close(); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported archive entry type %v: %s",
				hdr.Typeflag, hdr.Name)
		}
	}
	return nil
}
