package tokenize_test

import (
	"context"
	"testing"

	"github.com/fun7257/xai-sdk-go/internal/testutil"
	"github.com/fun7257/xai-sdk-go/tokenize"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
	"google.golang.org/grpc"
)

type mockTokenize struct {
	xaiv1.UnimplementedTokenizeServer
	lastText  string
	lastModel string
}

func (m *mockTokenize) TokenizeText(ctx context.Context, req *xaiv1.TokenizeTextRequest) (*xaiv1.TokenizeTextResponse, error) {
	m.lastText = req.GetText()
	m.lastModel = req.GetModel()
	return &xaiv1.TokenizeTextResponse{
		Tokens: []*xaiv1.Token{
			{TokenId: 1, StringToken: "hel"},
			{TokenId: 2, StringToken: "lo"},
		},
	}, nil
}

func TestTokenizeText(t *testing.T) {
	mock := &mockTokenize{}
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterTokenizeServer(s, mock)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	cli := tokenize.New(srv.Conn)
	toks, err := cli.TokenizeText(context.Background(), "hello", "grok-3")
	if err != nil {
		t.Fatal(err)
	}
	if mock.lastText != "hello" || mock.lastModel != "grok-3" {
		t.Fatalf("request text=%q model=%q", mock.lastText, mock.lastModel)
	}
	if len(toks) != 2 {
		t.Fatalf("len=%d", len(toks))
	}
	if toks[0].GetStringToken() != "hel" || toks[1].GetTokenId() != 2 {
		t.Fatalf("tokens=%v", toks)
	}
}
