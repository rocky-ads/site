package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/rocky-ads/site/cmd/init_db/seed"
	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/db/testdb"
	"github.com/rocky-ads/site/internal/facet"
	"github.com/rocky-ads/site/internal/imagestore"
	"github.com/rocky-ads/site/internal/journal"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/message"
	"github.com/rocky-ads/site/internal/user"
)

const testUserEncryptionKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

func TestMain(m *testing.M) {
	if err := chdirModuleRoot(); err != nil {
		panic(err)
	}
	testURL := testdb.PackageDatabaseURL("backup_db")
	if err := testdb.EnsureDatabase(testURL); err != nil {
		panic(err)
	}
	os.Setenv("DATABASE_URL", testURL)

	os.Setenv("USER_ENCRYPTION_KEY", testUserEncryptionKey)
	if key, err := base64.StdEncoding.DecodeString(testUserEncryptionKey); err == nil {
		reflect.ValueOf(&config.UserEncryptionKey).Elem().Set(reflect.ValueOf(key))
	}
	if err := logger.Init("info", "text", ""); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
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

func TestBackupRestoreRoundTrip(t *testing.T) {
	if err := testdb.InitSchema(); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	if err := seed.LoadAll(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := ad.LoadCategories(); err != nil {
		t.Fatalf("load categories: %v", err)
	}

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

	var testUserID int
	err = db.QueryRow(
		`SELECT id FROM users WHERE name_hash = $1`,
		db.HashString("test"),
	).Scan(&testUserID)
	if err != nil {
		t.Fatalf("lookup test user: %v", err)
	}

	clickedAt := time.Now().UTC().Truncate(time.Second)
	_, err = db.Exec(
		`INSERT INTO bookmarks (user_id, ad_id, bookmarked_at)
		 VALUES ($1, $2, $3)`,
		testUserID, adID, clickedAt,
	)
	if err != nil {
		t.Fatalf("insert bookmark: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO user_ad_clicks (ad_id, user_id, click_count, last_clicked_at)
		 VALUES ($1, $2, 3, $3)`,
		adID, testUserID, clickedAt,
	)
	if err != nil {
		t.Fatalf("insert click: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO user_ad_image_clicks (
			ad_id, user_id, image_index, click_count, last_clicked_at
		) VALUES ($1, $2, 1, 2, $3)`,
		adID, testUserID, clickedAt,
	)
	if err != nil {
		t.Fatalf("insert image click: %v", err)
	}

	var convID int
	err = db.QueryRow(`
		INSERT INTO conversations (
			ad_id, owner_id, inquirer_id, inquirer_has_unread
		) VALUES ($1, $2, $3, 1)
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

	backupArchive := filepath.Join(t.TempDir(), "backup")
	if err := backupToArchive(backupArchive, store, false, false); err != nil {
		t.Fatalf("backup: %v", err)
	}
	archivePath := resolveArchivePath(backupArchive)
	if _, err := os.Stat(archivePath); err != nil {
		t.Fatalf("archive missing: %v", err)
	}

	inspectDir := t.TempDir()
	if err := extractTarGz(archivePath, inspectDir); err != nil {
		t.Fatalf("extract for inspect: %v", err)
	}
	var backedAds []AdRow
	if err := readJSON(filepath.Join(inspectDir, fileAds), &backedAds); err != nil {
		t.Fatalf("read backed ads: %v", err)
	}
	var backedAd *AdRow
	for i := range backedAds {
		if backedAds[i].Title == "Backup Test Car" {
			backedAd = &backedAds[i]
			break
		}
	}
	if backedAd == nil {
		t.Fatalf("Backup Test Car missing from %d backed ads", len(backedAds))
	}

	restoreStore := imagestore.NewLocal(t.TempDir())
	if err := restoreFromArchive(backupArchive, restoreStore, false, false); err != nil {
		t.Fatalf("restore: %v", err)
	}

	var restoredID int
	err = db.QueryRow(
		`SELECT id FROM ads WHERE title = $1`, "Backup Test Car",
	).Scan(&restoredID)
	if err != nil {
		t.Fatalf("query restored ad: %v", err)
	}

	err = db.QueryRow(
		`SELECT id FROM users WHERE name_hash = $1`,
		db.HashString("test"),
	).Scan(&testUserID)
	if err != nil {
		t.Fatalf("lookup restored test user: %v", err)
	}

	var facetPrice int
	err = db.QueryRow(
		`SELECT num FROM ad_facets WHERE ad_id = $1 AND key = 'price'`,
		restoredID,
	).Scan(&facetPrice)
	if err != nil {
		t.Fatalf("query facet: %v", err)
	}
	if facetPrice != 12000 {
		t.Fatalf("price = %d, want 12000", facetPrice)
	}

	var bookmarkCount int
	err = db.QueryRow(
		`SELECT COUNT(*) FROM bookmarks WHERE ad_id = $1`, restoredID,
	).Scan(&bookmarkCount)
	if err != nil {
		t.Fatalf("count bookmarks: %v", err)
	}
	if bookmarkCount != 1 {
		t.Fatalf("bookmarks = %d, want 1", bookmarkCount)
	}

	var clickCount int
	err = db.QueryRow(
		`SELECT click_count FROM user_ad_clicks WHERE ad_id = $1 AND user_id = $2`,
		restoredID, testUserID,
	).Scan(&clickCount)
	if err != nil {
		t.Fatalf("query clicks: %v", err)
	}
	if clickCount != 3 {
		t.Fatalf("click_count = %d, want 3", clickCount)
	}

	var msgContent string
	err = db.QueryRow(`
		SELECT c.journal FROM conversations c
		WHERE c.ad_id = $1`,
		restoredID,
	).Scan(&msgContent)
	if err != nil {
		t.Fatalf("query conversation journal: %v", err)
	}
	content, _, ok := journal.LastMessagePreview(msgContent)
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

	var hasEmbedding bool
	err = db.QueryRow(
		`SELECT embedding IS NOT NULL FROM ads WHERE id = $1`, restoredID,
	).Scan(&hasEmbedding)
	if err != nil {
		t.Fatalf("query embedding: %v", err)
	}
	if hasEmbedding {
		t.Fatal("expected NULL embedding after restore")
	}
}
