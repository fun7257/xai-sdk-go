package chat

import (
	"testing"

	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

func TestChunkToolCallsIndexScoped(t *testing.T) {
	idx0, idx1 := 0, 1
	chunk := &xaiv1.GetChatCompletionChunk{
		Outputs: []*xaiv1.CompletionOutputChunk{
			{Index: 0, Delta: &xaiv1.Delta{Role: xaiv1.MessageRole_ROLE_ASSISTANT, ToolCalls: []*xaiv1.ToolCall{{Id: "t0"}}}},
			{Index: 1, Delta: &xaiv1.Delta{Role: xaiv1.MessageRole_ROLE_ASSISTANT, ToolCalls: []*xaiv1.ToolCall{{Id: "t1"}}}},
		},
	}
	c0 := NewChunk(chunk, &idx0)
	c1 := NewChunk(chunk, &idx1)
	if len(c0.ToolCalls()) != 1 || c0.ToolCalls()[0].Id != "t0" {
		t.Fatalf("c0=%v", c0.ToolCalls())
	}
	if len(c1.ToolCalls()) != 1 || c1.ToolCalls()[0].Id != "t1" {
		t.Fatalf("c1=%v", c1.ToolCalls())
	}
}

func TestResponseToolCallsIndexScoped(t *testing.T) {
	idx0, idx1 := 0, 1
	pb := &xaiv1.GetChatCompletionResponse{
		Outputs: []*xaiv1.CompletionOutput{
			{Index: 0, Message: &xaiv1.CompletionMessage{Role: xaiv1.MessageRole_ROLE_ASSISTANT, ToolCalls: []*xaiv1.ToolCall{{Id: "a"}}, Citations: []*xaiv1.InlineCitation{{}}}},
			{Index: 1, Message: &xaiv1.CompletionMessage{Role: xaiv1.MessageRole_ROLE_ASSISTANT, ToolCalls: []*xaiv1.ToolCall{{Id: "b"}}, Citations: []*xaiv1.InlineCitation{{}, {}}}},
		},
	}
	r0, r1 := NewResponse(pb, &idx0), NewResponse(pb, &idx1)
	if len(r0.ToolCalls()) != 1 || r0.ToolCalls()[0].Id != "a" {
		t.Fatalf("r0 tools=%v", r0.ToolCalls())
	}
	if len(r1.ToolCalls()) != 1 || r1.ToolCalls()[0].Id != "b" {
		t.Fatalf("r1 tools=%v", r1.ToolCalls())
	}
	if len(r0.InlineCitations()) != 1 || len(r1.InlineCitations()) != 2 {
		t.Fatalf("cites %d %d", len(r0.InlineCitations()), len(r1.InlineCitations()))
	}
}

func TestToolOutputsIndexScoped(t *testing.T) {
	idx0 := 0
	pb := &xaiv1.GetChatCompletionResponse{
		Outputs: []*xaiv1.CompletionOutput{
			{Index: 0, Message: &xaiv1.CompletionMessage{Role: xaiv1.MessageRole_ROLE_TOOL, Content: "t0"}},
			{Index: 1, Message: &xaiv1.CompletionMessage{Role: xaiv1.MessageRole_ROLE_TOOL, Content: "t1"}},
		},
	}
	r0 := NewResponse(pb, &idx0)
	if len(r0.ToolOutputs()) != 1 || r0.ToolOutputs()[0].Message.Content != "t0" {
		t.Fatalf("%v", r0.ToolOutputs())
	}
}
