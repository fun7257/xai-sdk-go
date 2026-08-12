package files_test

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fun7257/xai-sdk-go/files"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// rejectingFiles aborts every upload with RESOURCE_EXHAUSTED after the first
// message, simulating a quota / size rejection raised mid-stream.
type rejectingFiles struct {
	xaiv1.UnimplementedFilesServer
}

func (m *rejectingFiles) UploadFile(stream xaiv1.Files_UploadFileServer) error {
	_, _ = stream.Recv() // read init, then abort
	return status.Error(codes.ResourceExhausted, "file exceeds plan quota")
}

// Upload must surface the server's status instead of a bare io.EOF when the
// server terminates the stream while the client is still sending.
func TestUploadServerRejectionSurfacesStatus(t *testing.T) {
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, &rejectingFiles{}) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	cli := files.New(srv.Conn)
	// 8 MiB payload spans several 3 MiB chunks so Send happens after the abort.
	payload := bytes.Repeat([]byte("x"), 8<<20)
	_, err = cli.Upload(context.Background(), "big.bin", bytes.NewReader(payload), int64(len(payload)))
	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.ResourceExhausted {
		t.Fatalf("expected RESOURCE_EXHAUSTED status, got: %v", err)
	}
	if !strings.Contains(st.Message(), "quota") {
		t.Fatalf("server message lost: %v", err)
	}
}

// failingReader errors after the first read.
type failingReader struct{ reads int }

func (r *failingReader) Read(p []byte) (int, error) {
	r.reads++
	if r.reads == 1 {
		return copy(p, []byte("partial")), nil
	}
	return 0, context.DeadlineExceeded
}

func TestUploadReaderErrorPropagates(t *testing.T) {
	mock := &mockFiles{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	cli := files.New(srv.Conn)
	_, err = cli.Upload(context.Background(), "r.bin", &failingReader{}, -1)
	if err == nil || !strings.Contains(err.Error(), "read upload source") {
		t.Fatalf("expected wrapped reader error, got: %v", err)
	}
}

func TestUploadExpiresAfterValidation(t *testing.T) {
	mock := &mockFiles{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := files.New(srv.Conn)

	for _, d := range []time.Duration{0, -time.Hour, 500 * time.Millisecond, time.Minute, 31 * 24 * time.Hour} {
		_, err := cli.Upload(context.Background(), "x.txt", strings.NewReader("x"), 1, files.WithExpiresAfter(d))
		if err == nil {
			t.Fatalf("expected out-of-range error for %s", d)
		}
	}
	if _, err := cli.Upload(context.Background(), "x.txt", strings.NewReader("x"), 1, files.WithExpiresAfter(time.Hour)); err != nil {
		t.Fatalf("1h TTL should be accepted: %v", err)
	}
	if n := mock.uploadCount(); n != 1 {
		t.Fatalf("uploads=%d want 1 (invalid TTLs must fail before streaming)", n)
	}
}

func TestCreatePublicURLExpiresValidation(t *testing.T) {
	mock := &mockFiles{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := files.New(srv.Conn)

	bad := 500 * time.Millisecond
	if _, err := cli.CreatePublicURL(context.Background(), "f1", &bad); err == nil {
		t.Fatal("expected sub-second TTL rejection")
	}
	ok := time.Hour
	if _, err := cli.CreatePublicURL(context.Background(), "f1", &ok); err != nil {
		t.Fatalf("1h TTL should be accepted: %v", err)
	}
}

// errWriter fails on the first write.
type errWriter struct{}

func (errWriter) Write([]byte) (int, error) { return 0, context.DeadlineExceeded }

func TestContentWriterWriteErrorPropagates(t *testing.T) {
	mock := &mockFiles{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := files.New(srv.Conn)
	if err := cli.ContentWriter(context.Background(), "f1", errWriter{}); err == nil {
		t.Fatal("expected writer error to propagate")
	}
}

func TestBatchUploadItemsPathAndReaderMutuallyExclusive(t *testing.T) {
	mock := &mockFiles{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := files.New(srv.Conn)

	items := []files.BatchItem{{Path: "/tmp/x.txt", Reader: strings.NewReader("x"), Name: "x.txt", Size: 1}}
	_, errs := cli.BatchUploadItems(context.Background(), items)
	if errs[0] == nil || !strings.Contains(errs[0].Error(), "exactly one of Path or Reader") {
		t.Fatalf("expected mutual exclusion error, got: %v", errs[0])
	}
	if n := mock.uploadCount(); n != 0 {
		t.Fatalf("uploads=%d want 0", n)
	}
}

func TestBatchUploadCallbackPanicIsRecovered(t *testing.T) {
	mock := &mockFiles{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := files.New(srv.Conn)

	var mu sync.Mutex
	var calls int
	items := []files.BatchItem{
		{Name: "a.txt", Reader: strings.NewReader("a"), Size: 1},
		{Name: "b.txt", Reader: strings.NewReader("b"), Size: 1},
	}
	filesOut, errs := cli.BatchUploadItems(context.Background(), items,
		files.WithBatchConcurrency(1),
		files.WithBatchOnComplete(func(index int, path string, file *xaiv1.File, err error) {
			mu.Lock()
			calls++
			mu.Unlock()
			if index == 0 {
				panic("callback boom")
			}
		}),
	)
	if calls != 2 {
		t.Fatalf("callback calls=%d want 2 (panic must not stop the batch)", calls)
	}
	if errs[0] == nil || !strings.Contains(errs[0].Error(), "callback panicked") {
		t.Fatalf("expected recovered callback panic in errs[0], got: %v", errs[0])
	}
	if filesOut[0] == nil {
		t.Fatal("upload succeeded before callback panicked; file must still be returned")
	}
	if errs[1] != nil || filesOut[1] == nil {
		t.Fatalf("second item must be unaffected: %v %v", filesOut[1], errs[1])
	}
}
