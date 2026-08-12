package collections_test

import (
	"context"
	"testing"

	"google.golang.org/grpc"

	"github.com/fun7257/xai-sdk-go/collections"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

func TestValidateChunkConfiguration(t *testing.T) {
	if err := collections.ValidateChunkConfiguration(nil); err != nil {
		t.Fatal(err)
	}
	// The strategies are a proto oneof (at most one set); the zero value has
	// none and must be rejected.
	if err := collections.ValidateChunkConfiguration(&xaiv1.ChunkConfiguration{}); err == nil {
		t.Fatal("empty should error")
	}
	if err := collections.ValidateChunkConfiguration(&xaiv1.ChunkConfiguration{
		Config: &xaiv1.ChunkConfiguration_CharsConfiguration{
			CharsConfiguration: &xaiv1.CharsConfiguration{MaxChunkSizeChars: 100},
		},
	}); err != nil {
		t.Fatal(err)
	}
}

func TestFieldDefinitionHelpers(t *testing.T) {
	fd := collections.FieldDefinition("author", true, true, false, "doc author")
	if fd.Key != "author" || !fd.Required {
		t.Fatal(fd)
	}
	add := collections.AddFieldDefinition(fd)
	if add.Operation != xaiv1.FieldDefinitionOperation_FIELD_DEFINITION_ADD {
		t.Fatal(add.Operation)
	}
	del := collections.DeleteFieldDefinition("author")
	if del.Operation != xaiv1.FieldDefinitionOperation_FIELD_DEFINITION_DELETE {
		t.Fatal(del.Operation)
	}
}

type searchMock struct {
	xaiv1.UnimplementedDocumentsServer
	last *xaiv1.SearchRequest
}

func (m *searchMock) Search(ctx context.Context, req *xaiv1.SearchRequest) (*xaiv1.SearchResponse, error) {
	m.last = req
	return &xaiv1.SearchResponse{}, nil
}

func TestSearchRetrievalMode(t *testing.T) {
	mock := &searchMock{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterDocumentsServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := collections.New(srv.Conn, srv.Conn)
	_, err = cli.Search(context.Background(), "q", []string{"c1"},
		collections.WithSearchLimit(5),
		collections.WithSearchInstructions("prefer docs"),
		collections.WithRetrievalMode("semantic"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if mock.last.GetSemanticRetrieval() == nil {
		t.Fatal("expected semantic")
	}
	if mock.last.GetInstructions() != "prefer docs" {
		t.Fatalf("instructions=%q", mock.last.GetInstructions())
	}
	if mock.last.GetLimit() != 5 {
		t.Fatalf("limit=%d", mock.last.GetLimit())
	}
}
