package cost_test

import (
	"testing"

	"github.com/fun7257/xai-sdk-go/internal/cost"
)

func TestFromTicks(t *testing.T) {
	if _, ok := cost.FromTicks(0, false); ok {
		t.Fatal("expected missing")
	}
	usd, ok := cost.FromTicks(10_000_000_000, true)
	if !ok || usd != 1.0 {
		t.Fatalf("got %v %v", usd, ok)
	}
}
