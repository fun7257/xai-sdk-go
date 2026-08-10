package collections_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/fun7257/xai-sdk-go/collections"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

type uploadE2E struct {
	xaiv1.UnimplementedFilesServer
	xaiv1.UnimplementedCollectionsServer

	mu        sync.Mutex
	steps     []string
	failAdd   bool
	docStatus xaiv1.DocumentStatus
}

func (m *uploadE2E) UploadFile(stream xaiv1.Files_UploadFileServer) error {
	m.mu.Lock()
	m.steps = append(m.steps, "upload")
	m.mu.Unlock()
	for {
		_, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return stream.SendAndClose(&xaiv1.File{Id: "file_up", Filename: "doc.txt"})
}

func (m *uploadE2E) AddDocumentToCollection(ctx context.Context, req *xaiv1.AddDocumentToCollectionRequest) (*emptypb.Empty, error) {
	m.mu.Lock()
	m.steps = append(m.steps, "add:"+req.FileId)
	fail := m.failAdd
	m.mu.Unlock()
	if fail {
		return nil, io.ErrUnexpectedEOF
	}
	return &emptypb.Empty{}, nil
}

func (m *uploadE2E) GetDocumentMetadata(ctx context.Context, req *xaiv1.GetDocumentMetadataRequest) (*xaiv1.DocumentMetadata, error) {
	m.mu.Lock()
	m.steps = append(m.steps, "get:"+req.FileId)
	st := m.docStatus
	m.mu.Unlock()
	return &xaiv1.DocumentMetadata{
		Status:       st,
		FileMetadata: &xaiv1.FileMetadata{FileId: req.FileId},
	}, nil
}

func TestUploadDocumentE2E(t *testing.T) {
	mock := &uploadE2E{docStatus: xaiv1.DocumentStatus_DOCUMENT_STATUS_PROCESSED}
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterFilesServer(s, mock)
		xaiv1.RegisterCollectionsServer(s, mock)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	dir := t.TempDir()
	path := filepath.Join(dir, "doc.txt")
	if err := os.WriteFile(path, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	cli := collections.New(srv.Conn, srv.Conn)
	f, doc, err := cli.UploadDocument(context.Background(), "col1", path,
		collections.WithDocumentFields(map[string]string{"k": "v"}),
		collections.WithWaitForIndexing(time.Second, time.Millisecond),
	)
	if err != nil {
		t.Fatal(err)
	}
	if f == nil || f.Id != "file_up" {
		t.Fatalf("file=%v", f)
	}
	if doc == nil || doc.Status != xaiv1.DocumentStatus_DOCUMENT_STATUS_PROCESSED {
		t.Fatalf("doc=%v", doc)
	}
	mock.mu.Lock()
	steps := append([]string(nil), mock.steps...)
	mock.mu.Unlock()
	if len(steps) < 3 || steps[0] != "upload" || steps[1] != "add:file_up" {
		t.Fatalf("steps=%v", steps)
	}
}

func TestUploadDocumentAddErrorReturnsFile(t *testing.T) {
	mock := &uploadE2E{failAdd: true}
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterFilesServer(s, mock)
		xaiv1.RegisterCollectionsServer(s, mock)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	dir := t.TempDir()
	path := filepath.Join(dir, "doc.txt")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	cli := collections.New(srv.Conn, srv.Conn)
	f, doc, err := cli.UploadDocument(context.Background(), "col1", path)
	if err == nil {
		t.Fatal("expected add error")
	}
	if f == nil || f.Id != "file_up" {
		t.Fatalf("expected uploaded file returned for cleanup, got %v", f)
	}
	if doc != nil {
		t.Fatalf("doc should be nil, got %v", doc)
	}
}
