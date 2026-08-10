// Package auth retrieves API key metadata for the credential on a connection.
package auth

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// Client retrieves API key metadata.
type Client struct{ stub xaiv1.AuthClient }

// New creates an Auth client.
func New(cc grpc.ClientConnInterface) *Client {
	return &Client{stub: xaiv1.NewAuthClient(cc)}
}

// GetAPIKeyInfo returns metadata for the key used on this connection.
func (c *Client) GetAPIKeyInfo(ctx context.Context) (*xaiv1.ApiKey, error) {
	return c.stub.GetApiKeyInfo(ctx, &emptypb.Empty{})
}
