package backup

import (
	"testing"
	"time"

	"github.com/rocky-ads/site/internal/ad"
	"github.com/rocky-ads/site/internal/config"
)

func TestResolveExpireFieldsFromArchive(t *testing.T) {
	created := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	grant := ad.InitialExpireGrant(created)
	explicit := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)

	t.Run("new_archive", func(t *testing.T) {
		gotAt, gotNs := resolveExpireFields(AdRow{
			CreatedAt:     created,
			ExpiresAt:     explicit,
			ExpireGrantNs: int64(grant / 2),
		}, "")
		if !gotAt.Equal(explicit) {
			t.Fatalf("expires_at = %v, want %v", gotAt, explicit)
		}
		if gotNs != int64(grant/2) {
			t.Fatalf("grant_ns = %d, want %d", gotNs, grant/2)
		}
	})

	t.Run("old_archive_default", func(t *testing.T) {
		gotAt, gotNs := resolveExpireFields(AdRow{CreatedAt: created}, "")
		wantAt := created.Add(grant)
		if !gotAt.Equal(wantAt) {
			t.Fatalf("expires_at = %v, want %v", gotAt, wantAt)
		}
		if gotNs != int64(grant) {
			t.Fatalf("grant_ns = %d, want %d", gotNs, grant)
		}
	})

	t.Run("old_archive_sale_end", func(t *testing.T) {
		gotAt, _ := resolveExpireFields(AdRow{CreatedAt: created}, "2026-06-15")
		wantAt, err := ad.ExpiresAtFromSaleEnd("2026-06-15")
		if err != nil {
			t.Fatal(err)
		}
		if !gotAt.Equal(wantAt) {
			t.Fatalf("expires_at = %v, want %v", gotAt, wantAt)
		}
	})
}

func TestResolveExpireFieldsInitialMonths(t *testing.T) {
	if config.AdExpireInitialMonths != 3 {
		t.Fatalf("AdExpireInitialMonths = %d", config.AdExpireInitialMonths)
	}
}
