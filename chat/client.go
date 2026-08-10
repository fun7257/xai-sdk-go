// Package chat implements multi-turn completions, streaming, tools, deferred
// requests, structured Parse, and stored/compact helpers.
//
// Preferred call shapes: Sample / Samples+WithN, StreamReader, Defer / Defers+WithDeferN.
// See docs/PARITY.md and docs/DIFF.md for Go-first design notes.
package chat

import (
	"context"

	"google.golang.org/grpc"

	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// Client is the chat domain client.
type Client struct {
	stub xaiv1.ChatClient
}

// New creates a Chat client.
func New(cc grpc.ClientConnInterface) *Client {
	return &Client{stub: xaiv1.NewChatClient(cc)}
}

// GetStoredCompletion retrieves a stored response by id.
func (c *Client) GetStoredCompletion(ctx context.Context, responseID string) ([]*Response, error) {
	pb, err := c.stub.GetStoredCompletion(ctx, &xaiv1.GetStoredCompletionRequest{ResponseId: responseID})
	if err != nil {
		return nil, err
	}
	out := make([]*Response, 0, len(pb.Outputs))
	for i := range pb.Outputs {
		idx := i
		out = append(out, NewResponse(pb, &idx))
	}
	return out, nil
}

// DeleteStoredCompletion deletes a stored response.
func (c *Client) DeleteStoredCompletion(ctx context.Context, responseID string) (string, error) {
	pb, err := c.stub.DeleteStoredCompletion(ctx, &xaiv1.DeleteStoredCompletionRequest{ResponseId: responseID})
	if err != nil {
		return "", err
	}
	return pb.ResponseId, nil
}

// CompactContext compacts messages without a Chat session.
func (c *Client) CompactContext(ctx context.Context, model string, messages []*xaiv1.Message) (*CompactContextResponse, error) {
	pb, err := c.stub.CompactContext(ctx, &xaiv1.CompactContextRequest{Model: model, Input: messages})
	if err != nil {
		return nil, err
	}
	return &CompactContextResponse{proto: pb}, nil
}
