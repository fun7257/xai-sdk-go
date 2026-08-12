package tools_test

import (
	"testing"

	"github.com/fun7257/xai-sdk-go/tools"
)

func TestAttachmentSearch(t *testing.T) {
	tool := tools.AttachmentSearch(nil)
	if tool.GetAttachmentSearch() == nil {
		t.Fatal("nil attachment search")
	}
	if tool.GetAttachmentSearch().Limit != nil {
		t.Fatal("nil limit must stay unset")
	}

	lim := int32(7)
	tool = tools.AttachmentSearch(&lim)
	lim = 99 // caller mutation must not affect the built tool
	if got := tool.GetAttachmentSearch().GetLimit(); got != 7 {
		t.Fatalf("limit=%d want 7", got)
	}
}
