package otplimit

import (
	"reflect"
	"testing"
	"time"

	"github.com/rocky-ads/site/internal/config"
)

func TestAllowStartCooldown(t *testing.T) {
	origAllow := config.AllowTestRegistration
	reflect.ValueOf(&config.AllowTestRegistration).Elem().SetBool(false)
	t.Cleanup(func() {
		reflect.ValueOf(&config.AllowTestRegistration).Elem().SetBool(origAllow)
		mu.Lock()
		byPhone = map[string]*phoneState{}
		mu.Unlock()
	})

	phone := "+15559875001"
	if err := AllowStart(phone); err != nil {
		t.Fatalf("first start: %v", err)
	}
	if err := AllowStart(phone); err == nil {
		t.Fatal("second start within interval should fail")
	}
}

func TestAllowStartHourlyCap(t *testing.T) {
	origAllow := config.AllowTestRegistration
	reflect.ValueOf(&config.AllowTestRegistration).Elem().SetBool(false)
	t.Cleanup(func() {
		reflect.ValueOf(&config.AllowTestRegistration).Elem().SetBool(origAllow)
		mu.Lock()
		byPhone = map[string]*phoneState{}
		mu.Unlock()
	})

	phone := "+15559875002"
	now := time.Now().UTC()
	mu.Lock()
	starts := make([]time.Time, config.OTPStartMaxPerHour)
	for i := range starts {
		starts[i] = now.Add(-time.Duration(i+1) * time.Minute)
	}
	byPhone[phone] = &phoneState{
		lastStart: now.Add(-2 * config.OTPStartMinInterval),
		starts:    starts,
	}
	mu.Unlock()

	if err := AllowStart(phone); err == nil {
		t.Fatal("hourly cap should block")
	}
}

func TestAllowStartSkippedWhenTestRegistration(t *testing.T) {
	origAllow := config.AllowTestRegistration
	reflect.ValueOf(&config.AllowTestRegistration).Elem().SetBool(true)
	t.Cleanup(func() {
		reflect.ValueOf(&config.AllowTestRegistration).Elem().SetBool(origAllow)
	})
	for i := 0; i < 20; i++ {
		if err := AllowStart("+15559875003"); err != nil {
			t.Fatalf("iteration %d: %v", i, err)
		}
	}
}
