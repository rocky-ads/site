package imagestore

import (
	"testing"
	"time"
)

type countingStore struct {
	*LocalStore
	gets int
}

func (f *countingStore) PresignGet(adID, index int, suffix string,
	expiry time.Duration) (string, error) {
	f.gets++
	return f.LocalStore.PresignGet(adID, index, suffix, expiry)
}

func TestURLCacheReusesUntilTTL(t *testing.T) {
	dir := t.TempDir()
	base := NewLocal(dir)
	if err := base.Put(1, 1, "480w", []byte("jpeg")); err != nil {
		t.Fatal(err)
	}
	fake := &countingStore{LocalStore: base}
	cache, err := NewURLCache(fake)
	if err != nil {
		t.Fatal(err)
	}

	expiry := 80 * time.Millisecond

	u1, err := cache.ReusedGetURL(1, 1, "480w", expiry)
	if err != nil {
		t.Fatal(err)
	}
	cache.urls.Wait()

	u2, err := cache.ReusedGetURL(1, 1, "480w", expiry)
	if err != nil {
		t.Fatal(err)
	}
	if u1 != u2 {
		t.Fatalf("expected reused URL, got %q then %q", u1, u2)
	}
	if fake.gets != 1 {
		t.Fatalf("expected 1 mint, got %d", fake.gets)
	}

	time.Sleep(100 * time.Millisecond)

	_, err = cache.ReusedGetURL(1, 1, "480w", expiry)
	if err != nil {
		t.Fatal(err)
	}
	if fake.gets != 2 {
		t.Fatalf("expected remint after TTL, got %d mints", fake.gets)
	}
}
