package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/fun7257/xai-sdk-go/internal/poll"
	"github.com/fun7257/xai-sdk-go/search"
	"github.com/fun7257/xai-sdk-go/telemetry"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

// Chat is a stateful multi-turn conversation.
//
// A Chat instance is NOT safe for concurrent use. Callers must not invoke
// Sample, Stream, Parse, Defer, Compact, or Append on the same Chat from
// multiple goroutines without external synchronization. Prefer creating a
// new Chat per concurrent request.
type Chat struct {
	stub           xaiv1.ChatClient
	req            *xaiv1.GetCompletionsRequest
	conversationID string
	batchRequestID string
}

// Option configures Create.
type Option func(*createConfig)

type createConfig struct {
	messages            []*xaiv1.Message
	user                string
	maxTokens           *int32
	seed                *int32
	stop                []string
	temperature         *float32
	topP                *float32
	logprobs            bool
	topLogprobs         *int32
	tools               []*xaiv1.Tool
	toolChoice          *xaiv1.ToolChoice
	parallelToolCalls   *bool
	responseFormat      *xaiv1.ResponseFormat
	frequencyPenalty    *float32
	presencePenalty     *float32
	reasoningEffort     *xaiv1.ReasoningEffort
	searchParameters    *xaiv1.SearchParameters
	storeMessages       bool
	previousResponseID  *string
	useEncryptedContent bool
	maxTurns            *int32
	include             []xaiv1.IncludeOption
	agentCount          *xaiv1.AgentCount
	batchRequestID      string
	serviceTier         xaiv1.ServiceTier
	conversationID      string
}

func WithMessages(msgs ...*xaiv1.Message) Option {
	return func(c *createConfig) { c.messages = append(c.messages, msgs...) }
}
func WithUser(u string) Option                   { return func(c *createConfig) { c.user = u } }
func WithMaxTokens(n int32) Option               { return func(c *createConfig) { c.maxTokens = &n } }
func WithSeed(n int32) Option                    { return func(c *createConfig) { c.seed = &n } }
func WithStop(s ...string) Option                { return func(c *createConfig) { c.stop = s } }
func WithTemperature(t float32) Option           { return func(c *createConfig) { c.temperature = &t } }
func WithTopP(v float32) Option                  { return func(c *createConfig) { c.topP = &v } }
func WithLogprobs(v bool) Option                 { return func(c *createConfig) { c.logprobs = v } }
func WithTopLogprobs(n int32) Option             { return func(c *createConfig) { c.topLogprobs = &n } }
func WithTools(t ...*xaiv1.Tool) Option          { return func(c *createConfig) { c.tools = t } }
func WithToolChoice(tc *xaiv1.ToolChoice) Option { return func(c *createConfig) { c.toolChoice = tc } }
func WithParallelToolCalls(v bool) Option        { return func(c *createConfig) { c.parallelToolCalls = &v } }
func WithResponseFormat(rf *xaiv1.ResponseFormat) Option {
	return func(c *createConfig) { c.responseFormat = rf }
}

// WithResponseFormatText requests plain-text responses.
func WithResponseFormatText() Option {
	return WithResponseFormat(&xaiv1.ResponseFormat{FormatType: xaiv1.FormatType_FORMAT_TYPE_TEXT})
}

// WithResponseFormatJSONObject requests a JSON object response (no schema).
// For JSON Schema structured outputs, use WithResponseFormatJSONSchema or Chat.Parse.
func WithResponseFormatJSONObject() Option {
	return WithResponseFormat(&xaiv1.ResponseFormat{FormatType: xaiv1.FormatType_FORMAT_TYPE_JSON_OBJECT})
}

// WithResponseFormatJSONSchema sets create-time JSON Schema response format
// without calling Parse (CHAT-05).
func WithResponseFormatJSONSchema(schemaJSON []byte) Option {
	s := string(schemaJSON)
	return WithResponseFormat(&xaiv1.ResponseFormat{
		FormatType: xaiv1.FormatType_FORMAT_TYPE_JSON_SCHEMA,
		Schema:     &s,
	})
}
func WithFrequencyPenalty(v float32) Option { return func(c *createConfig) { c.frequencyPenalty = &v } }
func WithPresencePenalty(v float32) Option  { return func(c *createConfig) { c.presencePenalty = &v } }
func WithReasoningEffort(s string) Option {
	return func(c *createConfig) {
		var e xaiv1.ReasoningEffort
		switch s {
		case "none":
			e = xaiv1.ReasoningEffort_EFFORT_NONE
		case "low":
			e = xaiv1.ReasoningEffort_EFFORT_LOW
		case "high":
			e = xaiv1.ReasoningEffort_EFFORT_HIGH
		default:
			e = xaiv1.ReasoningEffort_EFFORT_MEDIUM
		}
		c.reasoningEffort = &e
	}
}
func WithSearchParameters(p search.Parameters) Option {
	return func(c *createConfig) { c.searchParameters = p.Proto() }
}
func WithStoreMessages(v bool) Option { return func(c *createConfig) { c.storeMessages = v } }
func WithPreviousResponseID(id string) Option {
	return func(c *createConfig) { c.previousResponseID = &id }
}
func WithUseEncryptedContent(v bool) Option {
	return func(c *createConfig) { c.useEncryptedContent = v }
}
func WithMaxTurns(n int32) Option { return func(c *createConfig) { c.maxTurns = &n } }
func WithInclude(opts ...string) Option {
	return func(c *createConfig) {
		for _, o := range opts {
			c.include = append(c.include, includeToProto(o))
		}
	}
}
func WithAgentCount(n int) Option {
	return func(c *createConfig) {
		var a xaiv1.AgentCount
		switch n {
		case 16:
			a = xaiv1.AgentCount_AGENT_COUNT_16
		default:
			a = xaiv1.AgentCount_AGENT_COUNT_4
		}
		c.agentCount = &a
	}
}

// WithBatchRequestID stores a client-side batch request id for use with
// [Chat.BatchRequestID] / [Chat.PrepareBatchRequest]. The chat completion proto
// has no batch_request_id field; this value is not sent on Sample/Stream.
func WithBatchRequestID(id string) Option { return func(c *createConfig) { c.batchRequestID = id } }

func WithServiceTier(tier string) Option {
	return func(c *createConfig) {
		if tier == "priority" {
			c.serviceTier = xaiv1.ServiceTier_SERVICE_TIER_PRIORITY
		} else {
			c.serviceTier = xaiv1.ServiceTier_SERVICE_TIER_DEFAULT
		}
	}
}

// WithConversationID sets a client-side conversation correlation id.
// There is no conversation_id field on GetCompletionsRequest in the public
// proto; the value is available via [Chat.ConversationID] for application logging
// or multi-session bookkeeping and is not sent on the wire.
func WithConversationID(id string) Option { return func(c *createConfig) { c.conversationID = id } }

func includeToProto(s string) xaiv1.IncludeOption {
	switch s {
	case "web_search_call_output":
		return xaiv1.IncludeOption_INCLUDE_OPTION_WEB_SEARCH_CALL_OUTPUT
	case "x_search_call_output":
		return xaiv1.IncludeOption_INCLUDE_OPTION_X_SEARCH_CALL_OUTPUT
	case "code_execution_call_output":
		return xaiv1.IncludeOption_INCLUDE_OPTION_CODE_EXECUTION_CALL_OUTPUT
	case "collections_search_call_output":
		return xaiv1.IncludeOption_INCLUDE_OPTION_COLLECTIONS_SEARCH_CALL_OUTPUT
	case "attachment_search_call_output":
		return xaiv1.IncludeOption_INCLUDE_OPTION_ATTACHMENT_SEARCH_CALL_OUTPUT
	case "mcp_call_output":
		return xaiv1.IncludeOption_INCLUDE_OPTION_MCP_CALL_OUTPUT
	case "inline_citations":
		return xaiv1.IncludeOption_INCLUDE_OPTION_INLINE_CITATIONS
	case "verbose_streaming":
		return xaiv1.IncludeOption_INCLUDE_OPTION_VERBOSE_STREAMING
	default:
		return xaiv1.IncludeOption_INCLUDE_OPTION_INVALID
	}
}

// CallOpt configures a single Sample / Stream / Defer execution (count of completions).
type CallOpt func(*callCfg)

type callCfg struct {
	n int32
}

// WithN requests n completions (default 1). Prefer Samples/Defers when n > 1
// for multi-result handling; StreamReader accepts WithN for multi-index streams.
func WithN(n int32) CallOpt {
	return func(c *callCfg) {
		if n < 1 {
			n = 1
		}
		c.n = n
	}
}

func applyCallOpts(opts []CallOpt) callCfg {
	cfg := callCfg{n: 1}
	for _, o := range opts {
		if o != nil {
			o(&cfg)
		}
	}
	if cfg.n < 1 {
		cfg.n = 1
	}
	return cfg
}

// Create builds a chat session without issuing an RPC.
func (c *Client) Create(model string, opts ...Option) *Chat {
	cfg := &createConfig{}
	for _, o := range opts {
		o(cfg)
	}
	req := &xaiv1.GetCompletionsRequest{
		Model:               model,
		Messages:            cfg.messages,
		User:                cfg.user,
		Stop:                cfg.stop,
		Logprobs:            cfg.logprobs,
		Tools:               cfg.tools,
		ToolChoice:          cfg.toolChoice,
		ResponseFormat:      cfg.responseFormat,
		SearchParameters:    cfg.searchParameters,
		StoreMessages:       cfg.storeMessages,
		UseEncryptedContent: cfg.useEncryptedContent,
		Include:             cfg.include,
		ServiceTier:         cfg.serviceTier,
	}
	req.MaxTokens = cfg.maxTokens
	req.Seed = cfg.seed
	req.Temperature = cfg.temperature
	req.TopP = cfg.topP
	req.TopLogprobs = cfg.topLogprobs
	req.ParallelToolCalls = cfg.parallelToolCalls
	req.FrequencyPenalty = cfg.frequencyPenalty
	req.PresencePenalty = cfg.presencePenalty
	req.ReasoningEffort = cfg.reasoningEffort
	req.PreviousResponseId = cfg.previousResponseID
	req.MaxTurns = cfg.maxTurns
	req.AgentCount = cfg.agentCount
	return &Chat{
		stub:           c.stub,
		req:            req,
		conversationID: cfg.conversationID,
		batchRequestID: cfg.batchRequestID,
	}
}

// ConversationID returns the client-side conversation id set via WithConversationID.
func (ch *Chat) ConversationID() string {
	if ch == nil {
		return ""
	}
	return ch.conversationID
}

// BatchRequestID returns the client-side batch request id set via WithBatchRequestID.
func (ch *Chat) BatchRequestID() string {
	if ch == nil {
		return ""
	}
	return ch.batchRequestID
}

// PrepareBatchRequest builds a GetCompletionsRequest snapshot suitable for batch
// helpers (e.g. batch.ChatBatchRequest). Uses n=1. Does not send an RPC.
// The batch request id is not on the proto; pass ch.BatchRequestID() to the
// batch helper separately.
func (ch *Chat) PrepareBatchRequest() (*xaiv1.GetCompletionsRequest, error) {
	return ch.makeRequest(1)
}

// Append adds a message, response, or compaction to history (CHAT-12).
//
// When appending *Response:
//   - If index is nil (multi/server-tool mode), every output message is appended.
//   - Otherwise only the selected assistant output is appended, including
//     reasoning_content, encrypted_content, and tool_calls for follow-up turns.
//
// CompactContextResponse replaces history with a single user message carrying
// the encrypted compaction blob (Python parity). previous_response_id and
// store_messages remain Create-level request fields and are not rewritten by Append.
func (ch *Chat) Append(msg any) error {
	switch v := msg.(type) {
	case *xaiv1.Message:
		ch.req.Messages = append(ch.req.Messages, v)
	case *Response:
		if v == nil || v.proto == nil {
			return fmt.Errorf("nil response")
		}
		if v.index == nil {
			for _, o := range v.proto.Outputs {
				if o == nil || o.Message == nil {
					continue
				}
				ch.req.Messages = append(ch.req.Messages, &xaiv1.Message{
					Role:             o.Message.Role,
					Content:          []*xaiv1.Content{Text(o.Message.Content)},
					ReasoningContent: strPtr(o.Message.ReasoningContent),
					EncryptedContent: o.Message.EncryptedContent,
					ToolCalls:        o.Message.ToolCalls,
				})
			}
		} else {
			o := v.assistantOutput()
			if o == nil || o.Message == nil {
				return fmt.Errorf("empty assistant output")
			}
			ch.req.Messages = append(ch.req.Messages, &xaiv1.Message{
				Role:             o.Message.Role,
				Content:          []*xaiv1.Content{Text(o.Message.Content)},
				ReasoningContent: strPtr(o.Message.ReasoningContent),
				EncryptedContent: o.Message.EncryptedContent,
				ToolCalls:        o.Message.ToolCalls,
			})
		}
	case *CompactContextResponse:
		if v == nil || v.proto == nil {
			return fmt.Errorf("nil compaction")
		}
		ch.req.Messages = []*xaiv1.Message{{
			Role:             xaiv1.MessageRole_ROLE_USER,
			EncryptedContent: v.proto.EncryptedContent,
		}}
	default:
		return fmt.Errorf("unsupported append type %T", msg)
	}
	return nil
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Messages returns a copy of conversation messages.
func (ch *Chat) Messages() []*xaiv1.Message {
	out := make([]*xaiv1.Message, len(ch.req.Messages))
	copy(out, ch.req.Messages)
	return out
}

func (ch *Chat) makeRequest(n int32) (*xaiv1.GetCompletionsRequest, error) {
	if ch.req == nil || len(ch.req.Messages) == 0 {
		return nil, fmt.Errorf("cannot create a completion request: no messages provided")
	}
	// Clone so concurrent Sample/Stream/Defers on different n values cannot race
	// the session's stored request (and N is not left sticky on the session).
	req := proto.Clone(ch.req).(*xaiv1.GetCompletionsRequest)
	if n > 0 {
		nn := n
		req.N = &nn
	} else {
		req.N = nil
	}
	return req, nil
}

func (ch *Chat) spanAttrs() []attribute.KeyValue {
	var extras []attribute.KeyValue
	if ch.req != nil {
		if ch.req.Temperature != nil {
			extras = append(extras, attribute.Float64("gen_ai.request.temperature", float64(*ch.req.Temperature)))
		}
		if ch.req.TopP != nil {
			extras = append(extras, attribute.Float64("gen_ai.request.top_p", float64(*ch.req.TopP)))
		}
		if ch.req.MaxTokens != nil {
			extras = append(extras, attribute.Int("gen_ai.request.max_tokens", int(*ch.req.MaxTokens)))
		}
		if ch.req.FrequencyPenalty != nil {
			extras = append(extras, attribute.Float64("gen_ai.request.frequency_penalty", float64(*ch.req.FrequencyPenalty)))
		}
		if ch.req.PresencePenalty != nil {
			extras = append(extras, attribute.Float64("gen_ai.request.presence_penalty", float64(*ch.req.PresencePenalty)))
		}
		if len(ch.req.Tools) > 0 {
			extras = append(extras, attribute.Int("gen_ai.request.tool_count", len(ch.req.Tools)))
		}
	}
	model := ""
	msgs := 0
	if ch.req != nil {
		model = ch.req.Model
		msgs = len(ch.req.Messages)
	}
	return telemetry.ChatRequestAttrs(model, msgs, ch.conversationID, extras...)
}

// wrapMultiResponses builds one Response per completion index for multi-n results.
func wrapMultiResponses(pb *xaiv1.GetChatCompletionResponse, n int32) []*Response {
	if pb == nil {
		return nil
	}
	// Collect max assistant index present.
	maxIdx := int32(-1)
	for _, o := range pb.Outputs {
		if o == nil {
			continue
		}
		if o.Index > maxIdx {
			maxIdx = o.Index
		}
	}
	count := int(n)
	if maxIdx+1 > int32(count) {
		count = int(maxIdx) + 1
	}
	if count < 1 {
		count = 1
	}
	out := make([]*Response, 0, count)
	for i := 0; i < count; i++ {
		idx := i
		out = append(out, NewResponse(pb, &idx))
	}
	return out
}

func (ch *Chat) usesServerSideTools() bool {
	for _, t := range ch.req.Tools {
		if t != nil && t.GetFunction() == nil {
			return true
		}
	}
	return false
}

func autoDetectMulti(index *int, outputs []*xaiv1.CompletionOutput) *int {
	if index != nil && *index == 0 && len(outputs) > 0 {
		maxIdx := int32(0)
		for _, o := range outputs {
			if o != nil && o.Index > maxIdx {
				maxIdx = o.Index
			}
		}
		if maxIdx > 0 {
			return nil
		}
	}
	return index
}

func autoDetectMultiChunk(index *int, outputs []*xaiv1.CompletionOutputChunk) *int {
	if index != nil && *index == 0 && len(outputs) > 0 {
		maxIdx := int32(0)
		for _, o := range outputs {
			if o != nil && o.Index > maxIdx {
				maxIdx = o.Index
			}
		}
		if maxIdx > 0 {
			return nil
		}
	}
	return index
}

// Sample returns one completion (n=1). This is the preferred single-shot API.
func (ch *Chat) Sample(ctx context.Context) (*Response, error) {
	all, err := ch.Samples(ctx)
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("empty completion response")
	}
	return all[0], nil
}

// Samples returns one or more completions. Use WithN(k) for k > 1.
// Each element is index-scoped. Server-side tool multi-outputs may share one Response.
func (ch *Chat) Samples(ctx context.Context, opts ...CallOpt) ([]*Response, error) {
	cfg := applyCallOpts(opts)
	return ch.sampleN(ctx, cfg.n)
}

func (ch *Chat) sampleN(ctx context.Context, n int32) ([]*Response, error) {
	if n < 1 {
		n = 1
	}
	ctx, span := telemetry.StartSpan(ctx, telemetry.SpanChatSample, ch.spanAttrs()...)
	var err error
	defer func() { telemetry.EndSpan(span, err) }()

	req, err := ch.makeRequest(n)
	if err != nil {
		return nil, err
	}
	var pb *xaiv1.GetChatCompletionResponse
	pb, err = ch.stub.GetCompletion(ctx, req)
	if err != nil {
		return nil, err
	}
	if ch.usesServerSideTools() {
		return []*Response{NewResponse(pb, autoDetectMulti(nil, pb.Outputs))}, nil
	}
	return wrapMultiResponses(pb, n), nil
}

// StreamEvent is one stream step.
type StreamEvent struct {
	Response *Response
	Chunk    *Chunk
}

// StreamReader is the primary streaming DX: call Recv until io.EOF.
//
//	sr, err := chat.StreamReader(ctx)
//	if err != nil { ... }
//	defer sr.Close() // always; ends span and cancels the RPC if abandoned early
//	for {
//	    ev, err := sr.Recv()
//	    if err == io.EOF { break }
//	    ...
//	}
//
// Stream (channel form) remains available for range-style consumers; it always
// closes the underlying StreamReader. Callers that abandon the channel must
// cancel the context so the producer can exit.
type StreamReader struct {
	ctx    context.Context
	cancel context.CancelFunc
	stream xaiv1.Chat_GetCompletionChunkClient
	resp   *Response
	idx    *int
	err    error
	done   bool
	closed bool
	span   trace.Span // ended on EOF, error, or Close
	n      int32
}

// StreamReader starts a streaming completion. Pass WithN(k) for multi-completion streams.
// Prefer this over Stream for sequential Recv-style consumption. Always call Close (defer).
func (ch *Chat) StreamReader(ctx context.Context, opts ...CallOpt) (*StreamReader, error) {
	cfg := applyCallOpts(opts)
	n := cfg.n
	if n < 1 {
		n = 1
	}
	ctx, span := telemetry.StartSpan(ctx, telemetry.SpanChatStream, ch.spanAttrs()...)
	ctx, cancel := context.WithCancel(ctx)
	req, err := ch.makeRequest(n)
	if err != nil {
		cancel()
		telemetry.EndSpan(span, err)
		return nil, err
	}
	idx := (*int)(nil)
	if !ch.usesServerSideTools() {
		z := 0
		idx = &z
	}
	stream, err := ch.stub.GetCompletionChunk(ctx, req)
	if err != nil {
		cancel()
		telemetry.EndSpan(span, err)
		return nil, err
	}
	return &StreamReader{
		ctx:    ctx,
		cancel: cancel,
		stream: stream,
		resp:   NewResponse(&xaiv1.GetChatCompletionResponse{Outputs: []*xaiv1.CompletionOutput{{}}}, idx),
		idx:    idx,
		span:   span,
		n:      n,
	}, nil
}

// Close cancels the stream context and ends the OTEL span if still open.
// It is safe to call multiple times and after Recv returns io.EOF.
func (sr *StreamReader) Close() error {
	if sr == nil {
		return nil
	}
	if sr.closed {
		return nil
	}
	sr.closed = true
	sr.done = true
	if sr.cancel != nil {
		sr.cancel()
	}
	// Best-effort close of the client stream when supported.
	if closer, ok := sr.stream.(interface{ CloseSend() error }); ok && sr.stream != nil {
		_ = closer.CloseSend()
	}
	sr.endSpan(sr.err)
	return nil
}

// Recv returns the next stream event. On completion it returns io.EOF.
// The Response pointer is cumulative across chunks (same pointer).
// Prefer defer Close() so early abandon still ends the span and cancels the RPC.
func (sr *StreamReader) Recv() (StreamEvent, error) {
	if sr == nil {
		return StreamEvent{}, fmt.Errorf("nil StreamReader")
	}
	if sr.closed || sr.done {
		return StreamEvent{}, io.EOF
	}
	if sr.err != nil {
		return StreamEvent{}, sr.err
	}
	if err := sr.ctx.Err(); err != nil {
		sr.err = err
		sr.done = true
		sr.endSpan(err)
		return StreamEvent{}, err
	}
	chunk, err := sr.stream.Recv()
	if err == io.EOF {
		sr.done = true
		sr.endSpan(nil)
		return StreamEvent{}, io.EOF
	}
	if err != nil {
		sr.err = err
		sr.done = true
		sr.endSpan(err)
		return StreamEvent{}, err
	}
	sr.idx = autoDetectMultiChunk(sr.idx, chunk.Outputs)
	sr.resp.index = sr.idx
	sr.resp.ProcessChunk(chunk)
	return StreamEvent{Response: sr.resp, Chunk: NewChunk(chunk, sr.idx)}, nil
}

func (sr *StreamReader) endSpan(err error) {
	if sr.span == nil {
		return
	}
	telemetry.EndSpan(sr.span, err)
	sr.span = nil
}

// Response returns the cumulative response (may be partial mid-stream).
// For n>1, index 0 is selected unless auto-detect expanded multi mode.
func (sr *StreamReader) Response() *Response {
	if sr == nil {
		return nil
	}
	return sr.resp
}

// Responses returns one Response per completion index from the cumulative proto.
// Call after the stream completes for multi-n results.
func (sr *StreamReader) Responses() []*Response {
	if sr == nil || sr.resp == nil {
		return nil
	}
	n := sr.n
	if n < 1 {
		n = 1
	}
	return wrapMultiResponses(sr.resp.proto, n)
}

// Stream is the channel form of StreamReader. Pass WithN(k) for multi-completion.
// Prefer StreamReader for sequential Recv; cancel ctx when abandoning the channel.
func (ch *Chat) Stream(ctx context.Context, opts ...CallOpt) (<-chan StreamEvent, <-chan error) {
	out := make(chan StreamEvent)
	errc := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errc)
		sr, err := ch.StreamReader(ctx, opts...)
		if err != nil {
			errc <- err
			return
		}
		defer func() { _ = sr.Close() }()
		for {
			ev, err := sr.Recv()
			if err == io.EOF {
				return
			}
			if err != nil {
				errc <- err
				return
			}
			select {
			case <-ctx.Done():
				errc <- ctx.Err()
				return
			case out <- ev:
			}
		}
	}()
	return out, errc
}

// Parse forces JSON schema output and unmarshals into dest.
// Schema must be provided as JSON Schema bytes (Go has no Pydantic equivalent;
// callers may use encoding/json struct tags + an external schema generator).
func (ch *Chat) Parse(ctx context.Context, schemaJSON []byte, dest any) (*Response, error) {
	ctx, span := telemetry.StartSpan(ctx, "chat.parse", ch.spanAttrs()...)
	var err error
	defer func() { telemetry.EndSpan(span, err) }()

	s := string(schemaJSON)
	// Apply on cloned request via temporary mutation of session format then Sample.
	prev := ch.req.ResponseFormat
	ch.req.ResponseFormat = &xaiv1.ResponseFormat{
		FormatType: xaiv1.FormatType_FORMAT_TYPE_JSON_SCHEMA,
		Schema:     &s,
	}
	var resp *Response
	resp, err = ch.Sample(ctx)
	ch.req.ResponseFormat = prev
	if err != nil {
		return nil, err
	}
	if err = json.Unmarshal([]byte(resp.Content()), dest); err != nil {
		return resp, err
	}
	return resp, nil
}

// deferCfg holds poll settings and completion count for Defers.
type deferCfg struct {
	n    int32
	poll poll.Config
}

// DeferOption configures deferred polling and completion count.
type DeferOption func(*deferCfg)

// WithDeferTimeout sets the max wait for deferred completion.
func WithDeferTimeout(d time.Duration) DeferOption {
	return func(c *deferCfg) { c.poll.Timeout = d }
}

// WithDeferInterval sets the poll interval.
func WithDeferInterval(d time.Duration) DeferOption {
	return func(c *deferCfg) { c.poll.Interval = d }
}

// WithDeferN requests n deferred completions (default 1). Prefer Defers when n > 1.
func WithDeferN(n int32) DeferOption {
	return func(c *deferCfg) {
		if n < 1 {
			n = 1
		}
		c.n = n
	}
}

// Defer runs a deferred completion with n=1 and polls until done.
// WithDeferN(k) for k > 1 is rejected — use Defers so no completions are dropped.
func (ch *Chat) Defer(ctx context.Context, opts ...DeferOption) (*Response, error) {
	dcfg := deferCfg{n: 1, poll: poll.Default()}
	for _, o := range opts {
		if o != nil {
			o(&dcfg)
		}
	}
	if dcfg.n > 1 {
		return nil, fmt.Errorf("chat.Defer: n=%d > 1; use Defers with WithDeferN to receive all completions", dcfg.n)
	}
	all, err := ch.Defers(ctx, opts...)
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("empty deferred response")
	}
	return all[0], nil
}

// Defers runs deferred completion(s) and polls until done. Use WithDeferN(k) for k > 1.
func (ch *Chat) Defers(ctx context.Context, opts ...DeferOption) ([]*Response, error) {
	dcfg := deferCfg{n: 1, poll: poll.Default()}
	dcfg.poll.Context = "waiting for deferred chat completion"
	for _, o := range opts {
		if o != nil {
			o(&dcfg)
		}
	}
	n := dcfg.n
	if n < 1 {
		n = 1
	}
	ctx, span := telemetry.StartSpan(ctx, "chat.defer", ch.spanAttrs()...)
	var err error
	defer func() { telemetry.EndSpan(span, err) }()

	req, err := ch.makeRequest(n)
	if err != nil {
		return nil, err
	}
	start, err := ch.stub.StartDeferredCompletion(ctx, req)
	if err != nil {
		return nil, err
	}
	var final *xaiv1.GetChatCompletionResponse
	err = poll.Wait(ctx, dcfg.poll, func(ctx context.Context) (bool, error) {
		r, err := ch.stub.GetDeferredCompletion(ctx, &xaiv1.GetDeferredRequest{RequestId: start.RequestId})
		if err != nil {
			return false, err
		}
		switch r.Status {
		case xaiv1.DeferredStatus_DONE:
			final = r.GetResponse()
			return true, nil
		case xaiv1.DeferredStatus_EXPIRED:
			return false, fmt.Errorf("deferred request expired")
		case xaiv1.DeferredStatus_PENDING:
			return false, nil
		case xaiv1.DeferredStatus_FAILED:
			return false, fmt.Errorf("deferred request failed")
		default:
			return false, fmt.Errorf("unknown deferred status: %v", r.Status)
		}
	})
	if err != nil {
		return nil, err
	}
	return wrapMultiResponses(final, n), nil
}

// Compact compacts the current conversation history.
func (ch *Chat) Compact(ctx context.Context) (*CompactContextResponse, error) {
	pb, err := ch.stub.CompactContext(ctx, &xaiv1.CompactContextRequest{
		Model: ch.req.Model,
		Input: ch.req.Messages,
	})
	if err != nil {
		return nil, err
	}
	res := &CompactContextResponse{proto: pb}
	_ = ch.Append(res)
	return res, nil
}
