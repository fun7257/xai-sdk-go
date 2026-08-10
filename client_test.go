package xai_test

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/batch"
	"github.com/fun7257/xai-sdk-go/chat"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	"github.com/fun7257/xai-sdk-go/video"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockAuth struct{ xaiv1.UnimplementedAuthServer }

func (m *mockAuth) GetApiKeyInfo(ctx context.Context, _ *emptypb.Empty) (*xaiv1.ApiKey, error) {
	return &xaiv1.ApiKey{Name: "test-key", TeamId: "team", RedactedApiKey: "xai-***"}, nil
}

type mockModels struct {
	xaiv1.UnimplementedModelsServer
}

func (m *mockModels) ListLanguageModels(ctx context.Context, _ *emptypb.Empty) (*xaiv1.ListLanguageModelsResponse, error) {
	return &xaiv1.ListLanguageModelsResponse{Models: []*xaiv1.LanguageModel{{Name: "grok-3"}}}, nil
}
func (m *mockModels) GetLanguageModel(ctx context.Context, req *xaiv1.GetModelRequest) (*xaiv1.LanguageModel, error) {
	return &xaiv1.LanguageModel{Name: req.Name}, nil
}
func (m *mockModels) ListEmbeddingModels(ctx context.Context, _ *emptypb.Empty) (*xaiv1.ListEmbeddingModelsResponse, error) {
	return &xaiv1.ListEmbeddingModelsResponse{}, nil
}
func (m *mockModels) GetEmbeddingModel(ctx context.Context, req *xaiv1.GetModelRequest) (*xaiv1.EmbeddingModel, error) {
	return &xaiv1.EmbeddingModel{Name: req.Name}, nil
}
func (m *mockModels) ListImageGenerationModels(ctx context.Context, _ *emptypb.Empty) (*xaiv1.ListImageGenerationModelsResponse, error) {
	return &xaiv1.ListImageGenerationModelsResponse{}, nil
}
func (m *mockModels) GetImageGenerationModel(ctx context.Context, req *xaiv1.GetModelRequest) (*xaiv1.ImageGenerationModel, error) {
	return &xaiv1.ImageGenerationModel{Name: req.Name}, nil
}

type mockTok struct {
	xaiv1.UnimplementedTokenizeServer
}

func (m *mockTok) TokenizeText(ctx context.Context, req *xaiv1.TokenizeTextRequest) (*xaiv1.TokenizeTextResponse, error) {
	return &xaiv1.TokenizeTextResponse{Model: req.Model, Tokens: []*xaiv1.Token{{TokenId: 1}}}, nil
}

type mockImage struct{ xaiv1.UnimplementedImageServer }

func (m *mockImage) GenerateImage(ctx context.Context, req *xaiv1.GenerateImageRequest) (*xaiv1.ImageResponse, error) {
	return &xaiv1.ImageResponse{
		Model: req.Model,
		Images: []*xaiv1.GeneratedImage{{
			RespectModeration: true,
			Image:             &xaiv1.GeneratedImage_Url{Url: "https://img.example/1.png"},
		}},
	}, nil
}

type mockVideo struct {
	xaiv1.UnimplementedVideoServer
	n int
}

func (m *mockVideo) GenerateVideo(ctx context.Context, req *xaiv1.GenerateVideoRequest) (*xaiv1.StartDeferredResponse, error) {
	return &xaiv1.StartDeferredResponse{RequestId: "vid-1"}, nil
}
func (m *mockVideo) ExtendVideo(ctx context.Context, req *xaiv1.ExtendVideoRequest) (*xaiv1.StartDeferredResponse, error) {
	return &xaiv1.StartDeferredResponse{RequestId: "vid-ext"}, nil
}
func (m *mockVideo) GetDeferredVideo(ctx context.Context, req *xaiv1.GetDeferredVideoRequest) (*xaiv1.GetDeferredVideoResponse, error) {
	m.n++
	if m.n == 1 {
		return &xaiv1.GetDeferredVideoResponse{Status: xaiv1.DeferredStatus_PENDING}, nil
	}
	return &xaiv1.GetDeferredVideoResponse{
		Status: xaiv1.DeferredStatus_DONE,
		Response: &xaiv1.VideoResponse{
			Model: "grok-imagine-video",
			Video: &xaiv1.GeneratedVideo{Url: "https://v.example/1.mp4", Duration: 5, RespectModeration: true},
		},
	}, nil
}

type mockFiles struct{ xaiv1.UnimplementedFilesServer }

func (m *mockFiles) UploadFile(stream xaiv1.Files_UploadFileServer) error {
	for {
		if _, err := stream.Recv(); err != nil {
			break
		}
	}
	return stream.SendAndClose(&xaiv1.File{Id: "file_1", Filename: "a.txt", Size: 3})
}
func (m *mockFiles) ListFiles(ctx context.Context, req *xaiv1.ListFilesRequest) (*xaiv1.ListFilesResponse, error) {
	return &xaiv1.ListFilesResponse{Data: []*xaiv1.File{{Id: "file_1"}}}, nil
}
func (m *mockFiles) RetrieveFile(ctx context.Context, req *xaiv1.RetrieveFileRequest) (*xaiv1.File, error) {
	return &xaiv1.File{Id: req.FileId}, nil
}
func (m *mockFiles) DeleteFile(ctx context.Context, req *xaiv1.DeleteFileRequest) (*xaiv1.DeleteFileResponse, error) {
	return &xaiv1.DeleteFileResponse{Id: req.FileId, Deleted: true}, nil
}
func (m *mockFiles) RetrieveFileContent(req *xaiv1.RetrieveFileContentRequest, stream xaiv1.Files_RetrieveFileContentServer) error {
	return stream.Send(&xaiv1.FileContentChunk{Data: []byte("abc")})
}
func (m *mockFiles) CreatePublicUrl(ctx context.Context, req *xaiv1.CreatePublicUrlRequest) (*xaiv1.CreatePublicUrlResponse, error) {
	return &xaiv1.CreatePublicUrlResponse{FileId: req.FileId, PublicUrl: "https://public.example/" + req.FileId}, nil
}
func (m *mockFiles) RevokePublicUrl(ctx context.Context, req *xaiv1.RevokePublicUrlRequest) (*xaiv1.RevokePublicUrlResponse, error) {
	return &xaiv1.RevokePublicUrlResponse{FileId: req.FileId}, nil
}

type mockBatch struct {
	xaiv1.UnimplementedBatchMgmtServer
}

func (m *mockBatch) CreateBatch(ctx context.Context, req *xaiv1.CreateBatchRequest) (*xaiv1.Batch, error) {
	return &xaiv1.Batch{BatchId: "b1", Name: req.Name}, nil
}
func (m *mockBatch) AddBatchRequests(ctx context.Context, req *xaiv1.AddBatchRequestsRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (m *mockBatch) GetBatch(ctx context.Context, req *xaiv1.GetBatchRequest) (*xaiv1.Batch, error) {
	return &xaiv1.Batch{BatchId: req.BatchId}, nil
}
func (m *mockBatch) CancelBatch(ctx context.Context, req *xaiv1.CancelBatchRequest) (*xaiv1.Batch, error) {
	return &xaiv1.Batch{BatchId: req.BatchId}, nil
}
func (m *mockBatch) ListBatches(ctx context.Context, req *xaiv1.ListBatchesRequest) (*xaiv1.ListBatchesResponse, error) {
	return &xaiv1.ListBatchesResponse{}, nil
}
func (m *mockBatch) ListBatchRequestMetadata(ctx context.Context, req *xaiv1.ListBatchRequestMetadataRequest) (*xaiv1.ListBatchRequestMetadataResponse, error) {
	return &xaiv1.ListBatchRequestMetadataResponse{}, nil
}
func (m *mockBatch) ListBatchResults(ctx context.Context, req *xaiv1.ListBatchResultsRequest) (*xaiv1.ListBatchResultsResponse, error) {
	return &xaiv1.ListBatchResultsResponse{}, nil
}
func (m *mockBatch) GetBatchRequestResult(ctx context.Context, req *xaiv1.GetBatchRequestResultRequest) (*xaiv1.GetBatchRequestResultResponse, error) {
	return &xaiv1.GetBatchRequestResultResponse{}, nil
}

type mockColl struct {
	xaiv1.UnimplementedCollectionsServer
}

func (m *mockColl) CreateCollection(ctx context.Context, req *xaiv1.CreateCollectionRequest) (*xaiv1.CollectionMetadata, error) {
	return &xaiv1.CollectionMetadata{CollectionId: "col1", CollectionName: req.CollectionName}, nil
}
func (m *mockColl) ListCollections(ctx context.Context, req *xaiv1.ListCollectionsRequest) (*xaiv1.ListCollectionsResponse, error) {
	return &xaiv1.ListCollectionsResponse{}, nil
}
func (m *mockColl) DeleteCollection(ctx context.Context, req *xaiv1.DeleteCollectionRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (m *mockColl) GetCollectionMetadata(ctx context.Context, req *xaiv1.GetCollectionMetadataRequest) (*xaiv1.CollectionMetadata, error) {
	return &xaiv1.CollectionMetadata{CollectionId: req.CollectionId}, nil
}
func (m *mockColl) GenerateCollectionDescription(ctx context.Context, req *xaiv1.GenerateCollectionDescriptionRequest) (*xaiv1.GenerateCollectionDescriptionResponse, error) {
	return &xaiv1.GenerateCollectionDescriptionResponse{Description: "desc"}, nil
}
func (m *mockColl) AddDocumentToCollection(ctx context.Context, req *xaiv1.AddDocumentToCollectionRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (m *mockColl) RemoveDocumentFromCollection(ctx context.Context, req *xaiv1.RemoveDocumentFromCollectionRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (m *mockColl) GetDocumentMetadata(ctx context.Context, req *xaiv1.GetDocumentMetadataRequest) (*xaiv1.DocumentMetadata, error) {
	return &xaiv1.DocumentMetadata{
		Status:       xaiv1.DocumentStatus_DOCUMENT_STATUS_PROCESSED,
		FileMetadata: &xaiv1.FileMetadata{FileId: req.FileId},
	}, nil
}
func (m *mockColl) ListDocuments(ctx context.Context, req *xaiv1.ListDocumentsRequest) (*xaiv1.ListDocumentsResponse, error) {
	return &xaiv1.ListDocumentsResponse{}, nil
}
func (m *mockColl) BatchGetDocuments(ctx context.Context, req *xaiv1.BatchGetDocumentsRequest) (*xaiv1.BatchGetDocumentsResponse, error) {
	return &xaiv1.BatchGetDocumentsResponse{}, nil
}
func (m *mockColl) ReIndexDocument(ctx context.Context, req *xaiv1.ReIndexDocumentRequest) (*emptypb.Empty, error) {
	return &emptypb.Empty{}, nil
}
func (m *mockColl) UpdateCollection(ctx context.Context, req *xaiv1.UpdateCollectionRequest) (*xaiv1.CollectionMetadata, error) {
	return &xaiv1.CollectionMetadata{CollectionId: req.CollectionId}, nil
}
func (m *mockColl) UpdateDocument(ctx context.Context, req *xaiv1.UpdateDocumentRequest) (*xaiv1.DocumentMetadata, error) {
	return &xaiv1.DocumentMetadata{}, nil
}

type mockDocs struct {
	xaiv1.UnimplementedDocumentsServer
}

func (m *mockDocs) Search(ctx context.Context, req *xaiv1.SearchRequest) (*xaiv1.SearchResponse, error) {
	return &xaiv1.SearchResponse{Matches: []*xaiv1.SearchMatch{{FileId: "f1", ChunkContent: "hi"}}}, nil
}

func TestClientDomains(t *testing.T) {
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterAuthServer(s, &mockAuth{})
		xaiv1.RegisterModelsServer(s, &mockModels{})
		xaiv1.RegisterTokenizeServer(s, &mockTok{})
		xaiv1.RegisterImageServer(s, &mockImage{})
		xaiv1.RegisterVideoServer(s, &mockVideo{})
		xaiv1.RegisterFilesServer(s, &mockFiles{})
		xaiv1.RegisterBatchMgmtServer(s, &mockBatch{})
		xaiv1.RegisterCollectionsServer(s, &mockColl{})
		xaiv1.RegisterDocumentsServer(s, &mockDocs{})
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	if _, err := xai.NewClient(xai.WithoutEnv()); err == nil {
		t.Fatal("expected missing key error")
	}

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

	ctx := context.Background()
	info, err := cli.Auth.GetAPIKeyInfo(ctx)
	if err != nil || info.Name != "test-key" {
		t.Fatalf("%v %#v", err, info)
	}
	models, err := cli.Models.ListLanguageModels(ctx)
	if err != nil || len(models) != 1 {
		t.Fatal(err, models)
	}
	toks, err := cli.Tokenize.TokenizeText(ctx, "hi", "grok-3")
	if err != nil || len(toks) != 1 {
		t.Fatal(err)
	}
	img, err := cli.Image.Sample(ctx, "a cat", "grok-imagine-image")
	if err != nil {
		t.Fatal(err)
	}
	u, err := img.URL()
	if err != nil || u == "" {
		t.Fatal(err, u)
	}

	v, err := cli.Video.Generate(ctx, "lake", "grok-imagine-video",
		video.WithPollInterval(time.Millisecond),
		video.WithPollTimeout(time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}
	vu, err := v.URL()
	if err != nil || vu == "" {
		t.Fatal(err, vu)
	}
	ext, err := cli.Video.Extend(ctx, "continue", "m", "https://v.example/in.mp4",
		video.WithPollInterval(time.Millisecond),
		video.WithPollTimeout(time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ext.URL(); err != nil {
		t.Fatal(err)
	}

	f, err := cli.Files.Upload(ctx, "a.txt", strings.NewReader("abc"), 3)
	if err != nil || f.Id != "file_1" {
		t.Fatal(err, f)
	}
	if _, err := cli.Files.List(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.Files.Get(ctx, "file_1"); err != nil {
		t.Fatal(err)
	}
	data, err := cli.Files.Content(ctx, "file_1")
	if err != nil || string(data) != "abc" {
		t.Fatal(err, data)
	}
	pub, err := cli.Files.CreatePublicURL(ctx, "file_1", nil)
	if err != nil || pub.PublicUrl == "" {
		t.Fatal(err, pub)
	}
	if _, err := cli.Files.RevokePublicURL(ctx, "file_1"); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.Files.Delete(ctx, "file_1"); err != nil {
		t.Fatal(err)
	}

	b, err := cli.Batch.Create(ctx, "batch", "")
	if err != nil || b.BatchId != "b1" {
		t.Fatal(err, b)
	}
	br := batch.ChatBatchRequest(&xaiv1.GetCompletionsRequest{
		Model: "m", Messages: []*xaiv1.Message{chat.User("hi")},
	}, "r1")
	if err := cli.Batch.Add(ctx, b.BatchId, br); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.Batch.Get(ctx, b.BatchId); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.Batch.List(ctx, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.Batch.ListBatchResults(ctx, b.BatchId, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.Batch.Cancel(ctx, b.BatchId); err != nil {
		t.Fatal(err)
	}

	col, err := cli.Collections.Create(ctx, "c")
	if err != nil || col.CollectionId != "col1" {
		t.Fatal(err, col)
	}
	if _, err := cli.Collections.Get(ctx, "col1"); err != nil {
		t.Fatal(err)
	}
	if _, err := cli.Collections.List(ctx); err != nil {
		t.Fatal(err)
	}
	desc, err := cli.Collections.GenerateDescription(ctx, "col1")
	if err != nil || desc != "desc" {
		t.Fatal(err, desc)
	}
	sr, err := cli.Collections.SearchSimple(ctx, "q", []string{"col1"}, nil, "")
	if err != nil || len(sr.Matches) != 1 {
		t.Fatal(err, sr)
	}
	doc, err := cli.Collections.WaitForIndexing(ctx, "col1", "f1", time.Second, time.Millisecond)
	if err != nil || doc.Status != xaiv1.DocumentStatus_DOCUMENT_STATUS_PROCESSED {
		t.Fatal(err, doc)
	}

	cliNoMgmt, err := xai.NewClient(xai.WithAPIKey("k"), xai.WithAPIConn(srv.Conn), xai.WithoutEnv())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = cliNoMgmt.Close() }()
	if _, err := cliNoMgmt.Collections.Create(ctx, "x"); err == nil {
		t.Fatal("expected management key error")
	}

	// pure local consumer path
	msg := chat.User("ping")
	if msg.Content[0].GetText() != "ping" {
		t.Fatal(msg)
	}
	_ = io.EOF
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
