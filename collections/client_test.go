package collections_test

import (
	"context"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/fun7257/xai-sdk-go/collections"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
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

// mockCollCRUD exercises Create/List/Get/GenerateDescription management RPCs.
type mockCollCRUD struct {
	xaiv1.UnimplementedCollectionsServer
	createdName string
	listed      bool
	gotID       string
	genID       string
}

func (m *mockCollCRUD) CreateCollection(ctx context.Context, req *xaiv1.CreateCollectionRequest) (*xaiv1.CollectionMetadata, error) {
	m.createdName = req.GetCollectionName()
	return &xaiv1.CollectionMetadata{CollectionId: "col1", CollectionName: req.GetCollectionName()}, nil
}

func (m *mockCollCRUD) ListCollections(ctx context.Context, req *xaiv1.ListCollectionsRequest) (*xaiv1.ListCollectionsResponse, error) {
	m.listed = true
	return &xaiv1.ListCollectionsResponse{
		Collections: []*xaiv1.CollectionMetadata{{CollectionId: "col1", CollectionName: "c"}},
	}, nil
}

func (m *mockCollCRUD) GetCollectionMetadata(ctx context.Context, req *xaiv1.GetCollectionMetadataRequest) (*xaiv1.CollectionMetadata, error) {
	m.gotID = req.GetCollectionId()
	return &xaiv1.CollectionMetadata{CollectionId: req.GetCollectionId()}, nil
}

func (m *mockCollCRUD) GenerateCollectionDescription(ctx context.Context, req *xaiv1.GenerateCollectionDescriptionRequest) (*xaiv1.GenerateCollectionDescriptionResponse, error) {
	m.genID = req.GetCollectionId()
	return &xaiv1.GenerateCollectionDescriptionResponse{Description: "auto-desc"}, nil
}

type mockDocs struct {
	xaiv1.UnimplementedDocumentsServer
	last *xaiv1.SearchRequest
}

func (m *mockDocs) Search(ctx context.Context, req *xaiv1.SearchRequest) (*xaiv1.SearchResponse, error) {
	m.last = req
	return &xaiv1.SearchResponse{Matches: []*xaiv1.SearchMatch{{FileId: "f1", ChunkContent: "hi"}}}, nil
}

// TestCRUDAndSearch covers Create/List/Get/GenerateDescription/SearchSimple via
// shipped collections.Client (sole domain coverage after root mega-test removal).
func TestCRUDAndSearch(t *testing.T) {
	mgmt := &mockCollCRUD{}
	docs := &mockDocs{}
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterCollectionsServer(s, mgmt)
		xaiv1.RegisterDocumentsServer(s, docs)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	cli := collections.New(srv.Conn, srv.Conn)
	ctx := context.Background()

	col, err := cli.Create(ctx, "my-col")
	if err != nil || col.GetCollectionId() != "col1" || mgmt.createdName != "my-col" {
		t.Fatalf("Create: %v %#v mock=%+v", err, col, mgmt)
	}

	list, err := cli.List(ctx)
	if err != nil || !mgmt.listed || len(list.GetCollections()) != 1 {
		t.Fatalf("List: %v %#v listed=%v", err, list, mgmt.listed)
	}

	got, err := cli.Get(ctx, "col1")
	if err != nil || got.GetCollectionId() != "col1" || mgmt.gotID != "col1" {
		t.Fatalf("Get: %v %#v last=%q", err, got, mgmt.gotID)
	}

	desc, err := cli.GenerateDescription(ctx, "col1")
	if err != nil || desc != "auto-desc" || mgmt.genID != "col1" {
		t.Fatalf("GenerateDescription: %v %q last=%q", err, desc, mgmt.genID)
	}

	// SearchSimple uses the business Documents channel (api conn), not management.
	lim := int32(3)
	sr, err := cli.SearchSimple(ctx, "query", []string{"col1"}, &lim, "be brief")
	if err != nil || len(sr.GetMatches()) != 1 || sr.GetMatches()[0].GetFileId() != "f1" {
		t.Fatalf("SearchSimple: %v %#v", err, sr)
	}
	if docs.last == nil || docs.last.GetQuery() != "query" {
		t.Fatalf("Search wire=%+v", docs.last)
	}
	if docs.last.GetLimit() != 3 || docs.last.GetInstructions() != "be brief" {
		t.Fatalf("SearchSimple opts wire limit=%v instructions=%q", docs.last.GetLimit(), docs.last.GetInstructions())
	}
	if len(docs.last.GetSource().GetCollectionIds()) != 1 || docs.last.GetSource().GetCollectionIds()[0] != "col1" {
		t.Fatalf("Search source=%+v", docs.last.GetSource())
	}
}

func TestWaitForIndexing(t *testing.T) {
	tests := []struct {
		name    string
		status  xaiv1.DocumentStatus
		errMsg  string
		wantErr string
	}{
		{"processed", xaiv1.DocumentStatus_DOCUMENT_STATUS_PROCESSED, "", ""},
		{"failed", xaiv1.DocumentStatus_DOCUMENT_STATUS_FAILED, "index boom", "index boom"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockColl{status: tt.status, errMsg: tt.errMsg}
			srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterCollectionsServer(s, mock) })
			if err != nil {
				t.Fatal(err)
			}
			defer srv.Close()

			cli := collections.New(srv.Conn, srv.Conn)
			doc, err := cli.WaitForIndexing(context.Background(), "c1", "f1", time.Second, time.Millisecond)
			if tt.wantErr != "" {
				if err == nil || err.Error() != tt.wantErr {
					t.Fatalf("err=%v want %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if doc.Status != xaiv1.DocumentStatus_DOCUMENT_STATUS_PROCESSED {
				t.Fatal(doc.Status)
			}
		})
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
	return &xaiv1.DocumentMetadata{FileMetadata: &xaiv1.FileMetadata{FileId: req.FileId}}, nil
}
