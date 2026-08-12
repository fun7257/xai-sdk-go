package tools_test

import (
	"testing"

	"github.com/fun7257/xai-sdk-go/tools"
)

// Explicit false must reach the wire so callers can override server defaults.
func TestWebSearchExplicitFalse(t *testing.T) {
	tool, err := tools.WebSearch(
		tools.WithImageUnderstanding(false),
		tools.WithImageSearch(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	ws := tool.GetWebSearch()
	if ws.EnableImageUnderstanding == nil || *ws.EnableImageUnderstanding {
		t.Fatalf("EnableImageUnderstanding=%v want explicit false", ws.EnableImageUnderstanding)
	}
	if ws.EnableImageSearch == nil || *ws.EnableImageSearch {
		t.Fatalf("EnableImageSearch=%v want explicit false", ws.EnableImageSearch)
	}
}

// Unset options must remain absent from the wire (server default preserved).
func TestWebSearchUnsetBoolsStayNil(t *testing.T) {
	tool, err := tools.WebSearch()
	if err != nil {
		t.Fatal(err)
	}
	ws := tool.GetWebSearch()
	if ws.EnableImageUnderstanding != nil || ws.EnableImageSearch != nil {
		t.Fatalf("unset options must stay nil: %+v", ws)
	}
}

func TestXSearchExplicitFalse(t *testing.T) {
	tool, err := tools.XSearch(tools.WithXMediaUnderstanding(false, false))
	if err != nil {
		t.Fatal(err)
	}
	xs := tool.GetXSearch()
	if xs.EnableImageUnderstanding == nil || *xs.EnableImageUnderstanding {
		t.Fatalf("EnableImageUnderstanding=%v want explicit false", xs.EnableImageUnderstanding)
	}
	if xs.EnableVideoUnderstanding == nil || *xs.EnableVideoUnderstanding {
		t.Fatalf("EnableVideoUnderstanding=%v want explicit false", xs.EnableVideoUnderstanding)
	}
}
