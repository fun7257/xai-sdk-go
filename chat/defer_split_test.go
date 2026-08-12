package chat_test

import (
	"context"
	"sync"
	"testing"

	"google.golang.org/grpc"

	"github.com/fun7257/xai-sdk-go/chat"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// deferSplitMock returns PENDING for the first poll and DONE afterwards.
type deferSplitMock struct {
	xaiv1.UnimplementedChatServer
	mu    sync.Mutex
	polls int
	reqID string
}

func (m *deferSplitMock) StartDeferredCompletion(ctx context.Context, req *xaiv1.GetCompletionsRequest) (*xaiv1.StartDeferredResponse, error) {
	return &xaiv1.StartDeferredResponse{RequestId: "req-42"}, nil
}

func (m *deferSplitMock) GetDeferredCompletion(ctx context.Context, req *xaiv1.GetDeferredRequest) (*xaiv1.GetDeferredCompletionResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reqID = req.GetRequestId()
	m.polls++
	if m.polls == 1 {
		return &xaiv1.GetDeferredCompletionResponse{Status: xaiv1.DeferredStatus_PENDING}, nil
	}
	return &xaiv1.GetDeferredCompletionResponse{
		Status: xaiv1.DeferredStatus_DONE,
		Response: &xaiv1.GetChatCompletionResponse{
			Outputs: []*xaiv1.CompletionOutput{
				{Index: 0, Message: &xaiv1.CompletionMessage{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "a"}},
				{Index: 1, Message: &xaiv1.CompletionMessage{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "b"}},
			},
		},
	}, nil
}

// DeferStart hands out the request id; DeferGet polls once per call so the
// caller owns the schedule (video Start/Get parity).
func TestDeferStartAndGet(t *testing.T) {
	mock := &deferSplitMock{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterChatServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	c := chat.New(srv.Conn).Create("m", chat.WithMessages(chat.User("hi")))
	ctx := context.Background()

	id, err := c.DeferStart(ctx, chat.WithN(2))
	if err != nil || id != "req-42" {
		t.Fatalf("DeferStart: %v id=%q", err, id)
	}

	st, resps, err := c.DeferGet(ctx, id)
	if err != nil || st != xaiv1.DeferredStatus_PENDING || resps != nil {
		t.Fatalf("first poll: %v %v %v", st, resps, err)
	}
	st, resps, err = c.DeferGet(ctx, id)
	if err != nil || st != xaiv1.DeferredStatus_DONE {
		t.Fatalf("second poll: %v %v", st, err)
	}
	if len(resps) != 2 || resps[0].Content() != "a" || resps[1].Content() != "b" {
		t.Fatalf("responses=%v", resps)
	}
	mock.mu.Lock()
	defer mock.mu.Unlock()
	if mock.reqID != "req-42" {
		t.Fatalf("wire request id=%q", mock.reqID)
	}
}

// DeferStart must reject an empty session like the other call shapes.
func TestDeferStartNoMessages(t *testing.T) {
	c := chat.New(nil).Create("m")
	if _, err := c.DeferStart(context.Background()); err == nil {
		t.Fatal("expected no-messages error")
	}
}
