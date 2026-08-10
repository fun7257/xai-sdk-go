package files_test

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"google.golang.org/grpc"

	"github.com/fun7257/xai-sdk-go/files"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

type mockFiles struct {
	xaiv1.UnimplementedFilesServer
	lastList *xaiv1.ListFilesRequest
	mu       sync.Mutex
	uploads  int
}

func (m *mockFiles) ListFiles(ctx context.Context, req *xaiv1.ListFilesRequest) (*xaiv1.ListFilesResponse, error) {
	m.lastList = req
	return &xaiv1.ListFilesResponse{}, nil
}

func (m *mockFiles) UploadFile(stream xaiv1.Files_UploadFileServer) error {
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		_ = chunk
	}
	m.mu.Lock()
	m.uploads++
	m.mu.Unlock()
	return stream.SendAndClose(&xaiv1.File{Id: "file_1", Filename: "x.txt"})
}

func (m *mockFiles) uploadCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.uploads
}

func TestListFilterAndSort(t *testing.T) {
	mock := &mockFiles{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	cli := files.New(srv.Conn)
	_, err = cli.List(context.Background(),
		files.WithListFilter(`content_type = "application/pdf"`),
		files.WithListSortBy("filename"),
		files.WithListLimit(10),
		files.WithListOrder(false),
	)
	if err != nil {
		t.Fatal(err)
	}
	if mock.lastList == nil {
		t.Fatal("no list request")
	}
	if mock.lastList.GetFilter() != `content_type = "application/pdf"` {
		t.Fatalf("filter=%q", mock.lastList.GetFilter())
	}
	if mock.lastList.GetSortBy() != xaiv1.FilesSortBy_FILES_SORT_BY_FILENAME {
		t.Fatalf("sort_by=%v", mock.lastList.GetSortBy())
	}
	if mock.lastList.Limit != 10 {
		t.Fatalf("limit=%d", mock.lastList.Limit)
	}
	if mock.lastList.Order != xaiv1.Ordering_DESCENDING {
		t.Fatalf("order=%v", mock.lastList.Order)
	}
}

func TestBatchUploadCallback(t *testing.T) {
	mock := &mockFiles{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	dir := t.TempDir()
	p1 := filepath.Join(dir, "a.txt")
	p2 := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(p1, []byte("a"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p2, []byte("b"), 0o644); err != nil {
		t.Fatal(err)
	}

	cli := files.New(srv.Conn)
	var mu sync.Mutex
	var calls int
	filesOut, errs := cli.BatchUploadWithOptions(context.Background(), []string{p1, p2},
		files.WithBatchConcurrency(2),
		files.WithBatchOnComplete(func(index int, path string, file *xaiv1.File, err error) {
			mu.Lock()
			calls++
			mu.Unlock()
			if err != nil {
				t.Errorf("callback err index=%d: %v", index, err)
			}
			if file == nil || file.Id == "" {
				t.Errorf("callback missing file index=%d", index)
			}
		}),
	)
	if len(filesOut) != 2 || len(errs) != 2 {
		t.Fatalf("len files=%d errs=%d", len(filesOut), len(errs))
	}
	for i, e := range errs {
		if e != nil {
			t.Fatalf("errs[%d]=%v", i, e)
		}
	}
	if calls != 2 {
		t.Fatalf("callback calls=%d", calls)
	}
	if n := mock.uploadCount(); n != 2 {
		t.Fatalf("uploads=%d", n)
	}
}

func TestStorageOptionsProto(t *testing.T) {
	d := files.StorageOptions{Filename: "out.png", PublicURL: true}
	exp := int64(3600)
	// ExpiresAfter via duration
	sec := int64(3600)
	_ = sec
	p := d.Proto()
	if p.Filename != "out.png" || p.PublicUrl == nil {
		t.Fatalf("%+v", p)
	}
	_ = exp
}

func TestBatchUploadItemsReader(t *testing.T) {
	mock := &mockFiles{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := files.New(srv.Conn)
	items := []files.BatchItem{
		{Name: "a.txt", Reader: strings.NewReader("hello-a"), Size: 7},
		{Name: "b.txt", Reader: strings.NewReader("hello-b"), Size: 7},
	}
	filesOut, errs := cli.BatchUploadItems(context.Background(), items, files.WithBatchConcurrency(2))
	if len(filesOut) != 2 || len(errs) != 2 {
		t.Fatalf("len files=%d errs=%d", len(filesOut), len(errs))
	}
	for i, e := range errs {
		if e != nil {
			t.Fatalf("errs[%d]=%v", i, e)
		}
		if filesOut[i] == nil || filesOut[i].Id == "" {
			t.Fatalf("missing file at %d", i)
		}
	}
	if n := mock.uploadCount(); n != 2 {
		t.Fatalf("uploads=%d want 2", n)
	}
}

func (m *mockFiles) RetrieveFileContent(req *xaiv1.RetrieveFileContentRequest, stream xaiv1.Files_RetrieveFileContentServer) error {
	_ = stream.Send(&xaiv1.FileContentChunk{Data: []byte("hel")})
	_ = stream.Send(&xaiv1.FileContentChunk{Data: []byte("lo")})
	return nil
}

func TestContentWriter(t *testing.T) {
	mock := &mockFiles{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := files.New(srv.Conn)
	var buf strings.Builder
	if err := cli.ContentWriter(context.Background(), "f1", &buf); err != nil {
		t.Fatal(err)
	}
	if buf.String() != "hello" {
		t.Fatalf("got %q", buf.String())
	}
	b, err := cli.Content(context.Background(), "f1")
	if err != nil || string(b) != "hello" {
		t.Fatalf("Content: %v %q", err, b)
	}
}
