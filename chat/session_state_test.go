package chat

import (
	"context"
	"errors"
	"testing"

	"google.golang.org/grpc"

	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// fakeStub overrides GetCompletion; other ChatClient methods come from the
// embedded interface and are never called in these tests.
type fakeStub struct {
	xaiv1.ChatClient
	resp *xaiv1.GetChatCompletionResponse
	err  error
}

func (f fakeStub) GetCompletion(ctx context.Context, in *xaiv1.GetCompletionsRequest, opts ...grpc.CallOption) (*xaiv1.GetChatCompletionResponse, error) {
	return f.resp, f.err
}

func newTestChat(stub xaiv1.ChatClient, prev *xaiv1.ResponseFormat) *Chat {
	return &Chat{
		stub: stub,
		req: &xaiv1.GetCompletionsRequest{
			Model:          "m",
			Messages:       []*xaiv1.Message{User("hi")},
			ResponseFormat: prev,
		},
	}
}

// Parse must restore the session's ResponseFormat on both success and error.
func TestParseRestoresResponseFormat(t *testing.T) {
	prev := &xaiv1.ResponseFormat{FormatType: xaiv1.FormatType_FORMAT_TYPE_TEXT}

	okResp := &xaiv1.GetChatCompletionResponse{Outputs: []*xaiv1.CompletionOutput{{
		Message: &xaiv1.CompletionMessage{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: `{"a":1}`},
	}}}
	ch := newTestChat(fakeStub{resp: okResp}, prev)
	var dest struct{ A int }
	if _, err := ch.Parse(context.Background(), []byte(`{}`), &dest); err != nil {
		t.Fatal(err)
	}
	if ch.req.ResponseFormat != prev {
		t.Fatalf("format not restored after success: %v", ch.req.ResponseFormat)
	}
	if dest.A != 1 {
		t.Fatalf("dest=%+v", dest)
	}

	ch = newTestChat(fakeStub{err: errors.New("boom")}, prev)
	if _, err := ch.Parse(context.Background(), []byte(`{}`), &dest); err == nil {
		t.Fatal("expected RPC error")
	}
	if ch.req.ResponseFormat != prev {
		t.Fatalf("format not restored after error: %v", ch.req.ResponseFormat)
	}
}

// A trailing chunk without id/model/usage must not erase accumulated values.
func TestProcessChunkPreservesFieldsFromEarlierChunks(t *testing.T) {
	ticks := int64(42)
	r := NewResponse(&xaiv1.GetChatCompletionResponse{}, nil)
	r.ProcessChunk(&xaiv1.GetChatCompletionChunk{
		Id:    "resp-1",
		Model: "grok",
		Usage: &xaiv1.SamplingUsage{CostInUsdTicks: &ticks},
		Outputs: []*xaiv1.CompletionOutputChunk{
			{Index: 0, Delta: &xaiv1.Delta{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "hel"}},
		},
	})
	// Final chunk carries only the last content delta.
	r.ProcessChunk(&xaiv1.GetChatCompletionChunk{
		Outputs: []*xaiv1.CompletionOutputChunk{
			{Index: 0, Delta: &xaiv1.Delta{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "lo"}},
		},
	})
	if r.ID() != "resp-1" {
		t.Fatalf("id erased: %q", r.ID())
	}
	if r.proto.Model != "grok" {
		t.Fatalf("model erased: %q", r.proto.Model)
	}
	if u := r.Usage(); u == nil || u.GetCostInUsdTicks() != 42 {
		t.Fatalf("usage erased: %+v", u)
	}
	if r.Content() != "hello" {
		t.Fatalf("content=%q", r.Content())
	}
}
