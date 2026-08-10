package chat_test

import (
	"context"
	"encoding/json"
	"io"
	"strconv"
	"testing"
	"time"

	"github.com/fun7257/xai-sdk-go/chat"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	"github.com/fun7257/xai-sdk-go/tools"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type mockChat struct {
	xaiv1.UnimplementedChatServer
	lastAuth         string
	deferred         map[string]*xaiv1.GetChatCompletionResponse
	forceDeferStatus bool
	deferStatus      xaiv1.DeferredStatus
}

func (m *mockChat) GetCompletion(ctx context.Context, req *xaiv1.GetCompletionsRequest) (*xaiv1.GetChatCompletionResponse, error) {
	m.lastAuth = testutil.AuthBearerFromContext(ctx)
	content := "hello"
	if req.ResponseFormat != nil && req.ResponseFormat.FormatType == xaiv1.FormatType_FORMAT_TYPE_JSON_SCHEMA {
		content = `{"answer":"ok"}`
	}
	ticks := int64(10_000_000_000)
	return &xaiv1.GetChatCompletionResponse{
		Id:    "resp-1",
		Model: req.Model,
		Outputs: []*xaiv1.CompletionOutput{{
			Index:        0,
			FinishReason: xaiv1.FinishReason_REASON_STOP,
			Message: &xaiv1.CompletionMessage{
				Role:             xaiv1.MessageRole_ROLE_ASSISTANT,
				Content:          content,
				ReasoningContent: "think",
				EncryptedContent: "enc",
			},
		}},
		Usage:   &xaiv1.SamplingUsage{PromptTokens: 1, CompletionTokens: 2, TotalTokens: 3, CostInUsdTicks: &ticks},
		Created: timestamppb.Now(),
	}, nil
}

func (m *mockChat) GetCompletionChunk(req *xaiv1.GetCompletionsRequest, stream xaiv1.Chat_GetCompletionChunkServer) error {
	chunks := []string{"hel", "lo"}
	for i, c := range chunks {
		fr := xaiv1.FinishReason_REASON_INVALID
		if i == len(chunks)-1 {
			fr = xaiv1.FinishReason_REASON_STOP
		}
		if err := stream.Send(&xaiv1.GetChatCompletionChunk{
			Id:    "resp-stream",
			Model: req.Model,
			Outputs: []*xaiv1.CompletionOutputChunk{{
				Index:        0,
				FinishReason: fr,
				Delta: &xaiv1.Delta{
					Role:    xaiv1.MessageRole_ROLE_ASSISTANT,
					Content: c,
				},
			}},
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m *mockChat) StartDeferredCompletion(ctx context.Context, req *xaiv1.GetCompletionsRequest) (*xaiv1.StartDeferredResponse, error) {
	if m.deferred == nil {
		m.deferred = map[string]*xaiv1.GetChatCompletionResponse{}
	}
	id := "def-1"
	m.deferred[id] = &xaiv1.GetChatCompletionResponse{
		Id: id,
		Outputs: []*xaiv1.CompletionOutput{{
			Index: 0, FinishReason: xaiv1.FinishReason_REASON_STOP,
			Message: &xaiv1.CompletionMessage{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "deferred"},
		}},
	}
	return &xaiv1.StartDeferredResponse{RequestId: id}, nil
}

func (m *mockChat) GetDeferredCompletion(ctx context.Context, req *xaiv1.GetDeferredRequest) (*xaiv1.GetDeferredCompletionResponse, error) {
	if m.forceDeferStatus {
		return &xaiv1.GetDeferredCompletionResponse{Status: m.deferStatus}, nil
	}
	r := m.deferred[req.RequestId]
	return &xaiv1.GetDeferredCompletionResponse{Status: xaiv1.DeferredStatus_DONE, Response: r}, nil
}

func (m *mockChat) GetStoredCompletion(ctx context.Context, req *xaiv1.GetStoredCompletionRequest) (*xaiv1.GetChatCompletionResponse, error) {
	return &xaiv1.GetChatCompletionResponse{
		Id: req.ResponseId,
		Outputs: []*xaiv1.CompletionOutput{{
			Index: 0, Message: &xaiv1.CompletionMessage{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "stored"},
		}},
	}, nil
}

func (m *mockChat) DeleteStoredCompletion(ctx context.Context, req *xaiv1.DeleteStoredCompletionRequest) (*xaiv1.DeleteStoredCompletionResponse, error) {
	return &xaiv1.DeleteStoredCompletionResponse{ResponseId: req.ResponseId}, nil
}

func (m *mockChat) CompactContext(ctx context.Context, req *xaiv1.CompactContextRequest) (*xaiv1.CompactContextResponse, error) {
	return &xaiv1.CompactContextResponse{Id: "c1", EncryptedContent: "blob", DroppedMessageCount: 2}, nil
}

func startChat(t *testing.T) (*chat.Client, *mockChat, func()) {
	t.Helper()
	mock := &mockChat{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterChatServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	return chat.New(srv.Conn), mock, srv.Close
}

func TestSampleStreamParseDeferStoredCompact(t *testing.T) {
	cli, mock, stop := startChat(t)
	defer stop()

	ch := cli.Create("grok-3", chat.WithMessages(chat.System("s")))
	if err := ch.Append(chat.User("hi")); err != nil {
		t.Fatal(err)
	}
	resp, err := ch.Sample(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if resp.Content() != "hello" {
		t.Fatalf("content=%q", resp.Content())
	}
	if resp.ReasoningContent() != "think" || resp.EncryptedContent() != "enc" {
		t.Fatal("reasoning/enc")
	}
	usd, ok := resp.CostUSD()
	if !ok || usd != 1.0 {
		t.Fatalf("cost %v %v", usd, ok)
	}
	if err := ch.Append(resp); err != nil {
		t.Fatal(err)
	}
	if len(ch.Messages()) != 3 {
		t.Fatalf("messages=%d", len(ch.Messages()))
	}

	// empty messages error
	empty := cli.Create("grok-3")
	if _, err := empty.Sample(context.Background()); err == nil {
		t.Fatal("expected empty error")
	}

	// stream
	ch2 := cli.Create("grok-3", chat.WithMessages(chat.User("stream")))
	events, errc := ch2.Stream(context.Background())
	var last *chat.Response
	for ev := range events {
		last = ev.Response
	}
	if err := <-errc; err != nil && err != io.EOF {
		// errc may be closed with nil
		if err != nil {
			t.Fatal(err)
		}
	}
	if last == nil || last.Content() != "hello" {
		t.Fatalf("stream content=%v", last)
	}

	// parse
	ch3 := cli.Create("grok-3", chat.WithMessages(chat.User("json")))
	schema, _ := json.Marshal(map[string]any{"type": "object", "properties": map[string]any{"answer": map[string]any{"type": "string"}}})
	var dest struct {
		Answer string `json:"answer"`
	}
	r, err := ch3.Parse(context.Background(), schema, &dest)
	if err != nil || dest.Answer != "ok" || r.Content() == "" {
		t.Fatalf("%v %#v", err, dest)
	}

	// defer
	ch4 := cli.Create("grok-3", chat.WithMessages(chat.User("d")))
	dr, err := ch4.Defer(context.Background(), chat.WithDeferInterval(time.Millisecond), chat.WithDeferTimeout(time.Second))
	if err != nil || dr.Content() != "deferred" {
		t.Fatalf("%v %v", err, dr)
	}

	// stored
	stored, err := cli.GetStoredCompletion(context.Background(), "sid")
	if err != nil || len(stored) != 1 || stored[0].Content() != "stored" {
		t.Fatalf("%v %#v", err, stored)
	}
	id, err := cli.DeleteStoredCompletion(context.Background(), "sid")
	if err != nil || id != "sid" {
		t.Fatal(err, id)
	}

	// compact
	ch5 := cli.Create("grok-3", chat.WithMessages(chat.User("a"), chat.Assistant("b")))
	cr, err := ch5.Compact(context.Background())
	if err != nil || cr.EncryptedContent() != "blob" || cr.DroppedMessageCount() != 2 {
		t.Fatalf("%v %#v", err, cr)
	}
	if len(ch5.Messages()) != 1 {
		t.Fatalf("after compact msgs=%d", len(ch5.Messages()))
	}

	// tools create option path
	fn, _ := tools.Function("f", "d", map[string]any{"type": "object"})
	_ = cli.Create("grok-3", chat.WithTools(fn), chat.WithToolChoice(tools.Mode("auto")), chat.WithReasoningEffort("high"))
	_ = mock
}

func TestCreateDoesNotCallRPC(t *testing.T) {
	cli, _, stop := startChat(t)
	defer stop()
	// Create must not panic / RPC; just builds session
	ch := cli.Create("m", chat.WithTemperature(0.2), chat.WithMaxTokens(10), chat.WithStoreMessages(true))
	if ch == nil {
		t.Fatal("nil chat")
	}
}

func TestDeferFailedStatus(t *testing.T) {
	cli, mock, stop := startChat(t)
	defer stop()
	mock.forceDeferStatus = true
	mock.deferStatus = xaiv1.DeferredStatus_FAILED
	ch := cli.Create("grok-3", chat.WithMessages(chat.User("x")))
	start := time.Now()
	_, err := ch.Defer(context.Background(), chat.WithDeferInterval(time.Millisecond), chat.WithDeferTimeout(200*time.Millisecond))
	if err == nil {
		t.Fatal("expected error on FAILED deferred status")
	}
	if time.Since(start) > 150*time.Millisecond {
		// Should fail immediately, not wait for full timeout
		// Allow some slack but fail if near 200ms timeout
		if time.Since(start) >= 180*time.Millisecond {
			t.Fatalf("Defer hung until timeout: %v err=%v", time.Since(start), err)
		}
	}
	if err.Error() != "deferred request failed" {
		t.Fatalf("err=%v", err)
	}
}

func TestDeferUnknownStatus(t *testing.T) {
	cli, mock, stop := startChat(t)
	defer stop()
	mock.forceDeferStatus = true
	mock.deferStatus = xaiv1.DeferredStatus_INVALID_DEFERRED_STATUS
	ch := cli.Create("grok-3", chat.WithMessages(chat.User("x")))
	_, err := ch.Defer(context.Background(), chat.WithDeferInterval(time.Millisecond), chat.WithDeferTimeout(200*time.Millisecond))
	if err == nil {
		t.Fatal("expected unknown status error")
	}
	if err.Error() == "deferred request failed" {
		t.Fatalf("wrong error for invalid status: %v", err)
	}
}

func TestSamplesMultiAndLogprobs(t *testing.T) {
	mock := &multiChat{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterChatServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := chat.New(srv.Conn)
	ch := cli.Create("m", chat.WithMessages(chat.User("hi")), chat.WithLogprobs(true))
	all, err := ch.Samples(context.Background(), chat.WithN(2))
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("len=%d", len(all))
	}
	if mock.lastN == nil || *mock.lastN != 2 {
		t.Fatalf("request n=%v", mock.lastN)
	}
	if all[0].Content() == all[1].Content() {
		// different content per index
		if all[0].Content() == "" {
			t.Fatal("empty")
		}
	}
	if all[0].Content() != "a0" || all[1].Content() != "a1" {
		t.Fatalf("contents %q %q", all[0].Content(), all[1].Content())
	}
	if all[0].Logprobs() == nil || len(all[0].Logprobs().Content) == 0 {
		t.Fatal("expected logprobs on index 0")
	}
	// JSON schema create option
	schema := []byte(`{"type":"object"}`)
	ch2 := cli.Create("m", chat.WithMessages(chat.User("x")), chat.WithResponseFormatJSONSchema(schema))
	if _, err := ch2.Sample(context.Background()); err != nil {
		t.Fatal(err)
	}
	if mock.lastRF == nil || mock.lastRF.FormatType != xaiv1.FormatType_FORMAT_TYPE_JSON_SCHEMA {
		t.Fatalf("rf=%v", mock.lastRF)
	}
}

func TestDefers(t *testing.T) {
	mock := &multiChat{deferDone: true}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterChatServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := chat.New(srv.Conn)
	ch := cli.Create("m", chat.WithMessages(chat.User("hi")))
	all, err := ch.Defers(context.Background(), chat.WithDeferN(2), chat.WithDeferInterval(time.Millisecond), chat.WithDeferTimeout(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("len=%d", len(all))
	}
}

func TestFileExactlyOneMode(t *testing.T) {
	if _, err := chat.File("id", []byte("x"), "", "", ""); err == nil {
		t.Fatal("expected error for multiple modes")
	}
	c, err := chat.File("id", nil, "", "", "")
	if err != nil || c.GetFile().GetFileId() != "id" {
		t.Fatalf("%v %v", err, c)
	}
}

type multiChat struct {
	xaiv1.UnimplementedChatServer
	lastN     *int32
	lastRF    *xaiv1.ResponseFormat
	deferDone bool
}

func (m *multiChat) GetCompletion(ctx context.Context, req *xaiv1.GetCompletionsRequest) (*xaiv1.GetChatCompletionResponse, error) {
	m.lastN = req.N
	m.lastRF = req.ResponseFormat
	n := int32(1)
	if req.N != nil {
		n = *req.N
	}
	outs := make([]*xaiv1.CompletionOutput, n)
	for i := int32(0); i < n; i++ {
		outs[i] = &xaiv1.CompletionOutput{
			Index: i,
			Message: &xaiv1.CompletionMessage{
				Role:    xaiv1.MessageRole_ROLE_ASSISTANT,
				Content: "a" + itoaChat(int(i)),
			},
			Logprobs: &xaiv1.LogProbs{Content: []*xaiv1.LogProb{{Token: "t", Logprob: -0.1}}},
		}
	}
	return &xaiv1.GetChatCompletionResponse{Id: "r", Model: req.Model, Outputs: outs}, nil
}

func (m *multiChat) StartDeferredCompletion(ctx context.Context, req *xaiv1.GetCompletionsRequest) (*xaiv1.StartDeferredResponse, error) {
	m.lastN = req.N
	return &xaiv1.StartDeferredResponse{RequestId: "d1"}, nil
}

func (m *multiChat) GetDeferredCompletion(ctx context.Context, req *xaiv1.GetDeferredRequest) (*xaiv1.GetDeferredCompletionResponse, error) {
	outs := []*xaiv1.CompletionOutput{
		{Index: 0, Message: &xaiv1.CompletionMessage{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "a0"}},
		{Index: 1, Message: &xaiv1.CompletionMessage{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "a1"}},
	}
	return &xaiv1.GetDeferredCompletionResponse{
		Status:   xaiv1.DeferredStatus_DONE,
		Response: &xaiv1.GetChatCompletionResponse{Outputs: outs},
	}, nil
}

func itoaChat(i int) string {
	return strconv.Itoa(i)
}

func TestSamplesToolCallsIndexScoped(t *testing.T) {
	mock := &multiChatWithTools{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterChatServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := chat.New(srv.Conn)
	ch := cli.Create("m", chat.WithMessages(chat.User("hi")))
	all, err := ch.Samples(context.Background(), chat.WithN(2))
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("len=%d", len(all))
	}
	tc0 := all[0].ToolCalls()
	tc1 := all[1].ToolCalls()
	if len(tc0) != 1 || tc0[0].GetFunction().GetName() != "f0" {
		t.Fatalf("tc0=%v", tc0)
	}
	if len(tc1) != 1 || tc1[0].GetFunction().GetName() != "f1" {
		t.Fatalf("tc1=%v", tc1)
	}
	// Inline citations index scoped
	if len(all[0].InlineCitations()) != 1 || len(all[1].InlineCitations()) != 1 {
		t.Fatalf("cites %d %d", len(all[0].InlineCitations()), len(all[1].InlineCitations()))
	}
}

func TestStreamReaderWithNMulti(t *testing.T) {
	mock := &multiStreamChat{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterChatServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := chat.New(srv.Conn)
	ch := cli.Create("m", chat.WithMessages(chat.User("hi")))
	sr, err := ch.StreamReader(context.Background(), chat.WithN(2))
	if err != nil {
		t.Fatal(err)
	}
	defer sr.Close()
	for {
		_, err := sr.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
	}
	if mock.lastN == nil || *mock.lastN != 2 {
		t.Fatalf("request n=%v", mock.lastN)
	}
	all := sr.Responses()
	if len(all) != 2 {
		t.Fatalf("Responses len=%d", len(all))
	}
	if all[0].Content() != "A0" || all[1].Content() != "B1" {
		t.Fatalf("contents %q %q", all[0].Content(), all[1].Content())
	}
	// chunk tool calls index scoped via last chunk path: verify Response tool calls if accumulated
}

type multiChatWithTools struct {
	xaiv1.UnimplementedChatServer
}

func (m *multiChatWithTools) GetCompletion(ctx context.Context, req *xaiv1.GetCompletionsRequest) (*xaiv1.GetChatCompletionResponse, error) {
	return &xaiv1.GetChatCompletionResponse{
		Id: "r", Model: req.Model,
		Outputs: []*xaiv1.CompletionOutput{
			{
				Index: 0,
				Message: &xaiv1.CompletionMessage{
					Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "a0",
					ToolCalls: []*xaiv1.ToolCall{{Id: "1", Tool: &xaiv1.ToolCall_Function{Function: &xaiv1.FunctionCall{Name: "f0"}}}},
					Citations: []*xaiv1.InlineCitation{{Id: "c0"}},
				},
			},
			{
				Index: 1,
				Message: &xaiv1.CompletionMessage{
					Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "a1",
					ToolCalls: []*xaiv1.ToolCall{{Id: "2", Tool: &xaiv1.ToolCall_Function{Function: &xaiv1.FunctionCall{Name: "f1"}}}},
					Citations: []*xaiv1.InlineCitation{{Id: "c1"}},
				},
			},
		},
	}, nil
}

type multiStreamChat struct {
	xaiv1.UnimplementedChatServer
	lastN *int32
}

func (m *multiStreamChat) GetCompletionChunk(req *xaiv1.GetCompletionsRequest, stream xaiv1.Chat_GetCompletionChunkServer) error {
	m.lastN = req.N
	// one chunk with two indices
	return stream.Send(&xaiv1.GetChatCompletionChunk{
		Id: "s", Model: req.Model,
		Outputs: []*xaiv1.CompletionOutputChunk{
			{Index: 0, FinishReason: xaiv1.FinishReason_REASON_STOP, Delta: &xaiv1.Delta{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "A0",
				ToolCalls: []*xaiv1.ToolCall{{Id: "t0", Tool: &xaiv1.ToolCall_Function{Function: &xaiv1.FunctionCall{Name: "sf0"}}}}}},
			{Index: 1, FinishReason: xaiv1.FinishReason_REASON_STOP, Delta: &xaiv1.Delta{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "B1",
				ToolCalls: []*xaiv1.ToolCall{{Id: "t1", Tool: &xaiv1.ToolCall_Function{Function: &xaiv1.FunctionCall{Name: "sf1"}}}}}},
		},
	})
}

func TestDeferRejectsNGreaterThanOne(t *testing.T) {
	mock := &multiChat{deferDone: true}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterChatServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := chat.New(srv.Conn)
	ch := cli.Create("m", chat.WithMessages(chat.User("hi")))
	_, err = ch.Defer(context.Background(), chat.WithDeferN(2))
	if err == nil {
		t.Fatal("expected error for Defer WithDeferN(2)")
	}
	// Defers still works
	all, err := ch.Defers(context.Background(), chat.WithDeferN(2), chat.WithDeferInterval(time.Millisecond), chat.WithDeferTimeout(time.Second))
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("len=%d", len(all))
	}
}
