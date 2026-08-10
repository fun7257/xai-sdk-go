package tools_test

import (
	"strings"
	"testing"

	"github.com/fun7257/xai-sdk-go/tools"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

func TestFunctionAndServerTools(t *testing.T) {
	// WebSearch/XSearch options + Unchecked: options_test.go
	// Mutual exclusion tables: TestWebSearchMutualExclusion / TestXSearchMutualExclusion
	fn, err := tools.Function("get_weather", "weather", map[string]any{
		"type":       "object",
		"properties": map[string]any{"city": map[string]any{"type": "string"}},
	})
	if err != nil || fn.GetFunction() == nil || fn.GetFunction().Name != "get_weather" {
		t.Fatalf("%v %#v", err, fn)
	}
	if tools.CodeExecution().GetCodeExecution() == nil {
		t.Fatal("code")
	}
	if tools.CollectionsSearch([]string{"c1"}, nil, "").GetCollectionsSearch() == nil {
		t.Fatal("collections")
	}
	if tools.MCP("https://example.com/mcp", tools.WithMCPLabel("lab")).GetMcp() == nil {
		t.Fatal("mcp")
	}
	tc := tools.RequiredTool("get_weather")
	if tc.GetFunctionName() != "get_weather" {
		t.Fatal(tc.GetFunctionName())
	}
}
func TestWebSearchMutualExclusion(t *testing.T) {
	tests := []struct {
		name    string
		opts    []tools.WebSearchOption
		wantErr bool
	}{
		{"allowed only", []tools.WebSearchOption{tools.WithAllowedDomains("a.com")}, false},
		{"excluded only", []tools.WebSearchOption{tools.WithExcludedDomains("b.com")}, false},
		{"both", []tools.WebSearchOption{tools.WithAllowedDomains("a.com"), tools.WithExcludedDomains("b.com")}, true},
		{"neither", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tool, err := tools.WebSearch(tt.opts...)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if tool.GetWebSearch() == nil {
				t.Fatal("nil tool")
			}
		})
	}
}

func TestXSearchMutualExclusion(t *testing.T) {
	tests := []struct {
		name    string
		opts    []tools.XSearchOption
		wantErr bool
	}{
		{"allowed", []tools.XSearchOption{tools.WithAllowedXHandles("xai")}, false},
		{"excluded", []tools.XSearchOption{tools.WithExcludedXHandles("spam")}, false},
		{"both", []tools.XSearchOption{tools.WithAllowedXHandles("a"), tools.WithExcludedXHandles("b")}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tools.XSearch(tt.opts...)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestCollectionsSearchRetrievalMode(t *testing.T) {
	tests := []struct {
		mode    string
		wantErr bool
		check   func(*xaiv1.CollectionsSearch) bool
	}{
		{"", false, func(cs *xaiv1.CollectionsSearch) bool { return cs.RetrievalMode == nil }},
		{"hybrid", false, func(cs *xaiv1.CollectionsSearch) bool { return cs.GetHybridRetrieval() != nil }},
		{"semantic", false, func(cs *xaiv1.CollectionsSearch) bool { return cs.GetSemanticRetrieval() != nil }},
		{"keyword", false, func(cs *xaiv1.CollectionsSearch) bool { return cs.GetKeywordRetrieval() != nil }},
		{"bogus", true, nil},
	}
	for _, tt := range tests {
		t.Run(tt.mode, func(t *testing.T) {
			tool, err := tools.CollectionsSearchOpts([]string{"c1"}, tools.WithCollectionsRetrievalMode(tt.mode))
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if !strings.Contains(err.Error(), "retrieval_mode") {
					t.Fatalf("err=%v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			cs := tool.GetCollectionsSearch()
			if cs == nil || !tt.check(cs) {
				t.Fatalf("%+v", cs)
			}
		})
	}
}
