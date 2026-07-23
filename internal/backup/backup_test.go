package backup_test

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/backup"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/db/testdb"
	"github.com/rocky-ads/site/internal/encryption"
	"github.com/rocky-ads/site/internal/facet"
	"github.com/rocky-ads/site/internal/imagestore"
	"github.com/rocky-ads/site/internal/journal"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/message"
	"github.com/rocky-ads/site/internal/seed"
	"github.com/rocky-ads/site/internal/user"
)

const keyA = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
const keyB = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAE="

func TestMain(m *testing.M) {
	if err := chdirModuleRoot(); err != nil {
		panic(err)
	}
	testURL := testdb.PackageDatabaseURL("backup")
	if err := testdb.EnsureDatabase(testURL); err != nil {
		panic(err)
	}
	os.Setenv("DATABASE_URL", testURL)
	os.Setenv("DB_ENCRYPTION_KEY", keyA)
	setDBKey(keyA)
	if err := logger.Init("error", "text", os.DevNull); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func setDBKey(b64 string) {
	key, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		panic(err)
	}
	reflect.ValueOf(&config.DBEncryptionKey).Elem().Set(reflect.ValueOf(key))
}

func chdirModuleRoot() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return os.Chdir(dir)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func TestBackupRestoreCrossKey(t *testing.T) {
	if err := testdb.InitSchema(); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	if err := seed.LoadAll(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := ad.LoadCategories(); err != nil {
		t.Fatalf("load categories: %v", err)
	}

	setDBKey(keyA)
	alice, err := user.CreateUser("alice", "+15559876543", "secret")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	bob, err := user.CreateUser("bob", "+15559876544", "secret")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	price := 12000
	currency := "USD"
	adID, err := ad.CreateAd(ad.CreateInput{
		CategoryID:  5,
		UserID:      alice.ID,
		Title:       "Backup Test Car",
		Description: "Round trip restore test",
		Facets: map[string]facet.Value{
			"price": {Num: &price, Text: &currency},
		},
		ImageCount: 1,
	})
	if err != nil {
		t.Fatalf("create ad: %v", err)
	}

	var convID int
	err = db.QueryRow(`
		INSERT INTO conversations (
			ad_id, owner_id, inquirer_id, inquirer_has_unread, journal
		) VALUES ($1, $2, $3, 1, '')
		RETURNING id`,
		adID, alice.ID, bob.ID,
	).Scan(&convID)
	if err != nil {
		t.Fatalf("insert conversation: %v", err)
	}
	_, err = message.CreateMessage(convID, bob.ID, "Is this still available?")
	if err != nil {
		t.Fatalf("create message: %v", err)
	}

	imageDir := t.TempDir()
	store := imagestore.NewLocal(imageDir)
	imageData := []byte("fake-webp-image-data")
	if err := store.Put(adID, 1, "480w", imageData); err != nil {
		t.Fatalf("put image: %v", err)
	}

	outDir := t.TempDir()
	archive := filepath.Join(outDir, "cross.tar.gz")
	path, err := backup.BackupToArchive(archive, store, false, false)
	if err != nil {
		t.Fatalf("backup: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("archive missing: %v", err)
	}

	keyABytes, _ := base64.StdEncoding.DecodeString(keyA)
	setDBKey(keyB)
	restoreStore := imagestore.NewLocal(t.TempDir())
	if err := backup.RestoreFromArchive(path, restoreStore, keyABytes, false, false); err != nil {
		t.Fatalf("restore: %v", err)
	}

	var restoredID int
	err = db.QueryRow(
		`SELECT id FROM ads WHERE title = $1`, "Backup Test Car",
	).Scan(&restoredID)
	if err != nil {
		t.Fatalf("query restored ad: %v", err)
	}

	restoredAlice, err := user.GetByName("alice")
	if err != nil {
		t.Fatalf("get alice under key B: %v", err)
	}
	if restoredAlice.Name != "alice" {
		t.Fatalf("alice name = %q", restoredAlice.Name)
	}

	// Seed admin has no ads; backup must still include all users.
	admin, err := user.GetByName("admin")
	if err != nil {
		t.Fatalf("get seed admin after restore: %v", err)
	}
	if !admin.IsAdmin {
		t.Fatal("expected restored admin.is_admin")
	}
	if _, err := user.GetByName("test"); err != nil {
		t.Fatalf("get seed test after restore: %v", err)
	}

	var sealedJournal string
	err = db.QueryRow(
		`SELECT journal FROM conversations WHERE ad_id = $1`, restoredID,
	).Scan(&sealedJournal)
	if err != nil {
		t.Fatalf("query journal: %v", err)
	}
	var convID2 int
	err = db.QueryRow(
		`SELECT id FROM conversations WHERE ad_id = $1`, restoredID,
	).Scan(&convID2)
	if err != nil {
		t.Fatalf("query conv id: %v", err)
	}
	plain, err := encryption.Open(convID2, sealedJournal, config.DBEncryptionKey)
	if err != nil {
		t.Fatalf("open journal: %v", err)
	}
	content, _, ok := journal.LastMessagePreview(plain)
	if !ok {
		t.Fatal("expected message in restored journal")
	}
	if content != "Is this still available?" {
		t.Fatalf("message = %q", content)
	}

	restoredImage, err := restoreStore.Get(restoredID, 1, "480w")
	if err != nil {
		t.Fatalf("get restored image: %v", err)
	}
	if string(restoredImage) != string(imageData) {
		t.Fatal("restored image data mismatch")
	}
}

func TestDefaultArchivePath(t *testing.T) {
	ts := time.Date(2026, 7, 22, 17, 20, 45, 0, time.UTC)
	got := backup.DefaultArchivePath(ts)
	want := filepath.Join("backups", "backup-20260722-172045.tar.gz")
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestSealOpenRoundTrip(t *testing.T) {
	key, _ := base64.StdEncoding.DecodeString(keyA)
	sealed, err := encryption.Seal(42, "hello journal", key)
	if err != nil {
		t.Fatal(err)
	}
	plain, err := encryption.Open(42, sealed, key)
	if err != nil {
		t.Fatal(err)
	}
	if plain != "hello journal" {
		t.Fatalf("got %q", plain)
	}
	passthrough, err := encryption.Open(1, "not sealed", key)
	if err != nil {
		t.Fatal(err)
	}
	if passthrough != "not sealed" {
		t.Fatalf("got %q", passthrough)
	}
}
