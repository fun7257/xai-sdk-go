package chat

import (
	"strings"
	"time"

	"github.com/fun7257/xai-sdk-go/internal/cost"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// Response wraps a chat completion response.
type Response struct {
	proto *xaiv1.GetChatCompletionResponse
	index *int
}

// NewResponse wraps a proto response.
func NewResponse(proto *xaiv1.GetChatCompletionResponse, index *int) *Response {
	return &Response{proto: proto, index: index}
}

// Proto returns the underlying proto.
func (r *Response) Proto() *xaiv1.GetChatCompletionResponse { return r.proto }

func (r *Response) assistantOutput() *xaiv1.CompletionOutput {
	if r == nil || r.proto == nil {
		return nil
	}
	var last *xaiv1.CompletionOutput
	for _, o := range r.proto.Outputs {
		if o == nil || o.Message == nil || o.Message.Role != xaiv1.MessageRole_ROLE_ASSISTANT {
			continue
		}
		if r.index != nil && o.Index != int32(*r.index) {
			continue
		}
		last = o
	}
	return last
}

// ID returns the completion response id, or empty if unset.
func (r *Response) ID() string {
	if r.proto == nil {
		return ""
	}
	return r.proto.Id
}

// Content returns the assistant text content for the selected output.
func (r *Response) Content() string {
	if o := r.assistantOutput(); o != nil && o.Message != nil {
		return o.Message.Content
	}
	return ""
}

// ReasoningContent returns the assistant reasoning text when present.
func (r *Response) ReasoningContent() string {
	if o := r.assistantOutput(); o != nil && o.Message != nil {
		return o.Message.ReasoningContent
	}
	return ""
}

// EncryptedContent returns encrypted assistant content when present.
func (r *Response) EncryptedContent() string {
	if o := r.assistantOutput(); o != nil && o.Message != nil {
		return o.Message.EncryptedContent
	}
	return ""
}

// Role returns the assistant message role string for the selected output.
func (r *Response) Role() string {
	if o := r.assistantOutput(); o != nil && o.Message != nil {
		return o.Message.Role.String()
	}
	return ""
}

// FinishReason returns the finish reason string for the selected output.
func (r *Response) FinishReason() string {
	if o := r.assistantOutput(); o != nil {
		return o.FinishReason.String()
	}
	return ""
}

// Logprobs returns token log probabilities for the selected assistant output when present.
func (r *Response) Logprobs() *xaiv1.LogProbs {
	if o := r.assistantOutput(); o != nil {
		return o.GetLogprobs()
	}
	return nil
}

// ToolCalls returns assistant tool_calls for the selected index when index is set;
// when index is nil (multi/server-tool mode), merges all assistant outputs.
func (r *Response) ToolCalls() []*xaiv1.ToolCall {
	if r == nil || r.proto == nil {
		return nil
	}
	if r.index != nil {
		if o := r.assistantOutput(); o != nil && o.Message != nil {
			return o.Message.ToolCalls
		}
		return nil
	}
	var out []*xaiv1.ToolCall
	for _, o := range r.proto.Outputs {
		if o != nil && o.Message != nil && o.Message.Role == xaiv1.MessageRole_ROLE_ASSISTANT {
			out = append(out, o.Message.ToolCalls...)
		}
	}
	return out
}

// ToolOutputs returns tool-role completion outputs.
// When index is set, only tool outputs with that Index are returned; when index
// is nil, all tool-role outputs are merged (server-side multi-tool mode).
func (r *Response) ToolOutputs() []*xaiv1.CompletionOutput {
	if r == nil || r.proto == nil {
		return nil
	}
	var out []*xaiv1.CompletionOutput
	for _, o := range r.proto.Outputs {
		if o == nil || o.Message == nil || o.Message.Role != xaiv1.MessageRole_ROLE_TOOL {
			continue
		}
		if r.index != nil && o.Index != int32(*r.index) {
			continue
		}
		out = append(out, o)
	}
	return out
}

// Citations returns top-level citation URLs from the response when present.
func (r *Response) Citations() []string {
	if r.proto == nil {
		return nil
	}
	return r.proto.Citations
}

// InlineCitations returns assistant citations for the selected index when set;
// when index is nil, merges all assistant outputs.
func (r *Response) InlineCitations() []*xaiv1.InlineCitation {
	if r == nil || r.proto == nil {
		return nil
	}
	if r.index != nil {
		if o := r.assistantOutput(); o != nil && o.Message != nil {
			return o.Message.Citations
		}
		return nil
	}
	var out []*xaiv1.InlineCitation
	for _, o := range r.proto.Outputs {
		if o != nil && o.Message != nil && o.Message.Role == xaiv1.MessageRole_ROLE_ASSISTANT {
			out = append(out, o.Message.Citations...)
		}
	}
	return out
}

// Usage returns sampling usage statistics when present.
func (r *Response) Usage() *xaiv1.SamplingUsage {
	if r.proto == nil {
		return nil
	}
	return r.proto.Usage
}

// CostUSD returns estimated cost in USD when usage reports cost ticks.
func (r *Response) CostUSD() (float64, bool) {
	u := r.Usage()
	if u == nil || u.CostInUsdTicks == nil {
		return 0, false
	}
	return cost.FromTicks(*u.CostInUsdTicks, true)
}

// SystemFingerprint returns the system fingerprint string when present.
func (r *Response) SystemFingerprint() string {
	if r.proto == nil {
		return ""
	}
	return r.proto.SystemFingerprint
}

// ServiceTier returns "priority" or "default" for the response service tier.
func (r *Response) ServiceTier() string {
	if r.proto == nil {
		return "default"
	}
	if r.proto.ServiceTier == xaiv1.ServiceTier_SERVICE_TIER_PRIORITY {
		return "priority"
	}
	return "default"
}

// ServerSideToolUsage returns counts of server-side tools used, keyed by tool name.
func (r *Response) ServerSideToolUsage() map[string]int {
	u := r.Usage()
	if u == nil {
		return nil
	}
	m := map[string]int{}
	for _, t := range u.ServerSideToolsUsed {
		m[t.String()]++
	}
	return m
}

// Settings returns request settings echoed by the server when present.
func (r *Response) Settings() *xaiv1.RequestSettings {
	if r.proto == nil {
		return nil
	}
	return r.proto.Settings
}

// DebugOutput returns debug payload from the response when present.
func (r *Response) DebugOutput() *xaiv1.DebugOutput {
	if r.proto == nil {
		return nil
	}
	return r.proto.DebugOutput
}

// Created returns the response creation time, or the zero time if unset.
func (r *Response) Created() time.Time {
	if r.proto == nil || r.proto.Created == nil {
		return time.Time{}
	}
	return r.proto.Created.AsTime()
}

// Chunk wraps a streaming chunk.
type Chunk struct {
	proto *xaiv1.GetChatCompletionChunk
	index *int
}

// NewChunk wraps a streaming completion chunk.
func NewChunk(proto *xaiv1.GetChatCompletionChunk, index *int) *Chunk {
	return &Chunk{proto: proto, index: index}
}

// Proto returns the underlying streaming chunk proto.
func (c *Chunk) Proto() *xaiv1.GetChatCompletionChunk { return c.proto }

// Content returns concatenated assistant text deltas in this chunk.
func (c *Chunk) Content() string {
	if c.proto == nil {
		return ""
	}
	var b strings.Builder
	for _, o := range c.proto.Outputs {
		if o == nil || o.Delta == nil || o.Delta.Role != xaiv1.MessageRole_ROLE_ASSISTANT {
			continue
		}
		if c.index != nil && o.Index != int32(*c.index) {
			continue
		}
		b.WriteString(o.Delta.Content)
	}
	return b.String()
}

// ReasoningContent returns concatenated reasoning deltas in this chunk.
func (c *Chunk) ReasoningContent() string {
	if c.proto == nil {
		return ""
	}
	var b strings.Builder
	for _, o := range c.proto.Outputs {
		if o == nil || o.Delta == nil || o.Delta.Role != xaiv1.MessageRole_ROLE_ASSISTANT {
			continue
		}
		if c.index != nil && o.Index != int32(*c.index) {
			continue
		}
		b.WriteString(o.Delta.ReasoningContent)
	}
	return b.String()
}

// ToolCalls returns assistant tool_call deltas for the selected index when set.
func (c *Chunk) ToolCalls() []*xaiv1.ToolCall {
	if c == nil || c.proto == nil {
		return nil
	}
	var out []*xaiv1.ToolCall
	for _, o := range c.proto.Outputs {
		if o == nil || o.Delta == nil || o.Delta.Role != xaiv1.MessageRole_ROLE_ASSISTANT {
			continue
		}
		if c.index != nil && o.Index != int32(*c.index) {
			continue
		}
		out = append(out, o.Delta.ToolCalls...)
	}
	return out
}

// Citations returns top-level citation URLs from this chunk when present.
func (c *Chunk) Citations() []string {
	if c.proto == nil {
		return nil
	}
	return c.proto.Citations
}

// EncryptedContent returns encrypted content deltas for the selected index.
func (c *Chunk) EncryptedContent() string {
	if c.proto == nil {
		return ""
	}
	var b strings.Builder
	for _, o := range c.proto.Outputs {
		if o == nil || o.Delta == nil || o.Delta.Role != xaiv1.MessageRole_ROLE_ASSISTANT {
			continue
		}
		if c.index != nil && o.Index != int32(*c.index) {
			continue
		}
		b.WriteString(o.Delta.EncryptedContent)
	}
	return b.String()
}

// InlineCitations returns citation deltas on assistant outputs.
func (c *Chunk) InlineCitations() []*xaiv1.InlineCitation {
	if c.proto == nil {
		return nil
	}
	var out []*xaiv1.InlineCitation
	for _, o := range c.proto.Outputs {
		if o == nil || o.Delta == nil {
			continue
		}
		if c.index != nil && o.Index != int32(*c.index) {
			continue
		}
		if o.Delta.Role == xaiv1.MessageRole_ROLE_ASSISTANT {
			out = append(out, o.Delta.Citations...)
		}
	}
	return out
}

// ToolOutputs returns tool-role output deltas in this chunk.
func (c *Chunk) ToolOutputs() []*xaiv1.CompletionOutputChunk {
	if c.proto == nil {
		return nil
	}
	var out []*xaiv1.CompletionOutputChunk
	for _, o := range c.proto.Outputs {
		if o != nil && o.Delta != nil && o.Delta.Role == xaiv1.MessageRole_ROLE_TOOL {
			out = append(out, o)
		}
	}
	return out
}

// ServerSideToolUsage aggregates server-side tool counts from chunk usage when present.
func (c *Chunk) ServerSideToolUsage() map[string]int {
	if c.proto == nil || c.proto.Usage == nil {
		return nil
	}
	m := map[string]int{}
	for _, t := range c.proto.Usage.ServerSideToolsUsed {
		m[t.String()]++
	}
	return m
}

// DebugOutput returns debug payload from the chunk when present.
func (c *Chunk) DebugOutput() *xaiv1.DebugOutput {
	if c.proto == nil {
		return nil
	}
	return c.proto.DebugOutput
}

// CompactContextResponse wraps compaction results.
type CompactContextResponse struct {
	proto *xaiv1.CompactContextResponse
}

// Proto returns the underlying compaction response proto.
func (c *CompactContextResponse) Proto() *xaiv1.CompactContextResponse { return c.proto }

// EncryptedContent returns the encrypted compaction blob when present.
func (c *CompactContextResponse) EncryptedContent() string {
	if c.proto == nil {
		return ""
	}
	return c.proto.EncryptedContent
}

// DroppedMessageCount returns how many messages were dropped during compaction.
func (c *CompactContextResponse) DroppedMessageCount() uint32 {
	if c.proto == nil {
		return 0
	}
	return c.proto.DroppedMessageCount
}

// ID returns the compaction response id, or empty if unset.
func (c *CompactContextResponse) ID() string {
	if c.proto == nil {
		return ""
	}
	return c.proto.Id
}

// Usage returns sampling usage from compaction when present.
func (c *CompactContextResponse) Usage() *xaiv1.SamplingUsage {
	if c.proto == nil {
		return nil
	}
	return c.proto.Usage
}

// CostUSD returns estimated cost from compaction usage when reported.
func (c *CompactContextResponse) CostUSD() (float64, bool) {
	u := c.Usage()
	if u == nil || u.CostInUsdTicks == nil {
		return 0, false
	}
	return cost.FromTicks(*u.CostInUsdTicks, true)
}

// ProcessChunk accumulates a stream chunk into the response.
func (r *Response) ProcessChunk(chunk *xaiv1.GetChatCompletionChunk) {
	if r.proto == nil {
		r.proto = &xaiv1.GetChatCompletionResponse{}
	}
	if chunk == nil {
		return
	}
	r.proto.Id = chunk.Id
	r.proto.Model = chunk.Model
	r.proto.SystemFingerprint = chunk.SystemFingerprint
	r.proto.Usage = chunk.Usage
	r.proto.ServiceTier = chunk.ServiceTier
	if chunk.Created != nil {
		r.proto.Created = chunk.Created
	}
	r.proto.Citations = append(r.proto.Citations, chunk.Citations...)
	for _, o := range chunk.Outputs {
		if o == nil {
			continue
		}
		for len(r.proto.Outputs) <= int(o.Index) {
			r.proto.Outputs = append(r.proto.Outputs, &xaiv1.CompletionOutput{Index: int32(len(r.proto.Outputs))})
		}
		choice := r.proto.Outputs[o.Index]
		choice.Index = o.Index
		choice.FinishReason = o.FinishReason
		if choice.Message == nil {
			choice.Message = &xaiv1.CompletionMessage{}
		}
		if o.Delta != nil {
			choice.Message.Role = o.Delta.Role
			choice.Message.Content += o.Delta.Content
			choice.Message.ReasoningContent += o.Delta.ReasoningContent
			choice.Message.EncryptedContent += o.Delta.EncryptedContent
			choice.Message.ToolCalls = append(choice.Message.ToolCalls, o.Delta.ToolCalls...)
			choice.Message.Citations = append(choice.Message.Citations, o.Delta.Citations...)
		}
		if o.Logprobs != nil {
			choice.Logprobs = o.Logprobs
		}
	}
}
