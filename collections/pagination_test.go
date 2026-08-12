package collections_test

import (
	"context"
	"sync"
	"testing"

	"google.golang.org/grpc"

	"github.com/fun7257/xai-sdk-go/collections"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// pagingColl serves two pages of collections and documents.
type pagingColl struct {
	xaiv1.UnimplementedCollectionsServer
	mu       sync.Mutex
	collReqs []*xaiv1.ListCollectionsRequest
	docReqs  []*xaiv1.ListDocumentsRequest
}

func (m *pagingColl) ListCollections(ctx context.Context, req *xaiv1.ListCollectionsRequest) (*xaiv1.ListCollectionsResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.collReqs = append(m.collReqs, req)
	if req.GetPaginationToken() == "" {
		return &xaiv1.ListCollectionsResponse{
			Collections:     []*xaiv1.CollectionMetadata{{CollectionId: "c1"}},
			PaginationToken: "next",
		}, nil
	}
	return &xaiv1.ListCollectionsResponse{
		Collections: []*xaiv1.CollectionMetadata{{CollectionId: "c2"}},
	}, nil
}

func (m *pagingColl) ListDocuments(ctx context.Context, req *xaiv1.ListDocumentsRequest) (*xaiv1.ListDocumentsResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.docReqs = append(m.docReqs, req)
	if req.GetPaginationToken() == "" {
		return &xaiv1.ListDocumentsResponse{
			Documents:       []*xaiv1.DocumentMetadata{{FileMetadata: &xaiv1.FileMetadata{FileId: "f1"}}},
			PaginationToken: "next",
		}, nil
	}
	return &xaiv1.ListDocumentsResponse{
		Documents: []*xaiv1.DocumentMetadata{{FileMetadata: &xaiv1.FileMetadata{FileId: "f2"}}},
	}, nil
}

func TestListAllCollectionsAndDocuments(t *testing.T) {
	mock := &pagingColl{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterCollectionsServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := collections.New(srv.Conn, srv.Conn)
	ctx := context.Background()

	cols, err := cli.ListAll(ctx)
	if err != nil || len(cols) != 2 || cols[1].GetCollectionId() != "c2" {
		t.Fatalf("ListAll: %v %+v", err, cols)
	}

	docs, err := cli.ListAllDocuments(ctx, "c1", collections.WithDocumentsLimit(1))
	if err != nil || len(docs) != 2 || docs[1].GetFileMetadata().GetFileId() != "f2" {
		t.Fatalf("ListAllDocuments: %v %+v", err, docs)
	}
	mock.mu.Lock()
	defer mock.mu.Unlock()
	if len(mock.docReqs) != 2 || mock.docReqs[1].GetPaginationToken() != "next" {
		t.Fatalf("doc pagination wire=%+v", mock.docReqs)
	}
	if mock.docReqs[0].GetLimit() != 1 || mock.docReqs[1].GetLimit() != 1 {
		t.Fatal("options must apply to every page")
	}
}

func TestWithDocumentsPaginationTokenWire(t *testing.T) {
	mock := &pagingColl{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterCollectionsServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := collections.New(srv.Conn, srv.Conn)

	if _, err := cli.ListDocuments(context.Background(), "c1",
		collections.WithDocumentsPaginationToken("resume-here")); err != nil {
		t.Fatal(err)
	}
	mock.mu.Lock()
	defer mock.mu.Unlock()
	if len(mock.docReqs) != 1 || mock.docReqs[0].GetPaginationToken() != "resume-here" {
		t.Fatalf("token wire=%+v", mock.docReqs)
	}
}
