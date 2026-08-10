// Package models lists and retrieves language, embedding, and image-generation models.
package models

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// Client lists and retrieves model metadata.
type Client struct{ stub xaiv1.ModelsClient }

// New creates a Models client.
func New(cc grpc.ClientConnInterface) *Client {
	return &Client{stub: xaiv1.NewModelsClient(cc)}
}

// ListLanguageModels returns available language models.
func (c *Client) ListLanguageModels(ctx context.Context) ([]*xaiv1.LanguageModel, error) {
	r, err := c.stub.ListLanguageModels(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	return r.Models, nil
}

// GetLanguageModel returns a language model by name.
func (c *Client) GetLanguageModel(ctx context.Context, name string) (*xaiv1.LanguageModel, error) {
	return c.stub.GetLanguageModel(ctx, &xaiv1.GetModelRequest{Name: name})
}

// ListEmbeddingModels returns available embedding models.
func (c *Client) ListEmbeddingModels(ctx context.Context) ([]*xaiv1.EmbeddingModel, error) {
	r, err := c.stub.ListEmbeddingModels(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	return r.Models, nil
}

// GetEmbeddingModel returns an embedding model by name.
func (c *Client) GetEmbeddingModel(ctx context.Context, name string) (*xaiv1.EmbeddingModel, error) {
	return c.stub.GetEmbeddingModel(ctx, &xaiv1.GetModelRequest{Name: name})
}

// ListImageGenerationModels returns available image generation models.
func (c *Client) ListImageGenerationModels(ctx context.Context) ([]*xaiv1.ImageGenerationModel, error) {
	r, err := c.stub.ListImageGenerationModels(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, err
	}
	return r.Models, nil
}

// GetImageGenerationModel returns an image generation model by name.
func (c *Client) GetImageGenerationModel(ctx context.Context, name string) (*xaiv1.ImageGenerationModel, error) {
	return c.stub.GetImageGenerationModel(ctx, &xaiv1.GetModelRequest{Name: name})
}
