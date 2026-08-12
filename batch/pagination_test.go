package batch_test

import (
	"context"
	"sync"
	"testing"

	"google.golang.org/grpc"

	"github.com/fun7257/xai-sdk-go/batch"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

type pagingBatches struct {
	xaiv1.UnimplementedBatchMgmtServer
	mu   sync.Mutex
	reqs []*xaiv1.ListBatchesRequest
}

func (m *pagingBatches) ListBatches(ctx context.Context, req *xaiv1.ListBatchesRequest) (*xaiv1.ListBatchesResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.reqs = append(m.reqs, req)
	if req.GetPaginationToken() == "" {
		tok := "p2"
		return &xaiv1.ListBatchesResponse{
			Batches:         []*xaiv1.Batch{{BatchId: "b1"}},
			PaginationToken: &tok,
		}, nil
	}
	return &xaiv1.ListBatchesResponse{Batches: []*xaiv1.Batch{{BatchId: "b2"}}}, nil
}

func TestListAllBatches(t *testing.T) {
	mock := &pagingBatches{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterBatchMgmtServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := batch.New(srv.Conn)

	all, err := cli.ListAll(context.Background())
	if err != nil || len(all) != 2 || all[0].GetBatchId() != "b1" || all[1].GetBatchId() != "b2" {
		t.Fatalf("ListAll: %v %+v", err, all)
	}
	mock.mu.Lock()
	defer mock.mu.Unlock()
	if len(mock.reqs) != 2 || mock.reqs[1].GetPaginationToken() != "p2" {
		t.Fatalf("pagination wire=%+v", mock.reqs)
	}
}
