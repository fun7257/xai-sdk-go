package chat

import (
	"testing"

	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

func TestMakeRequestClonesN(t *testing.T) {
	ch := &Chat{req: &xaiv1.GetCompletionsRequest{
		Model: "m",
		Messages: []*xaiv1.Message{{Role: xaiv1.MessageRole_ROLE_USER, Content: []*xaiv1.Content{Text("hi")}}},
	}}
	r1, err := ch.makeRequest(2)
	if err != nil {
		t.Fatal(err)
	}
	if r1.N == nil || *r1.N != 2 {
		t.Fatalf("r1.N=%v", r1.N)
	}
	if ch.req.N != nil {
		t.Fatalf("session N should not stick, got %v", *ch.req.N)
	}
	r2, err := ch.makeRequest(1)
	if err != nil {
		t.Fatal(err)
	}
	if r2.N == nil || *r2.N != 1 {
		t.Fatalf("r2.N=%v", r2.N)
	}
}
