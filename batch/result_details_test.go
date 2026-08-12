package batch_test

import (
	"testing"

	spb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/fun7257/xai-sdk-go/batch"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// Result.Error must preserve structured status details, not just code+message.
func TestResultErrorPreservesDetails(t *testing.T) {
	detail, err := anypb.New(&xaiv1.File{Id: "ctx-file"})
	if err != nil {
		t.Fatal(err)
	}
	r := batch.NewResult(&xaiv1.BatchResult{
		BatchRequestId: "req-1",
		Result: &xaiv1.BatchResult_Error{Error: &spb.Status{
			Code:    int32(codes.InvalidArgument),
			Message: "bad request",
			Details: []*anypb.Any{detail},
		}},
	})
	if !r.Failed() {
		t.Fatal("expected failed result")
	}
	st := status.Convert(r.Error())
	if st.Code() != codes.InvalidArgument || st.Message() != "bad request" {
		t.Fatalf("status=%v", st)
	}
	details := st.Details()
	if len(details) != 1 {
		t.Fatalf("details=%d want 1", len(details))
	}
	f, ok := details[0].(*xaiv1.File)
	if !ok || f.GetId() != "ctx-file" {
		t.Fatalf("detail=%T %v", details[0], details[0])
	}
}
