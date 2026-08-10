package chat_test

import (
	"strings"
	"testing"

	"github.com/fun7257/xai-sdk-go/chat"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

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
	// valid literal path
	if chat.User("hi").GetContent()[0].GetText() != "hi" {
		t.Fatal("User")
	}
	if chat.System("s").Role != xaiv1.MessageRole_ROLE_SYSTEM {
		t.Fatal("System")
	}
	if chat.Assistant("a").Role != xaiv1.MessageRole_ROLE_ASSISTANT {
		t.Fatal("Assistant")
	}
	if chat.Developer("d").Role != xaiv1.MessageRole_ROLE_DEVELOPER {
		t.Fatal("Developer")
	}

	for _, name := range []string{"User", "System", "Assistant", "Developer"} {
		t.Run("panic_"+name, func(t *testing.T) {
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
