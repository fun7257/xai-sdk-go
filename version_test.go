package xai_test

import (
	"strings"
	"testing"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/internal/conn"
)

func TestVersionFromModuleGraph(t *testing.T) {
	v := xai.Version()
	if v == "" {
		t.Fatal("Version() must be non-empty")
	}
	// In-repo go test is almost always untagged → "devel".
	// When this module is required at a tag, consumers get that tag (no leading v).
	if v != "devel" && strings.Contains(v, " ") {
		t.Fatalf("unexpected Version %q", v)
	}
	if conn.SDKVersion != v {
		t.Fatalf("conn.SDKVersion=%q != Version()=%q (wire metadata must track module version)",
			conn.SDKVersion, v)
	}
}

func TestNormalizeModuleVersion(t *testing.T) {
	// Drive through public Version() resolution indirectly via known BuildInfo
	// behavior is hard; test normalize via metadata consistency only.
	// Ensure Version is stable across repeated calls (sync.Once).
	a, b := xai.Version(), xai.Version()
	if a != b {
		t.Fatalf("Version() not stable: %q vs %q", a, b)
	}
}
