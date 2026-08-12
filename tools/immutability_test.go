package tools_test

import (
	"strings"
	"testing"

	"github.com/fun7257/xai-sdk-go/tools"
)

// Pre-encoded parameters must be valid JSON.
func TestFunctionRejectsInvalidJSON(t *testing.T) {
	if _, err := tools.Function("f", "d", `{"broken":`); err == nil {
		t.Fatal("expected invalid JSON error for string input")
	}
	if _, err := tools.Function("f", "d", []byte("not json")); err == nil {
		t.Fatal("expected invalid JSON error for []byte input")
	}
	if _, err := tools.Function("f", "d", `{"type":"object"}`); err != nil {
		t.Fatalf("valid JSON rejected: %v", err)
	}
	if _, err := tools.Function("f", "d", map[string]string{"type": "object"}); err != nil {
		t.Fatalf("marshalable input rejected: %v", err)
	}
}

// The headers map must be copied at option time.
func TestMCPHeadersCopied(t *testing.T) {
	h := map[string]string{"x-a": "1"}
	tool := tools.MCP("https://mcp.example", tools.WithMCPHeaders(h))
	h["x-a"] = "mutated"
	h["x-b"] = "2"
	got := tool.GetMcp().GetExtraHeaders()
	if got["x-a"] != "1" || len(got) != 1 {
		t.Fatalf("headers not isolated from caller mutations: %v", got)
	}
}

// CollectionsSearch never returns nil and copies the limit value.
func TestCollectionsSearchShape(t *testing.T) {
	lim := int32(5)
	tool := tools.CollectionsSearch([]string{"c1"}, &lim, "hint")
	if tool == nil || tool.GetCollectionsSearch() == nil {
		t.Fatal("nil tool")
	}
	lim = 99 // caller mutation must not affect the built tool
	cs := tool.GetCollectionsSearch()
	if cs.GetLimit() != 5 || cs.GetInstructions() != "hint" {
		t.Fatalf("wire=%+v", cs)
	}
	if !strings.Contains(strings.Join(cs.GetCollectionIds(), ","), "c1") {
		t.Fatalf("ids=%v", cs.GetCollectionIds())
	}
}
