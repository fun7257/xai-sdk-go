// Package tokenize exposes text tokenization for a named model.
package tokenize

import (
	"context"

	"google.golang.org/grpc"

	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// Client tokenizes text for a model.
type Client struct{ stub xaiv1.TokenizeClient }

// New creates a Tokenize client.
func New(cc grpc.ClientConnInterface) *Client {
	return &Client{stub: xaiv1.NewTokenizeClient(cc)}
}

// TokenizeText returns tokens for text under the given model.
func (c *Client) TokenizeText(ctx context.Context, text, model string) ([]*xaiv1.Token, error) {
	r, err := c.stub.TokenizeText(ctx, &xaiv1.TokenizeTextRequest{Text: text, Model: model})
	if err != nil {
		return nil, err
	}
	return r.Tokens, nil
}
