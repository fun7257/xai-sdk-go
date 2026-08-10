package chat_test

import (
	"context"
	"fmt"
	"io"

	"google.golang.org/grpc"

	"github.com/fun7257/xai-sdk-go/chat"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// Preferred chat path: Sample for a single completion (offline bufconn mock).
func ExampleChat_Sample() {
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterChatServer(s, &exampleChat{})
	})
	if err != nil {
		fmt.Println("start:", err)
		return
	}
	defer srv.Close()

	c := chat.New(srv.Conn).Create("grok-3",
		chat.WithMessages(
			chat.System("You are helpful."),
			chat.User("Hello!"),
		),
	)
	resp, err := c.Sample(context.Background())
	if err != nil {
		fmt.Println("sample:", err)
		return
	}
	fmt.Println(resp.Content())
	// Output: hello
}

// StreamReader is the primary streaming API (Recv until io.EOF).
func ExampleChat_StreamReader() {
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterChatServer(s, &exampleChat{})
	})
	if err != nil {
		fmt.Println("start:", err)
		return
	}
	defer srv.Close()

	c := chat.New(srv.Conn).Create("grok-3", chat.WithMessages(chat.User("hi")))
	sr, err := c.StreamReader(context.Background())
	if err != nil {
		fmt.Println("stream:", err)
		return
	}
	defer func() { _ = sr.Close() }()

	var text string
	for {
		ev, err := sr.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("recv:", err)
			return
		}
		if ev.Chunk != nil {
			text += ev.Chunk.Content()
		}
	}
	fmt.Println(text)
	// Output: hello
}

type exampleChat struct{ xaiv1.UnimplementedChatServer }

func (exampleChat) GetCompletion(ctx context.Context, req *xaiv1.GetCompletionsRequest) (*xaiv1.GetChatCompletionResponse, error) {
	return &xaiv1.GetChatCompletionResponse{
		Id: "ex-1", Model: req.Model,
		Outputs: []*xaiv1.CompletionOutput{{
			Message: &xaiv1.CompletionMessage{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: "hello"},
		}},
	}, nil
}

func (exampleChat) GetCompletionChunk(req *xaiv1.GetCompletionsRequest, stream xaiv1.Chat_GetCompletionChunkServer) error {
	for _, part := range []string{"hel", "lo"} {
		if err := stream.Send(&xaiv1.GetChatCompletionChunk{
			Id: "ex-s", Model: req.Model,
			Outputs: []*xaiv1.CompletionOutputChunk{{
				Delta: &xaiv1.Delta{Role: xaiv1.MessageRole_ROLE_ASSISTANT, Content: part},
			}},
		}); err != nil {
			return err
		}
	}
	return nil
}
