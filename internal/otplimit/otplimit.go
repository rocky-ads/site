package otplimit

import (
	"fmt"
	"sync"
	"time"

	"github.com/rocky-ads/site/internal/config"
)

type phoneState struct {
	lastStart time.Time
	starts    []time.Time
}

var (
	mu      sync.Mutex
	byPhone = map[string]*phoneState{}
)

// AllowStart returns an error if this phone may not start another OTP yet.
// In-memory only; not shared across processes.
func AllowStart(phoneE64 string) error {
	if config.AllowTestRegistration {
		return nil
	}

	now := time.Now().UTC()
	mu.Lock()
	defer mu.Unlock()

	st := byPhone[phoneE64]
	if st == nil {
		st = &phoneState{}
		byPhone[phoneE64] = st
	}

	if !st.lastStart.IsZero() &&
		now.Sub(st.lastStart) < config.OTPStartMinInterval {
		wait := config.OTPStartMinInterval - now.Sub(st.lastStart)
		return fmt.Errorf(
			"please wait %d seconds before requesting another code",
			int(wait.Seconds())+1)
	}

	cutoff := now.Add(-time.Hour)
	kept := st.starts[:0]
	for _, t := range st.starts {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	st.starts = kept

	if len(st.starts) >= config.OTPStartMaxPerHour {
		return fmt.Errorf(
			"too many verification codes requested for this number. " +
				"Please try again later.")
	}

	st.lastStart = now
	st.starts = append(st.starts, now)
	return nil
}
