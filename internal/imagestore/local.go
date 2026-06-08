package imagestore

import (
	"fmt"
	"os"
	"path/filepath"
)

// LocalStore writes ad images under baseDir/{adID}/{index}-{suffix}.webp.
type LocalStore struct {
	baseDir string
}

func NewLocal(baseDir string) *LocalStore {
	return &LocalStore{baseDir: baseDir}
}

func (s *LocalStore) Put(adID, index int, suffix string, data []byte) error {
	dir := filepath.Join(s.baseDir, fmt.Sprintf("%d", adID))
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("create image dir: %w", err)
	}
	path := filepath.Join(dir, fmt.Sprintf("%d-%s.webp", index, suffix))
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write image: %w", err)
	}
	return nil
}
