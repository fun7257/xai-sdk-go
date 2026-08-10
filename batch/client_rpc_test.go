package batch_test

import (
	"context"
	"testing"

	"github.com/fun7257/xai-sdk-go/batch"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type mockBatch struct {
	xaiv1.UnimplementedBatchMgmtServer
	createdName string
	added       int
	gotID       string
	canceled    string
	listCalled  bool
}

func (m *mockBatch) CreateBatch(ctx context.Context, req *xaiv1.CreateBatchRequest) (*xaiv1.Batch, error) {
	m.createdName = req.GetName()
	id := req.GetInputFileId()
	b := &xaiv1.Batch{BatchId: "b1", Name: req.GetName()}
	if id != "" {
		b.InputFileId = &id
	}
	return b, nil
}

func (m *mockBatch) AddBatchRequests(ctx context.Context, req *xaiv1.AddBatchRequestsRequest) (*emptypb.Empty, error) {
	m.added = len(req.GetBatchRequests())
	return &emptypb.Empty{}, nil
}

func (m *mockBatch) GetBatch(ctx context.Context, req *xaiv1.GetBatchRequest) (*xaiv1.Batch, error) {
	m.gotID = req.GetBatchId()
	return &xaiv1.Batch{BatchId: req.GetBatchId(), Name: "n"}, nil
}

func (m *mockBatch) CancelBatch(ctx context.Context, req *xaiv1.CancelBatchRequest) (*xaiv1.Batch, error) {
	m.canceled = req.GetBatchId()
	return &xaiv1.Batch{BatchId: req.GetBatchId()}, nil
}

func (m *mockBatch) ListBatches(ctx context.Context, req *xaiv1.ListBatchesRequest) (*xaiv1.ListBatchesResponse, error) {
	m.listCalled = true
	return &xaiv1.ListBatchesResponse{Batches: []*xaiv1.Batch{{BatchId: "b1"}}}, nil
}

func (m *mockBatch) ListBatchRequestMetadata(ctx context.Context, req *xaiv1.ListBatchRequestMetadataRequest) (*xaiv1.ListBatchRequestMetadataResponse, error) {
	return &xaiv1.ListBatchRequestMetadataResponse{
		BatchRequestMetadata: []*xaiv1.BatchRequestMetadata{{BatchRequestId: "r1"}},
	}, nil
}

func (m *mockBatch) ListBatchResults(ctx context.Context, req *xaiv1.ListBatchResultsRequest) (*xaiv1.ListBatchResultsResponse, error) {
	return &xaiv1.ListBatchResultsResponse{
		Results: []*xaiv1.BatchResult{{BatchRequestId: "r1"}},
	}, nil
}

func (m *mockBatch) GetBatchRequestResult(ctx context.Context, req *xaiv1.GetBatchRequestResultRequest) (*xaiv1.GetBatchRequestResultResponse, error) {
	return &xaiv1.GetBatchRequestResultResponse{
		Request: &xaiv1.BatchRequest{},
		Result:  &xaiv1.BatchResult{BatchRequestId: req.GetBatchRequestId()},
	}, nil
}

func TestClientRPCMethods(t *testing.T) {
	mock := &mockBatch{}
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterBatchMgmtServer(s, mock)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	cli := batch.New(srv.Conn)
	ctx := context.Background()

	b, err := cli.Create(ctx, "job", "file-1")
	if err != nil || b.GetBatchId() != "b1" || mock.createdName != "job" {
		t.Fatalf("create: %v %#v mock=%+v", err, b, mock)
	}

	chatReq := &xaiv1.GetCompletionsRequest{Model: "grok-3"}
	imgReq := &xaiv1.GenerateImageRequest{Prompt: "cat"}
	vidReq := &xaiv1.GenerateVideoRequest{Prompt: "run"}
	brs := []*xaiv1.BatchRequest{
		batch.ChatBatchRequest(chatReq, "c1"),
		batch.ImageBatchRequest(imgReq, "i1"),
		batch.VideoBatchRequest(vidReq, "v1"),
	}
	if brs[0].GetCompletionRequest() == nil || brs[0].GetBatchRequestId() != "c1" {
		t.Fatalf("chat batch req: %#v", brs[0])
	}
	if brs[1].GetImageRequest() == nil || brs[2].GetVideoRequest() == nil {
		t.Fatalf("image/video batch reqs")
	}

	if err := cli.Add(ctx, "b1", brs...); err != nil || mock.added != 3 {
		t.Fatalf("add: %v added=%d", err, mock.added)
	}

	got, err := cli.Get(ctx, "b1")
	if err != nil || got.GetBatchId() != "b1" || mock.gotID != "b1" {
		t.Fatalf("get: %v %#v", err, got)
	}

	canceled, err := cli.Cancel(ctx, "b1")
	if err != nil || canceled.GetBatchId() != "b1" || mock.canceled != "b1" {
		t.Fatalf("cancel: %v %#v", err, canceled)
	}

	listed, err := cli.List(ctx, "tok")
	if err != nil || len(listed.GetBatches()) != 1 || !mock.listCalled {
		t.Fatalf("list: %v %#v", err, listed)
	}
	// empty pagination path
	if _, err := cli.List(ctx, ""); err != nil {
		t.Fatal(err)
	}

	meta, err := cli.ListBatchRequests(ctx, "b1", "p")
	if err != nil || len(meta.GetBatchRequestMetadata()) != 1 {
		t.Fatalf("list reqs: %v %#v", err, meta)
	}
	if _, err := cli.ListBatchRequests(ctx, "b1", ""); err != nil {
		t.Fatal(err)
	}

	res, err := cli.ListBatchResults(ctx, "b1", "p")
	if err != nil || len(res.GetResults()) != 1 {
		t.Fatalf("list results: %v %#v", err, res)
	}
	if _, err := cli.ListBatchResults(ctx, "b1", ""); err != nil {
		t.Fatal(err)
	}

	one, err := cli.GetBatchRequestResult(ctx, "b1", "r9")
	if err != nil || one.GetResult().GetBatchRequestId() != "r9" {
		t.Fatalf("get result: %v %#v", err, one)
	}

	wrapped := batch.WrapResults(res.GetResults())
	if len(wrapped) != 1 || wrapped[0].BatchRequestID() != "r1" {
		t.Fatalf("wrap: %#v", wrapped)
	}
	if wrapped[0].Proto() == nil {
		t.Fatal("proto nil")
	}
	// nil-safe helpers
	if batch.NewResult(nil).BatchRequestID() != "" {
		t.Fatal("nil result id")
	}
	if (*batch.Result)(nil).Succeeded() || (*batch.Result)(nil).Failed() {
		t.Fatal("nil result status")
	}
}
