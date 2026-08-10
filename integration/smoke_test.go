//go:build integration

// Package integration holds optional live API smoke tests.
//
// Default `go test ./...` excludes this package (build tag). Run intentionally:
//
//	export XAI_API_KEY=xai-...
//	go test -tags=integration ./integration/ -count=1 -v
//
// Or: make integration
//
// Without XAI_API_KEY, tests skip (do not fail).
package integration_test

import (
	"context"
	"os"
	"testing"
	"time"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/chat"
	"github.com/fun7257/xai-sdk-go/types"
)

func requireAPIKey(t *testing.T) string {
	t.Helper()
	key := os.Getenv("XAI_API_KEY")
	if key == "" {
		t.Skip("XAI_API_KEY not set; skipping live integration smoke")
	}
	return key
}

func TestLiveAuthAndChatSample(t *testing.T) {
	key := requireAPIKey(t)

	client, err := xai.NewClient(
		xai.WithAPIKey(key),
		xai.WithTimeout(2*time.Minute),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	defer func() { _ = client.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	info, err := client.Auth.GetAPIKeyInfo(ctx)
	if err != nil {
		t.Fatalf("GetAPIKeyInfo: %v", err)
	}
	if info == nil {
		t.Fatal("nil ApiKey info")
	}

	model := os.Getenv("SMOKE_MODEL")
	if model == "" {
		model = types.ModelGrok45Latest
	}
	c := client.Chat.Create(model,
		chat.WithMessages(chat.User("Reply with exactly: pong")),
		chat.WithMaxTokens(16),
	)
	resp, err := c.Sample(ctx)
	if err != nil {
		t.Fatalf("Sample: %v", err)
	}
	if resp.Content() == "" {
		t.Fatal("empty completion content")
	}
	t.Logf("version=%s model=%s content=%q", xai.Version, model, resp.Content())
}
