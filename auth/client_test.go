package auth_test

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/fun7257/xai-sdk-go/auth"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

type mockAuth struct {
	xaiv1.UnimplementedAuthServer
	redacted string
}

func (m *mockAuth) GetApiKeyInfo(ctx context.Context, _ *emptypb.Empty) (*xaiv1.ApiKey, error) {
	return &xaiv1.ApiKey{RedactedApiKey: m.redacted, Name: "test-key"}, nil
}

func TestGetAPIKeyInfo(t *testing.T) {
	mock := &mockAuth{redacted: "xai-***"}
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterAuthServer(s, mock)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	cli := auth.New(srv.Conn)
	info, err := cli.GetAPIKeyInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info == nil {
		t.Fatal("nil ApiKey")
	}
	if info.GetRedactedApiKey() != "xai-***" {
		t.Fatalf("redacted=%q", info.GetRedactedApiKey())
	}
	if info.GetName() != "test-key" {
		t.Fatalf("name=%q", info.GetName())
	}
}
