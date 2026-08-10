package xai_test

import (
	"testing"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/internal/conn"
)

// Version is the single public version constant; dial metadata uses it after init.
func TestVersionSingleSourceOfTruth(t *testing.T) {
	if xai.Version == "" {
		t.Fatal("xai.Version must be non-empty")
	}
	if conn.SDKVersion != xai.Version {
		t.Fatalf("conn.SDKVersion=%q != xai.Version=%q (wire metadata must track Version)",
			conn.SDKVersion, xai.Version)
	}
	// Semver-ish: at least major.minor.patch digits
	if len(xai.Version) < 5 {
		t.Fatalf("Version looks too short: %q", xai.Version)
	}
}
