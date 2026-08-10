package files_test

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

	"github.com/fun7257/xai-sdk-go/files"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

type mockFiles struct {
	xaiv1.UnimplementedFilesServer
	lastList   *xaiv1.ListFilesRequest
	lastGet    string
	lastDelete string
	lastPub    *xaiv1.CreatePublicUrlRequest
	lastRevoke string
	mu         sync.Mutex
	uploads    int
}

func (m *mockFiles) ListFiles(ctx context.Context, req *xaiv1.ListFilesRequest) (*xaiv1.ListFilesResponse, error) {
	m.lastList = req
	return &xaiv1.ListFilesResponse{}, nil
}

func (m *mockFiles) RetrieveFile(ctx context.Context, req *xaiv1.RetrieveFileRequest) (*xaiv1.File, error) {
	m.lastGet = req.GetFileId()
	return &xaiv1.File{Id: req.GetFileId(), Filename: "meta.txt"}, nil
}

func (m *mockFiles) DeleteFile(ctx context.Context, req *xaiv1.DeleteFileRequest) (*xaiv1.DeleteFileResponse, error) {
	m.lastDelete = req.GetFileId()
	return &xaiv1.DeleteFileResponse{Id: req.GetFileId(), Deleted: true}, nil
}

func (m *mockFiles) CreatePublicUrl(ctx context.Context, req *xaiv1.CreatePublicUrlRequest) (*xaiv1.CreatePublicUrlResponse, error) {
	m.lastPub = req
	return &xaiv1.CreatePublicUrlResponse{
		FileId:    req.GetFileId(),
		PublicUrl: "https://public.example/" + req.GetFileId(),
	}, nil
}

func (m *mockFiles) RevokePublicUrl(ctx context.Context, req *xaiv1.RevokePublicUrlRequest) (*xaiv1.RevokePublicUrlResponse, error) {
	m.lastRevoke = req.GetFileId()
	return &xaiv1.RevokePublicUrlResponse{FileId: req.GetFileId()}, nil
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
	zero := files.StorageOptions{}
	if zero.Proto() == nil {
		// zero value still produces a proto (empty filename)
		t.Fatal("expected non-nil proto for zero StorageOptions")
	}
	if files.StorageFromProto(nil) != nil {
		t.Fatal("StorageFromProto(nil)")
	}
	exp := time.Hour
	pubExp := 2 * time.Hour
	d := files.StorageOptions{
		Filename: "out.png", ExpiresAfter: &exp, PublicURL: true, PublicURLExpiresAfter: &pubExp,
	}
	p := d.Proto()
	if p.Filename != "out.png" || p.ExpiresAfter == nil || *p.ExpiresAfter != 3600 {
		t.Fatalf("%+v", p)
	}
	if p.PublicUrl == nil || p.PublicUrl.ExpiresAfter == nil || *p.PublicUrl.ExpiresAfter != 7200 {
		t.Fatalf("public_url=%+v", p.PublicUrl)
	}
	round := files.StorageFromProto(p)
	if round == nil || round.Filename != "out.png" || !round.PublicURL {
		t.Fatalf("%+v", round)
	}
	if round.ExpiresAfter == nil || *round.ExpiresAfter != time.Hour {
		t.Fatalf("expires=%v", round.ExpiresAfter)
	}
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

// TestGetDeletePublicURL covers Get/Delete/CreatePublicURL/RevokePublicURL via
// the shipped files.Client (sole domain coverage after root mega-test removal).
func TestGetDeletePublicURL(t *testing.T) {
	mock := &mockFiles{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := files.New(srv.Conn)
	ctx := context.Background()

	meta, err := cli.Get(ctx, "file_meta")
	if err != nil || meta.GetId() != "file_meta" || mock.lastGet != "file_meta" {
		t.Fatalf("Get: %v %#v last=%q", err, meta, mock.lastGet)
	}

	exp := time.Hour
	pub, err := cli.CreatePublicURL(ctx, "file_pub", &exp)
	if err != nil || pub.GetPublicUrl() == "" || mock.lastPub == nil {
		t.Fatalf("CreatePublicURL: %v %#v last=%v", err, pub, mock.lastPub)
	}
	if mock.lastPub.GetFileId() != "file_pub" || mock.lastPub.ExpiresAfter == nil || *mock.lastPub.ExpiresAfter != 3600 {
		t.Fatalf("CreatePublicURL wire=%+v", mock.lastPub)
	}

	rev, err := cli.RevokePublicURL(ctx, "file_pub")
	if err != nil || rev.GetFileId() != "file_pub" || mock.lastRevoke != "file_pub" {
		t.Fatalf("RevokePublicURL: %v %#v last=%q", err, rev, mock.lastRevoke)
	}

	del, err := cli.Delete(ctx, "file_del")
	if err != nil || !del.GetDeleted() || del.GetId() != "file_del" || mock.lastDelete != "file_del" {
		t.Fatalf("Delete: %v %#v last=%q", err, del, mock.lastDelete)
	}
}
