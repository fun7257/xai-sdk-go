package xai_test

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// Minimal mock kept for errors_test.go (same package) and wiring smoke.
type mockAuth struct{ xaiv1.UnimplementedAuthServer }

func (m *mockAuth) GetApiKeyInfo(ctx context.Context, _ *emptypb.Empty) (*xaiv1.ApiKey, error) {
	return &xaiv1.ApiKey{Name: "test-key", TeamId: "team", RedactedApiKey: "xai-***"}, nil
}

// TestNewClientWiring proves NewClient attaches all domain clients and that
// dual injected conns (API + management) wire without dial. Domain RPC happy
// paths live in auth/, models/, tokenize/, image/, video/, files/, batch/,
// collections/ package tests.
func TestNewClientWiring(t *testing.T) {
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterAuthServer(s, &mockAuth{})
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	cli, err := xai.NewClient(
		xai.WithAPIKey("test-key"),
		xai.WithManagementAPIKey("mgmt-key"),
		xai.WithAPIConn(srv.Conn),
		xai.WithManagementConn(srv.Conn),
		xai.WithoutEnv(),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cli.Close() }()

	if cli.Auth == nil || cli.Batch == nil || cli.Chat == nil || cli.Collections == nil ||
		cli.Files == nil || cli.Image == nil || cli.Models == nil || cli.Tokenize == nil || cli.Video == nil {
		t.Fatalf("nil domain client(s): %+v", cli)
	}

	// One light RPC through root-attached Auth proves the injected API conn is live.
	info, err := cli.Auth.GetAPIKeyInfo(context.Background())
	if err != nil || info.GetName() != "test-key" {
		t.Fatalf("auth via wired client: %v %#v", err, info)
	}
}

func TestCloseWithoutOwnedConn(t *testing.T) {
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterAuthServer(s, &mockAuth{})
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli, err := xai.NewClient(xai.WithAPIKey("k"), xai.WithAPIConn(srv.Conn), xai.WithoutEnv())
	if err != nil {
		t.Fatal(err)
	}
	if err := cli.Close(); err != nil {
		t.Fatal(err)
	}
}
