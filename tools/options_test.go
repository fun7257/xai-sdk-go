package tools_test

import (
	"testing"
	"time"

	"github.com/fun7257/xai-sdk-go/tools"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

func TestWebSearchOptionsAndUnchecked(t *testing.T) {
	web, err := tools.WebSearch(
		tools.WithAllowedDomains("a.com"),
		tools.WithImageUnderstanding(true),
		tools.WithImageSearch(true),
		tools.WithUserLocation("US", "SF", "CA", "America/Los_Angeles"),
	)
	if err != nil {
		t.Fatal(err)
	}
	ws := web.GetWebSearch()
	if ws == nil || !ws.GetEnableImageUnderstanding() || !ws.GetEnableImageSearch() {
		t.Fatalf("%+v", ws)
	}
	if ws.GetUserLocation() == nil || ws.GetUserLocation().GetCountry() != "US" {
		t.Fatalf("location %+v", ws.GetUserLocation())
	}
	// deprecated alias
	if _, err := tools.WebSearchChecked(tools.WithExcludedDomains("x.com")); err != nil {
		t.Fatal(err)
	}
	// escape hatch skips validation
	u := tools.UncheckedWebSearch(tools.WithAllowedDomains("a.com"), tools.WithExcludedDomains("b.com"))
	if u.GetWebSearch() == nil {
		t.Fatal("unchecked web")
	}
}

func TestXSearchOptionsAndUnchecked(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	xs, err := tools.XSearch(
		tools.WithAllowedXHandles("xai"),
		tools.WithXDateRange(&from, &to),
		tools.WithXMediaUnderstanding(true, true),
	)
	if err != nil {
		t.Fatal(err)
	}
	x := xs.GetXSearch()
	if x == nil || x.GetFromDate() == nil || x.GetToDate() == nil {
		t.Fatalf("%+v", x)
	}
	if !x.GetEnableImageUnderstanding() || !x.GetEnableVideoUnderstanding() {
		t.Fatal("media flags")
	}
	if _, err := tools.XSearchChecked(); err != nil {
		t.Fatal(err)
	}
	u := tools.UncheckedXSearch(tools.WithAllowedXHandles("a"), tools.WithExcludedXHandles("b"))
	if u.GetXSearch() == nil {
		t.Fatal("unchecked x")
	}
}

func TestCollectionsSearchLegacyAndMCPOptions(t *testing.T) {
	lim := int32(5)
	tool := tools.CollectionsSearch([]string{"c1"}, &lim, "be thorough")
	cs := tool.GetCollectionsSearch()
	if cs == nil || cs.GetLimit() != 5 || cs.GetInstructions() != "be thorough" {
		t.Fatalf("%+v", cs)
	}

	mcp := tools.MCP("https://example.com/mcp",
		tools.WithMCPLabel("lab"),
		tools.WithMCPDescription("d"),
		tools.WithMCPAuth("tok"),
		tools.WithMCPAllowedTools("t1", "t2"),
		tools.WithMCPHeaders(map[string]string{"X": "1"}),
	)
	m := mcp.GetMcp()
	if m == nil || m.GetServerLabel() != "lab" || m.GetAuthorization() != "tok" {
		t.Fatalf("%+v", m)
	}
	if len(m.GetAllowedToolNames()) != 2 || m.GetExtraHeaders()["X"] != "1" {
		t.Fatalf("%+v", m)
	}
}

func TestFunctionModeCallTypeAndServerTools(t *testing.T) {
	// string / []byte / map parameters
	fn, err := tools.Function("f", "d", `{"type":"object"}`)
	if err != nil || fn.GetFunction().GetParameters() != `{"type":"object"}` {
		t.Fatalf("%v %#v", err, fn)
	}
	raw, err := tools.Function("f2", "d", []byte(`{"type":"object"}`))
	if err != nil || raw.GetFunction() == nil {
		t.Fatal(err)
	}
	mapped, err := tools.Function("get_weather", "weather", map[string]any{
		"type": "object", "properties": map[string]any{"city": map[string]any{"type": "string"}},
	})
	if err != nil || mapped.GetFunction() == nil || mapped.GetFunction().Name != "get_weather" {
		t.Fatalf("map params: %v %#v", err, mapped)
	}
	if tools.RequiredTool("get_weather").GetFunctionName() != "get_weather" {
		t.Fatal("RequiredTool")
	}
	if tools.CodeExecution().GetCodeExecution() == nil {
		t.Fatal("CodeExecution")
	}

	if tools.Mode("none").GetMode() != xaiv1.ToolMode_TOOL_MODE_NONE {
		t.Fatal("none")
	}
	if tools.Mode("required").GetMode() != xaiv1.ToolMode_TOOL_MODE_REQUIRED {
		t.Fatal("required")
	}
	if tools.Mode("auto").GetMode() != xaiv1.ToolMode_TOOL_MODE_AUTO {
		t.Fatal("auto")
	}
	if tools.Mode("other").GetMode() != xaiv1.ToolMode_TOOL_MODE_AUTO {
		t.Fatal("default mode should be auto")
	}
	if tools.CallType(nil) != "" {
		t.Fatal("nil call type")
	}
	tc := &xaiv1.ToolCall{Type: xaiv1.ToolCallType_TOOL_CALL_TYPE_WEB_SEARCH_TOOL}
	if got := tools.CallType(tc); got == "" {
		t.Fatalf("call type empty for %v", tc.GetType())
	}
}
