package imagestore

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

// LocalStore writes ad images under baseDir/{adID}/{index}-{suffix}.webp.
type LocalStore struct {
	baseDir string
	mu      sync.Mutex
	puts    map[string]string
}

func NewLocal(baseDir string) *LocalStore {
	return &LocalStore{
		baseDir: baseDir,
		puts:    make(map[string]string),
	}
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
	path := filepath.Join(s.baseDir, fmt.Sprintf("%d", adID),
		fmt.Sprintf("%d-%s.webp", index, suffix))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}
	return data, nil
}

func (s *LocalStore) Stat(adID, index int, suffix string) (bool, error) {
	path := filepath.Join(s.baseDir, fmt.Sprintf("%d", adID),
		fmt.Sprintf("%d-%s.webp", index, suffix))
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *LocalStore) ListAd(adID int) ([]ImageRef, error) {
	dir := filepath.Join(s.baseDir, fmt.Sprintf("%d", adID))
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list ad images: %w", err)
	}

	var refs []ImageRef
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := imageFilePattern.FindStringSubmatch(entry.Name())
		if len(matches) != 3 {
			continue
		}
		index, err := strconv.Atoi(matches[1])
		if err != nil {
			continue
		}
		refs = append(refs, ImageRef{Index: index, Suffix: matches[2]})
	}
	return refs, nil
}

func (s *LocalStore) DeleteAd(adID int) error {
	dir := filepath.Join(s.baseDir, fmt.Sprintf("%d", adID))
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("delete ad images: %w", err)
	}
	return nil
}

func (s *LocalStore) PresignPut(adID, index int, suffix string,
	expiry time.Duration) (string, error) {
	_ = expiry
	key := objectKey(adID, index, suffix)
	u := url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:9",
		Path:   "/" + key,
	}
	s.mu.Lock()
	s.puts[u.String()] = key
	s.mu.Unlock()
	return u.String(), nil
}

func (s *LocalStore) PresignGet(adID, index int, suffix string,
	expiry time.Duration) (string, error) {
	_ = expiry
	ok, err := s.Stat(adID, index, suffix)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("object not found")
	}
	return fmt.Sprintf("http://127.0.0.1/local/%d/%d-%s.webp",
		adID, index, suffix), nil
}

func (s *LocalStore) userAccountPath(userID int) string {
	return filepath.Join(s.baseDir, "users", fmt.Sprintf("%d", userID),
		"account.webp")
}

func (s *LocalStore) PutUserAccount(userID int, data []byte) error {
	path := s.userAccountPath(userID)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create user image dir: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write user account image: %w", err)
	}
	return nil
}

func (s *LocalStore) StatUserAccount(userID int) (bool, error) {
	_, err := os.Stat(s.userAccountPath(userID))
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *LocalStore) DeleteUserAccount(userID int) error {
	path := s.userAccountPath(userID)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete user account image: %w", err)
	}
	dir := filepath.Dir(path)
	_ = os.Remove(dir) // best-effort remove empty users/{id}
	return nil
}

func (s *LocalStore) PresignPutUserAccount(userID int,
	expiry time.Duration) (string, error) {
	_ = expiry
	key := userAccountObjectKey(userID)
	u := url.URL{
		Scheme: "http",
		Host:   "127.0.0.1:9",
		Path:   "/" + key,
	}
	s.mu.Lock()
	s.puts[u.String()] = key
	s.mu.Unlock()
	return u.String(), nil
}

func (s *LocalStore) PresignGetUserAccount(userID int,
	expiry time.Duration) (string, error) {
	_ = expiry
	ok, err := s.StatUserAccount(userID)
	if err != nil {
		return "", err
	}
	if !ok {
		return "", fmt.Errorf("object not found")
	}
	return fmt.Sprintf("http://127.0.0.1/local/users/%d/account.webp",
		userID), nil
}
