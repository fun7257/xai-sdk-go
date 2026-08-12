package files_test

import (
	"context"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/fun7257/xai-sdk-go/files"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// erroringContentFiles streams one chunk, then fails.
type erroringContentFiles struct {
	xaiv1.UnimplementedFilesServer
}

func (m *erroringContentFiles) RetrieveFileContent(req *xaiv1.RetrieveFileContentRequest, stream xaiv1.Files_RetrieveFileContentServer) error {
	_ = stream.Send(&xaiv1.FileContentChunk{Data: []byte("partial")})
	return status.Error(codes.Internal, "storage hiccup")
}

// Content must not hand back partial data alongside an error.
func TestContentReturnsNoPartialDataOnError(t *testing.T) {
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterFilesServer(s, &erroringContentFiles{}) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := files.New(srv.Conn)

	data, err := cli.Content(context.Background(), "f1")
	if err == nil {
		t.Fatal("expected stream error")
	}
	if data != nil {
		t.Fatalf("expected nil data on error, got %d bytes", len(data))
	}
}
