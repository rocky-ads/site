package ad

import (
	"time"

	"github.com/rocky-ads/site/internal/config"
	"github.com/rocky-ads/site/internal/facet"
)

// InitialExpireGrant is the lifetime granted when an ad is first created.
func InitialExpireGrant(now time.Time) time.Duration {
	return now.AddDate(0, config.AdExpireInitialMonths, 0).Sub(now)
}

// HalfExpireGrant returns the next reactivate grant (half of prev, floored).
func HalfExpireGrant(prev time.Duration) time.Duration {
	half := prev / 2
	if half < config.AdExpireMinGrant {
		return config.AdExpireMinGrant
	}
	return half
}

// ExpiresAtFromSaleEnd is end of sale_end_date (UTC) plus the sale-end delay.
func ExpiresAtFromSaleEnd(dateYYYYMMDD string) (time.Time, error) {
	day, err := time.ParseInLocation("2006-01-02", dateYYYYMMDD, time.UTC)
	if err != nil {
		return time.Time{}, err
	}
	endOfSaleDay := day.AddDate(0, 0, 1)
	return endOfSaleDay.Add(config.AdExpireSaleEndDelay), nil
}

// SaleEndDateString returns the sale_end_date facet value if present.
func SaleEndDateString(facets map[string]facet.Value) (string, bool) {
	v, ok := facets["sale_end_date"]
	if !ok {
		return "", false
	}
	s := v.DateString()
	if s == "" {
		return "", false
	}
	return s, true
}

// ComputeExpiresAt prefers sale_end_date + delay when present; otherwise now+grant.
func ComputeExpiresAt(facets map[string]facet.Value, now time.Time,
	grant time.Duration) time.Time {
	if s, ok := SaleEndDateString(facets); ok {
		if t, err := ExpiresAtFromSaleEnd(s); err == nil {
			return t
		}
	}
	return now.Add(grant)
}
