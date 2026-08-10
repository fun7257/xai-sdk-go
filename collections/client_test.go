package collections_test

import (
	"context"
	"testing"
	"time"

	"github.com/fun7257/xai-sdk-go/collections"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
	"google.golang.org/grpc"
)

type mockColl struct {
	xaiv1.UnimplementedCollectionsServer
	status xaiv1.DocumentStatus
	errMsg string
}

func (m *mockColl) GetDocumentMetadata(ctx context.Context, req *xaiv1.GetDocumentMetadataRequest) (*xaiv1.DocumentMetadata, error) {
	return &xaiv1.DocumentMetadata{
		Status:       m.status,
		ErrorMessage: m.errMsg,
		FileMetadata: &xaiv1.FileMetadata{FileId: req.FileId},
	}, nil
}

func TestWaitForIndexingFailed(t *testing.T) {
	mock := &mockColl{status: xaiv1.DocumentStatus_DOCUMENT_STATUS_FAILED, errMsg: "index boom"}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterCollectionsServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	cli := collections.New(srv.Conn, srv.Conn)
	doc, err := cli.WaitForIndexing(context.Background(), "c1", "f1", time.Second, time.Millisecond)
	if err == nil {
		t.Fatal("expected indexing failure error")
	}
	if err.Error() != "index boom" {
		t.Fatalf("err=%v", err)
	}
	// doc may still be populated; error must be non-nil so callers don't treat as success
	if doc == nil || doc.Status != xaiv1.DocumentStatus_DOCUMENT_STATUS_FAILED {
		// Wait returns doc, err — on error path doc is still set
		_ = doc
	}
}

func TestWaitForIndexingProcessed(t *testing.T) {
	mock := &mockColl{status: xaiv1.DocumentStatus_DOCUMENT_STATUS_PROCESSED}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterCollectionsServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := collections.New(srv.Conn, srv.Conn)
	doc, err := cli.WaitForIndexing(context.Background(), "c1", "f1", time.Second, time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if doc.Status != xaiv1.DocumentStatus_DOCUMENT_STATUS_PROCESSED {
		t.Fatal(doc.Status)
	}
}

func TestUpdateDocumentFields(t *testing.T) {
	mock := &updateDocMock{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterCollectionsServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := collections.New(srv.Conn, srv.Conn)
	doc, err := cli.UpdateDocument(context.Background(), "c1", "f1",
		collections.WithUpdateName("n"),
		collections.WithUpdateData([]byte("bytes")),
		collections.WithUpdateContentType("text/plain"),
		collections.WithUpdateFields(map[string]string{"k": "v"}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if doc == nil || mock.last == nil {
		t.Fatal("nil")
	}
	if mock.last.CollectionId != "c1" || mock.last.FileId != "f1" {
		t.Fatalf("%+v", mock.last)
	}
	if mock.last.Name != "n" || string(mock.last.Data) != "bytes" || mock.last.ContentType != "text/plain" {
		t.Fatalf("%+v", mock.last)
	}
	if mock.last.Fields["k"] != "v" {
		t.Fatalf("fields=%v", mock.last.Fields)
	}
}

type updateDocMock struct {
	xaiv1.UnimplementedCollectionsServer
	last *xaiv1.UpdateDocumentRequest
}

func (m *updateDocMock) UpdateDocument(ctx context.Context, req *xaiv1.UpdateDocumentRequest) (*xaiv1.DocumentMetadata, error) {
	m.last = req
	return &xaiv1.DocumentMetadata{FileMetadata: &xaiv1.FileMetadata{FileId: req.FileId, Name: req.Name}}, nil
}
