package collections_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/fun7257/xai-sdk-go/collections"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// mgmtMock records the last request of every management RPC not covered elsewhere.
type mgmtMock struct {
	xaiv1.UnimplementedCollectionsServer

	mu          sync.Mutex
	deletedID   string
	lastUpdate  *xaiv1.UpdateCollectionRequest
	lastRemove  *xaiv1.RemoveDocumentFromCollectionRequest
	lastList    *xaiv1.ListDocumentsRequest
	lastBatch   *xaiv1.BatchGetDocumentsRequest
	lastReindex *xaiv1.ReIndexDocumentRequest
	lastAdd     *xaiv1.AddDocumentToCollectionRequest
}

func (m *mgmtMock) DeleteCollection(ctx context.Context, req *xaiv1.DeleteCollectionRequest) (*emptypb.Empty, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.deletedID = req.GetCollectionId()
	return &emptypb.Empty{}, nil
}

func (m *mgmtMock) UpdateCollection(ctx context.Context, req *xaiv1.UpdateCollectionRequest) (*xaiv1.CollectionMetadata, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastUpdate = req
	return &xaiv1.CollectionMetadata{CollectionId: req.GetCollectionId(), CollectionName: req.GetCollectionName()}, nil
}

func (m *mgmtMock) RemoveDocumentFromCollection(ctx context.Context, req *xaiv1.RemoveDocumentFromCollectionRequest) (*emptypb.Empty, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastRemove = req
	return &emptypb.Empty{}, nil
}

func (m *mgmtMock) ListDocuments(ctx context.Context, req *xaiv1.ListDocumentsRequest) (*xaiv1.ListDocumentsResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastList = req
	return &xaiv1.ListDocumentsResponse{
		Documents: []*xaiv1.DocumentMetadata{{FileMetadata: &xaiv1.FileMetadata{FileId: "f1"}}},
	}, nil
}

func (m *mgmtMock) BatchGetDocuments(ctx context.Context, req *xaiv1.BatchGetDocumentsRequest) (*xaiv1.BatchGetDocumentsResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastBatch = req
	return &xaiv1.BatchGetDocumentsResponse{
		Documents: []*xaiv1.DocumentMetadata{
			{FileMetadata: &xaiv1.FileMetadata{FileId: "f1"}},
			{FileMetadata: &xaiv1.FileMetadata{FileId: "f2"}},
		},
	}, nil
}

func (m *mgmtMock) ReIndexDocument(ctx context.Context, req *xaiv1.ReIndexDocumentRequest) (*emptypb.Empty, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastReindex = req
	return &emptypb.Empty{}, nil
}

func (m *mgmtMock) AddDocumentToCollection(ctx context.Context, req *xaiv1.AddDocumentToCollectionRequest) (*emptypb.Empty, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastAdd = req
	return &emptypb.Empty{}, nil
}

func TestManagementDocumentLifecycle(t *testing.T) {
	mock := &mgmtMock{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterCollectionsServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := collections.New(srv.Conn, srv.Conn)
	ctx := context.Background()

	if err := cli.AddExistingDocument(ctx, "c1", "f1", map[string]string{"k": "v"}); err != nil {
		t.Fatalf("AddExistingDocument: %v", err)
	}
	if mock.lastAdd.GetCollectionId() != "c1" || mock.lastAdd.GetFileId() != "f1" || mock.lastAdd.GetFields()["k"] != "v" {
		t.Fatalf("add wire=%+v", mock.lastAdd)
	}

	docs, err := cli.ListDocuments(ctx, "c1",
		collections.WithDocumentsFilter(`content_type = "text/plain"`),
		collections.WithDocumentsName("doc"),
		collections.WithDocumentsLimit(7),
	)
	if err != nil || len(docs.GetDocuments()) != 1 {
		t.Fatalf("ListDocuments: %v %+v", err, docs)
	}
	if mock.lastList.GetCollectionId() != "c1" || mock.lastList.GetLimit() != 7 ||
		mock.lastList.GetName() != "doc" || !strings.Contains(mock.lastList.GetFilter(), "text/plain") {
		t.Fatalf("list wire=%+v", mock.lastList)
	}

	batch, err := cli.BatchGetDocuments(ctx, "c1", []string{"f1", "f2"})
	if err != nil || len(batch.GetDocuments()) != 2 {
		t.Fatalf("BatchGetDocuments: %v %+v", err, batch)
	}
	if got := mock.lastBatch.GetFileIds(); len(got) != 2 || got[0] != "f1" || got[1] != "f2" {
		t.Fatalf("batch wire=%+v", mock.lastBatch)
	}

	if err := cli.ReindexDocument(ctx, "c1", "f1"); err != nil {
		t.Fatalf("ReindexDocument: %v", err)
	}
	if mock.lastReindex.GetCollectionId() != "c1" || mock.lastReindex.GetFileId() != "f1" {
		t.Fatalf("reindex wire=%+v", mock.lastReindex)
	}

	if err := cli.RemoveDocument(ctx, "c1", "f1"); err != nil {
		t.Fatalf("RemoveDocument: %v", err)
	}
	if mock.lastRemove.GetCollectionId() != "c1" || mock.lastRemove.GetFileId() != "f1" {
		t.Fatalf("remove wire=%+v", mock.lastRemove)
	}

	if err := cli.Delete(ctx, "c1"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if mock.deletedID != "c1" {
		t.Fatalf("deleted=%q", mock.deletedID)
	}
}

func TestUpdateCollection(t *testing.T) {
	mock := &mgmtMock{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterCollectionsServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := collections.New(srv.Conn, srv.Conn)
	ctx := context.Background()

	// nil request: still sends with the collection id filled in.
	meta, err := cli.Update(ctx, "c1", nil)
	if err != nil || meta.GetCollectionId() != "c1" || mock.lastUpdate.GetCollectionId() != "c1" {
		t.Fatalf("Update(nil): %v %+v wire=%+v", err, meta, mock.lastUpdate)
	}

	if _, err = cli.Update(ctx, "c1", &xaiv1.UpdateCollectionRequest{CollectionName: "renamed"}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if mock.lastUpdate.GetCollectionName() != "renamed" {
		t.Fatalf("update wire=%+v", mock.lastUpdate)
	}

	// Invalid chunk configuration is rejected locally.
	_, err = cli.Update(ctx, "c1", &xaiv1.UpdateCollectionRequest{
		ChunkConfiguration: &xaiv1.ChunkConfiguration{},
	})
	if err == nil || !strings.Contains(err.Error(), "chunk configuration") {
		t.Fatalf("expected chunk validation error, got: %v", err)
	}

	// The caller's request must not be mutated (id is set on a clone).
	callerReq := &xaiv1.UpdateCollectionRequest{CollectionName: "keep"}
	if _, err = cli.Update(ctx, "c9", callerReq); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if callerReq.GetCollectionId() != "" {
		t.Fatalf("caller request mutated: %+v", callerReq)
	}
	if mock.lastUpdate.GetCollectionId() != "c9" || mock.lastUpdate.GetCollectionName() != "keep" {
		t.Fatalf("clone wire=%+v", mock.lastUpdate)
	}
}

// Every management method must fail with ErrNoManagementKey when the client
// was built without a management connection; Search (business channel) must
// still work.
func TestManagementMethodsRequireKey(t *testing.T) {
	docs := &mockDocs{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterDocumentsServer(s, docs) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := collections.New(srv.Conn, nil) // no management channel
	ctx := context.Background()

	calls := map[string]func() error{
		"Create":              func() error { _, err := cli.Create(ctx, "c"); return err },
		"List":                func() error { _, err := cli.List(ctx); return err },
		"Get":                 func() error { _, err := cli.Get(ctx, "c1"); return err },
		"Delete":              func() error { return cli.Delete(ctx, "c1") },
		"Update":              func() error { _, err := cli.Update(ctx, "c1", nil); return err },
		"GenerateDescription": func() error { _, err := cli.GenerateDescription(ctx, "c1"); return err },
		"AddExistingDocument": func() error { return cli.AddExistingDocument(ctx, "c1", "f1", nil) },
		"UploadDocument":      func() error { _, _, err := cli.UploadDocument(ctx, "c1", "/nope"); return err },
		"RemoveDocument":      func() error { return cli.RemoveDocument(ctx, "c1", "f1") },
		"UpdateDocument":      func() error { _, err := cli.UpdateDocument(ctx, "c1", "f1"); return err },
		"GetDocument":         func() error { _, err := cli.GetDocument(ctx, "c1", "f1"); return err },
		"ListDocuments":       func() error { _, err := cli.ListDocuments(ctx, "c1"); return err },
		"BatchGetDocuments":   func() error { _, err := cli.BatchGetDocuments(ctx, "c1", nil); return err },
		"ReindexDocument":     func() error { return cli.ReindexDocument(ctx, "c1", "f1") },
	}
	for name, call := range calls {
		if err := call(); !errors.Is(err, collections.ErrNoManagementKey) {
			t.Errorf("%s: want ErrNoManagementKey, got %v", name, err)
		}
	}

	if _, err := cli.Search(ctx, "q", []string{"c1"}); err != nil {
		t.Fatalf("Search must work without management key: %v", err)
	}
}
