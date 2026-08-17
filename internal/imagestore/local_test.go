package imagestore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestLocalStoreJPEG(t *testing.T) {
	dir := t.TempDir()
	s := NewLocal(dir)
	data := []byte{0xff, 0xd8, 0xff, 0xdb}

	if err := s.Put(1, 1, "480w", data); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "1", "1-480w.jpg")
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected jpg file: %v", err)
	}

	ok, err := s.Stat(1, 1, "480w")
	if err != nil || !ok {
		t.Fatalf("stat: ok=%v err=%v", ok, err)
	}
	got, err := s.Get(1, 1, "480w")
	if err != nil || string(got) != string(data) {
		t.Fatalf("get: %v %q", err, got)
	}
	u, err := s.PresignGet(1, 1, "480w", time.Minute)
	if err != nil || !strings.HasSuffix(u, "1-480w.jpg") {
		t.Fatalf("presign get: %q %v", u, err)
	}
	putURL, err := s.PresignPut(1, 2, "160w", time.Minute)
	if err != nil || !strings.Contains(putURL, "1/2-160w.jpg") {
		t.Fatalf("presign put: %q %v", putURL, err)
	}

	refs, err := s.ListAd(1)
	if err != nil || len(refs) != 1 || refs[0].Suffix != "480w" {
		t.Fatalf("list: %+v %v", refs, err)
	}
}

func TestLocalStoreAccountPicture(t *testing.T) {
	dir := t.TempDir()
	s := NewLocal(dir)
	data := []byte{0xff, 0xd8, 0xff, 0xdb}

	if err := s.PutUserAccount(9, data); err != nil {
		t.Fatal(err)
	}
	ok, err := s.StatUserAccount(9)
	if err != nil || !ok {
		t.Fatalf("stat: ok=%v err=%v", ok, err)
	}
	u, err := s.PresignGetUserAccount(9, time.Minute)
	if err != nil || !strings.HasSuffix(u, "account.jpg") {
		t.Fatalf("presign get: %q %v", u, err)
	}
	if err := s.DeleteUserAccount(9); err != nil {
		t.Fatal(err)
	}
	ok, err = s.StatUserAccount(9)
	if err != nil || ok {
		t.Fatalf("expected removed, ok=%v err=%v", ok, err)
	}
}
