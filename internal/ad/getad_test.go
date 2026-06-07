package ad

import (
	"testing"
	"time"

	"github.com/rocky-ads/site/internal/db"
)

func TestGetAdNullLocationID(t *testing.T) {
	if err := db.Init("project.db"); err != nil {
		t.Skip("no project.db:", err)
	}
	defer db.Close()
	if err := LoadCategories(); err != nil {
		t.Fatal(err)
	}
	loc, _ := time.LoadLocation("America/Los_Angeles")
	_, err := GetAd(0, 990, loc)
	if err != nil {
		t.Fatalf("GetAd(990) failed: %v", err)
	}
}
