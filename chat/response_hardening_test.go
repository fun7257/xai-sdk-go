package chat

import (
	"testing"

	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// A malformed / hostile chunk with a negative output index must be dropped,
// not panic the consumer (indexes come straight off the wire).
func TestProcessChunkDropsNegativeIndex(t *testing.T) {
	r := NewResponse(&xaiv1.GetChatCompletionResponse{}, nil)
	r.ProcessChunk(&xaiv1.GetChatCompletionChunk{
		Outputs: []*xaiv1.CompletionOutputChunk{
			{Index: -1, Delta: &xaiv1.Delta{Content: "bad"}},
			{Index: 0, Delta: &xaiv1.Delta{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "ok"}},
		},
	})
	if got := len(r.Proto().Outputs); got != 1 {
		t.Fatalf("outputs=%d want 1 (negative index dropped)", got)
	}
	if r.Proto().Outputs[0].Message.Content != "ok" {
		t.Fatalf("valid output lost: %+v", r.Proto().Outputs[0])
	}
}

// A huge output index must not force an unbounded slice allocation.
func TestProcessChunkCapsHugeIndex(t *testing.T) {
	r := NewResponse(&xaiv1.GetChatCompletionResponse{}, nil)
	r.ProcessChunk(&xaiv1.GetChatCompletionChunk{
		Outputs: []*xaiv1.CompletionOutputChunk{
			{Index: 1 << 30, Delta: &xaiv1.Delta{Content: "bad"}},
			{Index: 1, Delta: &xaiv1.Delta{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "ok"}},
		},
	})
	if got := len(r.Proto().Outputs); got != 2 {
		t.Fatalf("outputs=%d want 2 (index 0 backfill + index 1)", got)
	}
}

// wrapMultiResponses must ignore out-of-range indexes when sizing the result.
func TestWrapMultiResponsesCapsHugeIndex(t *testing.T) {
	pb := &xaiv1.GetChatCompletionResponse{
		Outputs: []*xaiv1.CompletionOutput{
			{Index: 1 << 30, Message: &xaiv1.CompletionMessage{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "x"}},
			{Index: -5},
			{Index: 1, Message: &xaiv1.CompletionMessage{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "y"}},
		},
	}
	rs := wrapMultiResponses(pb, 1)
	if len(rs) != 2 {
		t.Fatalf("responses=%d want 2 (max in-range index 1)", len(rs))
	}
	if rs[1].Content() != "y" {
		t.Fatalf("content=%q want y", rs[1].Content())
	}
}
