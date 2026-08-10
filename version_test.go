package xai_test

import (
	"strings"
	"testing"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/internal/conn"
)

// Version() comes from the module graph; wire metadata must match after init.
// normalizeModuleVersion cases: version_internal_test.go
func TestVersionFromModuleGraph(t *testing.T) {
	v := xai.Version()
	if v == "" {
		t.Fatal("Version() must be non-empty")
	}
	// In-repo go test is almost always untagged → "devel".
	if v != "devel" && strings.Contains(v, " ") {
		t.Fatalf("unexpected Version %q", v)
	}
	if conn.SDKVersion != v {
		t.Fatalf("conn.SDKVersion=%q != Version()=%q", conn.SDKVersion, v)
	}
	if a, b := xai.Version(), xai.Version(); a != b {
		t.Fatalf("Version() not stable: %q vs %q", a, b)
	}
}
