package chat_test

import (
	"strings"
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
	if chat.System("sys").Role != xaiv1.MessageRole_ROLE_SYSTEM {
		t.Fatal("System")
	}
	if chat.Assistant("a").Role != xaiv1.MessageRole_ROLE_ASSISTANT {
		t.Fatal("Assistant")
	}
	if chat.Developer("dev").Role != xaiv1.MessageRole_ROLE_DEVELOPER {
		t.Fatal("Developer")
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

func TestNewMessageErrorPaths(t *testing.T) {
	for _, tc := range []struct {
		name string
		fn   func(...any) (*xaiv1.Message, error)
	}{
		{"NewUser", chat.NewUser},
		{"NewSystem", chat.NewSystem},
		{"NewAssistant", chat.NewAssistant},
		{"NewDeveloper", chat.NewDeveloper},
	} {
		t.Run(tc.name, func(t *testing.T) {
			m, err := tc.fn("ok", chat.Text("t"))
			if err != nil || m == nil || len(m.Content) != 2 {
				t.Fatalf("valid: %v %#v", err, m)
			}
			_, err = tc.fn(123)
			if err == nil {
				t.Fatal("expected error for int part")
			}
			if !strings.Contains(err.Error(), "unsupported content part") {
				t.Fatalf("err=%v", err)
			}
		})
	}
}

func TestPanicFactories(t *testing.T) {
	for _, name := range []string{"User", "System", "Assistant", "Developer"} {
		t.Run(name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Fatal("expected panic on illegal part")
				}
			}()
			switch name {
			case "User":
				_ = chat.User(struct{}{})
			case "System":
				_ = chat.System(true)
			case "Assistant":
				_ = chat.Assistant(3.14)
			case "Developer":
				_ = chat.Developer([]byte("x"))
			}
		})
	}
}
