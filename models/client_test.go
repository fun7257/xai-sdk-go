package models_test

import (
	"context"
	"testing"

	"github.com/fun7257/xai-sdk-go/internal/testutil"
	"github.com/fun7257/xai-sdk-go/models"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockModels struct {
	xaiv1.UnimplementedModelsServer
}

func (m *mockModels) ListLanguageModels(ctx context.Context, _ *emptypb.Empty) (*xaiv1.ListLanguageModelsResponse, error) {
	return &xaiv1.ListLanguageModelsResponse{
		Models: []*xaiv1.LanguageModel{{Name: "grok-3"}, {Name: "grok-4.5"}},
	}, nil
}

func (m *mockModels) GetLanguageModel(ctx context.Context, req *xaiv1.GetModelRequest) (*xaiv1.LanguageModel, error) {
	return &xaiv1.LanguageModel{Name: req.GetName()}, nil
}

func (m *mockModels) ListEmbeddingModels(ctx context.Context, _ *emptypb.Empty) (*xaiv1.ListEmbeddingModelsResponse, error) {
	return &xaiv1.ListEmbeddingModelsResponse{
		Models: []*xaiv1.EmbeddingModel{{Name: "embed-test"}},
	}, nil
}

func (m *mockModels) GetEmbeddingModel(ctx context.Context, req *xaiv1.GetModelRequest) (*xaiv1.EmbeddingModel, error) {
	return &xaiv1.EmbeddingModel{Name: req.GetName()}, nil
}

func (m *mockModels) ListImageGenerationModels(ctx context.Context, _ *emptypb.Empty) (*xaiv1.ListImageGenerationModelsResponse, error) {
	return &xaiv1.ListImageGenerationModelsResponse{
		Models: []*xaiv1.ImageGenerationModel{{Name: "grok-imagine-image"}},
	}, nil
}

func (m *mockModels) GetImageGenerationModel(ctx context.Context, req *xaiv1.GetModelRequest) (*xaiv1.ImageGenerationModel, error) {
	return &xaiv1.ImageGenerationModel{Name: req.GetName()}, nil
}

func TestListAndGetModels(t *testing.T) {
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterModelsServer(s, &mockModels{})
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	cli := models.New(srv.Conn)
	ctx := context.Background()

	langs, err := cli.ListLanguageModels(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(langs) != 2 || langs[0].GetName() != "grok-3" {
		t.Fatalf("langs=%v", langs)
	}
	m, err := cli.GetLanguageModel(ctx, "grok-3")
	if err != nil || m.GetName() != "grok-3" {
		t.Fatalf("get lang: %v %#v", err, m)
	}

	emb, err := cli.ListEmbeddingModels(ctx)
	if err != nil || len(emb) != 1 || emb[0].GetName() != "embed-test" {
		t.Fatalf("emb: %v %v", err, emb)
	}
	em, err := cli.GetEmbeddingModel(ctx, "embed-test")
	if err != nil || em.GetName() != "embed-test" {
		t.Fatalf("get emb: %v %#v", err, em)
	}

	imgs, err := cli.ListImageGenerationModels(ctx)
	if err != nil || len(imgs) != 1 {
		t.Fatalf("imgs: %v %v", err, imgs)
	}
	im, err := cli.GetImageGenerationModel(ctx, "grok-imagine-image")
	if err != nil || im.GetName() != "grok-imagine-image" {
		t.Fatalf("get img model: %v %#v", err, im)
	}
}
