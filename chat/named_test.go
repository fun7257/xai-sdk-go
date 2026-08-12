package chat_test

import (
	"testing"

	"github.com/fun7257/xai-sdk-go/chat"
)

func TestNamedSetsParticipantName(t *testing.T) {
	m := chat.Named(chat.User("hi"), "alice")
	if m.GetName() != "alice" {
		t.Fatalf("name=%q", m.GetName())
	}
	if m.GetRole().String() != "ROLE_USER" || len(m.GetContent()) != 1 {
		t.Fatalf("message mangled: %+v", m)
	}
	if got := chat.Named(nil, "x"); got != nil {
		t.Fatal("nil message must pass through")
	}
}
