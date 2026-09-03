package vector

import (
	"math"
	"testing"
)

func TestDampSiteWeight(t *testing.T) {
	if got := dampSiteWeight(0); got != 0 {
		t.Fatalf("zero = %v", got)
	}
	one := dampSiteWeight(1)
	many := dampSiteWeight(20)
	if one <= 0 || many <= one {
		t.Fatalf("expected many > one > 0, got one=%v many=%v", one, many)
	}
	if many >= 20 {
		t.Fatalf("expected log damp, got %v", many)
	}
	want := float32(math.Log1p(20))
	if many != want {
		t.Fatalf("got %v, want %v", many, want)
	}
}
