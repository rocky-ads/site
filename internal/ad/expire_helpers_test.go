package ad_test

import (
	"testing"
	"time"

	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/facet"
)

func TestInitialExpireGrant(t *testing.T) {
	now := time.Date(2026, 1, 15, 12, 0, 0, 0, time.UTC)
	got := ad.InitialExpireGrant(now)
	want := now.AddDate(0, config.AdExpireInitialMonths, 0).Sub(now)
	if got != want {
		t.Fatalf("InitialExpireGrant = %v, want %v", got, want)
	}
}

func TestHalfExpireGrant(t *testing.T) {
	full := 90 * 24 * time.Hour
	half := ad.HalfExpireGrant(full)
	if half != full/2 {
		t.Fatalf("half = %v, want %v", half, full/2)
	}
	tiny := ad.HalfExpireGrant(config.AdExpireMinGrant)
	if tiny != config.AdExpireMinGrant {
		t.Fatalf("floored half = %v, want %v", tiny, config.AdExpireMinGrant)
	}
}

func TestExpiresAtFromSaleEnd(t *testing.T) {
	got, err := ad.ExpiresAtFromSaleEnd("2026-07-04")
	if err != nil {
		t.Fatal(err)
	}
	// End of 2026-07-04 UTC is 2026-07-05 00:00; plus 1 week => 2026-07-12.
	want := time.Date(2026, 7, 12, 0, 0, 0, 0, time.UTC)
	if !got.Equal(want) {
		t.Fatalf("ExpiresAtFromSaleEnd = %v, want %v", got, want)
	}
}

func TestComputeExpiresAtPrefersSaleEnd(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	end := "2026-01-10"
	facets := map[string]facet.Value{
		"sale_end_date": {Text: &end},
	}
	got := ad.ComputeExpiresAt(facets, now, 90*24*time.Hour)
	want, err := ad.ExpiresAtFromSaleEnd(end)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Equal(want) {
		t.Fatalf("ComputeExpiresAt = %v, want %v", got, want)
	}
}
