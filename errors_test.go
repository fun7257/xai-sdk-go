package xai_test

import (
	"errors"
	"testing"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/collections"
	"github.com/fun7257/xai-sdk-go/internal/conn"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
	"google.golang.org/grpc"
)

func TestNewClientErrNoAPIKey(t *testing.T) {
	_, err := xai.NewClient(xai.WithoutEnv())
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, xai.ErrNoAPIKey) {
		t.Fatalf("want ErrNoAPIKey, got %v", err)
	}
	if !errors.Is(err, conn.ErrNoAPIKey) {
		t.Fatalf("want same sentinel as conn.ErrNoAPIKey, got %v", err)
	}
}

func TestNewClientErrNoAPIKeyFromEnv(t *testing.T) {
	t.Setenv("XAI_API_KEY", "")
	_, err := xai.NewClient()
	if err == nil {
		t.Fatal("expected error when XAI_API_KEY empty")
	}
	if !errors.Is(err, xai.ErrNoAPIKey) {
		t.Fatalf("want ErrNoAPIKey, got %v", err)
	}
}

func TestSentinelAliases(t *testing.T) {
	// Public re-exports must be the same sentinel values (errors.Is works across packages).
	if xai.ErrNoAPIKey != conn.ErrNoAPIKey {
		t.Fatal("ErrNoAPIKey alias mismatch")
	}
	if xai.ErrEmptyAPIKey != conn.ErrEmptyAPIKey {
		t.Fatal("ErrEmptyAPIKey alias mismatch")
	}
	if xai.ErrNoManagementKey != collections.ErrNoManagementKey {
		t.Fatal("ErrNoManagementKey alias mismatch")
	}
	// Dial empty key wrapping: internal/conn/conn_test.go TestDialEmptyKey
}

func TestCollectionsErrNoManagementKey(t *testing.T) {
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
	defer func() { _ = cli.Close() }()

	_, err = cli.Collections.Create(t.Context(), "c")
	if err == nil {
		t.Fatal("expected management key error")
	}
	if !errors.Is(err, xai.ErrNoManagementKey) {
		t.Fatalf("want ErrNoManagementKey, got %v", err)
	}
	if !errors.Is(err, collections.ErrNoManagementKey) {
		t.Fatalf("want collections.ErrNoManagementKey, got %v", err)
	}
}
