package collections_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/fun7257/xai-sdk-go/collections"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

type hardeningMock struct {
	xaiv1.UnimplementedFilesServer
	xaiv1.UnimplementedCollectionsServer

	mu         sync.Mutex
	lastCreate *xaiv1.CreateCollectionRequest
	deleted    []string
	blockAdd   bool
}

func (m *hardeningMock) CreateCollection(ctx context.Context, req *xaiv1.CreateCollectionRequest) (*xaiv1.CollectionMetadata, error) {
	m.mu.Lock()
	m.lastCreate = req
	m.mu.Unlock()
	return &xaiv1.CollectionMetadata{CollectionId: "col1"}, nil
}

func (m *hardeningMock) UploadFile(stream xaiv1.Files_UploadFileServer) error {
	for {
		if _, err := stream.Recv(); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}
	return stream.SendAndClose(&xaiv1.File{Id: "file_up"})
}

// AddDocumentToCollection blocks until the caller's deadline expires so the
// add fails with a context error.
func (m *hardeningMock) AddDocumentToCollection(ctx context.Context, req *xaiv1.AddDocumentToCollectionRequest) (*emptypb.Empty, error) {
	m.mu.Lock()
	block := m.blockAdd
	m.mu.Unlock()
	if block {
		<-ctx.Done()
		return nil, status.FromContextError(ctx.Err()).Err()
	}
	return &emptypb.Empty{}, nil
}

func (m *hardeningMock) DeleteFile(ctx context.Context, req *xaiv1.DeleteFileRequest) (*xaiv1.DeleteFileResponse, error) {
	m.mu.Lock()
	m.deleted = append(m.deleted, req.GetFileId())
	m.mu.Unlock()
	return &xaiv1.DeleteFileResponse{Id: req.GetFileId(), Deleted: true}, nil
}

func (m *hardeningMock) deletedFiles() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string(nil), m.deleted...)
}

// Unknown metrics must fail Create locally instead of silently using cosine.
func TestCreateRejectsUnknownMetric(t *testing.T) {
	mock := &hardeningMock{}
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterCollectionsServer(s, mock)
		xaiv1.RegisterFilesServer(s, mock)
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := collections.New(srv.Conn, srv.Conn)

	_, err = cli.Create(context.Background(), "c", collections.WithMetric("euclidian")) // typo
	if err == nil || !strings.Contains(err.Error(), "unknown metric") {
		t.Fatalf("expected unknown metric error, got: %v", err)
	}
	if _, err := cli.Create(context.Background(), "c", collections.WithMetric("euclidean")); err != nil {
		t.Fatal(err)
	}
	mock.mu.Lock()
	got := mock.lastCreate.GetMetricSpace()
	mock.mu.Unlock()
	if got != xaiv1.HNSWMetric_HNSW_METRIC_EUCLIDEAN {
		t.Fatalf("metric=%v", got)
	}
}

// WithDeleteOnAddFailure must clean up the orphan file even when the add
// failed because the caller's context expired.
func TestDeleteOnAddFailureSurvivesContextExpiry(t *testing.T) {
	mock := &hardeningMock{blockAdd: true}
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterCollectionsServer(s, mock)
		xaiv1.RegisterFilesServer(s, mock)
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
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	f, _, err := cli.UploadDocument(ctx, "col1", path, collections.WithDeleteOnAddFailure())
	if err == nil {
		t.Fatal("expected add failure from expired context")
	}
	if f == nil || f.Id != "file_up" {
		t.Fatalf("file=%v", f)
	}
	deleted := mock.deletedFiles()
	if len(deleted) != 1 || deleted[0] != "file_up" {
		t.Fatalf("orphan cleanup missing, deleted=%v", deleted)
	}
}
