package chat_test

import (
	"testing"

	"github.com/fun7257/xai-sdk-go/chat"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

func TestMessageFactories(t *testing.T) {
	u := chat.User("hello", chat.Image("https://example.com/a.png", "high"))
	if u.Role != xaiv1.MessageRole_ROLE_USER {
		t.Fatalf("role=%v", u.Role)
	}
	if len(u.Content) != 2 {
		t.Fatalf("content len=%d", len(u.Content))
	}
	if u.Content[0].GetText() != "hello" {
		t.Fatalf("text=%q", u.Content[0].GetText())
	}
	if u.Content[1].GetImageUrl() == nil || u.Content[1].GetImageUrl().GetImageUrl() == "" {
		t.Fatal("missing image")
	}
	s := chat.System("sys")
	if s.Role != xaiv1.MessageRole_ROLE_SYSTEM {
		t.Fatal(s.Role)
	}
	d := chat.Developer("dev")
	if d.Role != xaiv1.MessageRole_ROLE_DEVELOPER {
		t.Fatal(d.Role)
	}
	tr := chat.ToolResult("ok", "call-1")
	if tr.Role != xaiv1.MessageRole_ROLE_TOOL || tr.GetToolCallId() != "call-1" {
		t.Fatalf("%v %q", tr.Role, tr.GetToolCallId())
	}
	f := chat.FileByID("file_1")
	if f.GetFile() == nil || f.GetFile().FileId != "file_1" {
		t.Fatal("file id")
	}
	fd := chat.FileByData([]byte("abc"), "a.pdf", "application/pdf")
	if string(fd.GetFile().Data) != "abc" {
		t.Fatal("file data")
	}
}

// Unsupported-part error + panic paths: see message_panic_test.go.
