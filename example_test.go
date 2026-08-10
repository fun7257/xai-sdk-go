package xai_test

import (
	"context"
	"fmt"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/chat"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Offline construction: inject a bufconn API connection so Example does not dial the public API.
func ExampleNewClient() {
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterAuthServer(s, &exampleAuth{})
	})
	if err != nil {
		fmt.Println("start:", err)
		return
	}
	defer srv.Close()

	client, err := xai.NewClient(
		xai.WithAPIKey("example-key"),
		xai.WithAPIConn(srv.Conn),
		xai.WithoutEnv(),
	)
	if err != nil {
		fmt.Println("new:", err)
		return
	}
	defer func() { _ = client.Close() }()

	// Domain clients are ready; preferred chat shape uses Sample / Samples+WithN.
	_ = client.Chat.Create("grok-3", chat.WithMessages(chat.User("hello")))
	fmt.Println("ok")
	// Output: ok
}

type exampleAuth struct{ xaiv1.UnimplementedAuthServer }

func (exampleAuth) GetApiKeyInfo(context.Context, *emptypb.Empty) (*xaiv1.ApiKey, error) {
	return &xaiv1.ApiKey{Name: "example"}, nil
}
