package ad_test

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/db"
	"github.com/rocky-ads/site/internal/db/testdb"
	"github.com/rocky-ads/site/internal/facet"
	"github.com/rocky-ads/site/internal/logger"
	"github.com/rocky-ads/site/internal/seed"
	"github.com/rocky-ads/site/internal/user"
)

const testDBEncryptionKey = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

func TestMain(m *testing.M) {
	if err := chdirModuleRoot(); err != nil {
		panic(err)
	}
	testURL := testdb.PackageDatabaseURL("ad")
	if err := testdb.EnsureDatabase(testURL); err != nil {
		panic(err)
	}
	os.Setenv("DATABASE_URL", testURL)
	os.Setenv("DB_ENCRYPTION_KEY", testDBEncryptionKey)
	if key, err := base64.StdEncoding.DecodeString(testDBEncryptionKey); err == nil {
		reflect.ValueOf(&config.DBEncryptionKey).Elem().Set(reflect.ValueOf(key))
	}
	if err := logger.Init("error", "text", ""); err != nil {
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

func resetAdDB(t *testing.T) {
	t.Helper()
	if err := testdb.InitSchema(); err != nil {
		t.Fatalf("init schema: %v", err)
	}
	if err := seed.LoadAll(); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := ad.LoadCategories(); err != nil {
		t.Fatalf("load categories: %v", err)
	}
}

func createTestAd(t *testing.T, userID int, title string) int {
	t.Helper()
	price := 1000
	currency := "USD"
	adID, err := ad.CreateAd(ad.CreateInput{
		CategoryID:  5,
		UserID:      userID,
		Title:       title,
		Description: "desc",
		Facets: map[string]facet.Value{
			"price": {Num: &price, Text: &currency},
		},
	})
	if err != nil {
		t.Fatalf("create ad %q: %v", title, err)
	}
	return adID
}

func TestActivateResetsCreatedAt(t *testing.T) {
	resetAdDB(t)

	u, err := user.CreateUser("expireowner", "+15559872001", "password1")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	adID := createTestAd(t, u.ID, "Activate bump")

	oldCreated := time.Now().UTC().AddDate(0, -1, 0)
	_, err = db.Exec(`UPDATE ads SET created_at = $1 WHERE id = $2`,
		oldCreated, adID)
	if err != nil {
		t.Fatalf("backdate created_at: %v", err)
	}
	if err := ad.Pause(adID); err != nil {
		t.Fatalf("pause: %v", err)
	}

	before := time.Now().UTC().Add(-time.Second)
	if err := ad.Activate(adID); err != nil {
		t.Fatalf("activate: %v", err)
	}
	after := time.Now().UTC().Add(time.Second)

	a, err := ad.GetAd(u.ID, adID, time.UTC)
	if err != nil {
		t.Fatalf("get ad: %v", err)
	}
	if !a.IsActive() {
		t.Fatal("expected ad to be active")
	}
	if a.CreatedAt.Before(before) || a.CreatedAt.After(after) {
		t.Fatalf("created_at = %v, want between %v and %v",
			a.CreatedAt, before, after)
	}
}

func TestListAdsDueToExpire(t *testing.T) {
	resetAdDB(t)

	u, err := user.CreateUser("expirelist", "+15559872002", "password1")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	freshID := createTestAd(t, u.ID, "Fresh ad")
	staleID := createTestAd(t, u.ID, "Stale ad")

	cutoff := time.Now().UTC().AddDate(0, -config.AdExpireAfterMonths, 0).
		Add(-time.Hour)
	_, err = db.Exec(`UPDATE ads SET created_at = $1 WHERE id = $2`,
		cutoff, staleID)
	if err != nil {
		t.Fatalf("backdate stale: %v", err)
	}

	due, err := ad.ListAdsDueToExpire()
	if err != nil {
		t.Fatalf("list due: %v", err)
	}
	foundStale := false
	for _, row := range due {
		if row.ID == freshID {
			t.Fatal("fresh ad should not be due to expire")
		}
		if row.ID == staleID {
			foundStale = true
			if row.UserID != u.ID {
				t.Fatalf("user_id = %d, want %d", row.UserID, u.ID)
			}
		}
	}
	if !foundStale {
		t.Fatal("expected stale ad in due list")
	}

	if err := ad.Pause(staleID); err != nil {
		t.Fatalf("pause stale: %v", err)
	}
	due, err = ad.ListAdsDueToExpire()
	if err != nil {
		t.Fatalf("list due after pause: %v", err)
	}
	for _, row := range due {
		if row.ID == staleID {
			t.Fatal("paused ad should not be due to expire")
		}
	}
}
