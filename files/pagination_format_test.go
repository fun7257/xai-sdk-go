package files_test

import (
	"context"
	"strings"
	"sync"
	"testing"

	"google.golang.org/grpc"

	"github.com/fun7257/xai-sdk-go/files"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// pagingFiles serves two pages of files and records list/content requests.
type pagingFiles struct {
	xaiv1.UnimplementedFilesServer
	mu         sync.Mutex
	listCalls  []*xaiv1.ListFilesRequest
	lastFormat *xaiv1.DownloadFormat
}

func (m *pagingFiles) ListFiles(ctx context.Context, req *xaiv1.ListFilesRequest) (*xaiv1.ListFilesResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.listCalls = append(m.listCalls, req)
	if req.GetPaginationToken() == "" {
		tok := "page2"
		return &xaiv1.ListFilesResponse{
			Data:            []*xaiv1.File{{Id: "f1"}, {Id: "f2"}},
			PaginationToken: &tok,
		}, nil
	}
	return &xaiv1.ListFilesResponse{Data: []*xaiv1.File{{Id: "f3"}}}, nil
}

func (m *pagingFiles) RetrieveFileContent(req *xaiv1.RetrieveFileContentRequest, stream xaiv1.Files_RetrieveFileContentServer) error {
	m.mu.Lock()
	m.lastFormat = req.Format
	m.mu.Unlock()
	return stream.Send(&xaiv1.FileContentChunk{Data: []byte("txt")})
}

func TestListAllFollowsPagination(t *testing.T) {
	mock := &pagingFiles{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := files.New(srv.Conn)

	all, err := cli.ListAll(context.Background(), files.WithListLimit(2))
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 || all[0].GetId() != "f1" || all[2].GetId() != "f3" {
		t.Fatalf("files=%v", all)
	}
	mock.mu.Lock()
	defer mock.mu.Unlock()
	if len(mock.listCalls) != 2 {
		t.Fatalf("list calls=%d want 2", len(mock.listCalls))
	}
	if mock.listCalls[0].Limit != 2 || mock.listCalls[1].Limit != 2 {
		t.Fatal("options must apply to every page")
	}
	if mock.listCalls[1].GetPaginationToken() != "page2" {
		t.Fatalf("second page token=%q", mock.listCalls[1].GetPaginationToken())
	}
}

func TestContentFormatOption(t *testing.T) {
	mock := &pagingFiles{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := files.New(srv.Conn)

	data, err := cli.Content(context.Background(), "f1", files.WithContentFormat("text"))
	if err != nil || string(data) != "txt" {
		t.Fatalf("Content: %v %q", err, data)
	}
	mock.mu.Lock()
	got := mock.lastFormat
	mock.mu.Unlock()
	if got == nil || *got != xaiv1.DownloadFormat_DOWNLOAD_FORMAT_TEXT {
		t.Fatalf("format=%v want TEXT", got)
	}

	if _, err := cli.Content(context.Background(), "f1", files.WithContentFormat("pdf")); err == nil {
		t.Fatal("expected unknown format rejection")
	}
	// Default (no option) sends no format.
	if _, err := cli.Content(context.Background(), "f1"); err != nil {
		t.Fatal(err)
	}
	mock.mu.Lock()
	got = mock.lastFormat
	mock.mu.Unlock()
	if got != nil {
		t.Fatalf("unset format must stay nil, got %v", got)
	}
}

func TestUploadFilenameValidation(t *testing.T) {
	mock := &mockFiles{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := files.New(srv.Conn)
	ctx := context.Background()

	bad := []string{
		"",                                // required
		"a\nb.txt",                        // line terminator
		"a\x00b.txt",                      // null byte
		`quote".txt`,                      // double quote
		"semi;colon.txt",                  // semicolon
		`back\slash.txt`,                  // backslash
		"nel\u0085.txt",                   // NEL
		"ls\u2028.txt",                    // line separator
		strings.Repeat("x", 256) + ".txt", // > 255 chars
	}
	for _, name := range bad {
		if _, err := cli.Upload(ctx, name, strings.NewReader("x"), 1); err == nil {
			t.Errorf("filename %q must be rejected", name)
		}
	}
	if n := mock.uploadCount(); n != 0 {
		t.Fatalf("uploads=%d want 0 (validation is local)", n)
	}
	if _, err := cli.Upload(ctx, "ok 文件-1.txt", strings.NewReader("x"), 1); err != nil {
		t.Fatalf("valid unicode filename rejected: %v", err)
	}
}
