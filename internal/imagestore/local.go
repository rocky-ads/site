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

func (s *LocalStore) Get(adID, index int, suffix string) ([]byte, error) {
	path := filepath.Join(s.baseDir, fmt.Sprintf("%d", adID), fmt.Sprintf("%d-%s.webp", index, suffix))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}
	return data, nil
}

func (s *LocalStore) DeleteAd(adID int) error {
	dir := filepath.Join(s.baseDir, fmt.Sprintf("%d", adID))
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("delete ad images: %w", err)
	}
	return nil
}
