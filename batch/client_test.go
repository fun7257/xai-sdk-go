package batch_test

import (
	"testing"

	spb "google.golang.org/genproto/googleapis/rpc/status"

	"github.com/fun7257/xai-sdk-go/batch"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

func TestResultSucceededFailed(t *testing.T) {
	ok := batch.NewResult(&xaiv1.BatchResult{
		BatchRequestId: "r1",
		Result: &xaiv1.BatchResult_Response{Response: &xaiv1.BatchResultData{
			Response: &xaiv1.BatchResultData_CompletionResponse{
				CompletionResponse: &xaiv1.GetChatCompletionResponse{
					Id: "c1",
					Outputs: []*xaiv1.CompletionOutput{{
						Message: &xaiv1.CompletionMessage{
							Role:    xaiv1.MessageRole_ROLE_ASSISTANT,
							Content: "hi",
						},
					}},
				},
			},
		}},
	})
	if !ok.Succeeded() || ok.Failed() {
		t.Fatal("expected succeeded")
	}
	cr, okChat := ok.ChatResponse()
	if !okChat || cr.Content() != "hi" {
		t.Fatalf("chat=%v ok=%v", cr, okChat)
	}

	fail := batch.NewResult(&xaiv1.BatchResult{
		BatchRequestId: "r2",
		Result:         &xaiv1.BatchResult_Error{Error: &spb.Status{Code: 3, Message: "bad"}},
	})
	if !fail.Failed() || fail.Succeeded() {
		t.Fatal("expected failed")
	}
	if fail.Error() == nil {
		t.Fatal("expected error")
	}

	img := batch.NewResult(&xaiv1.BatchResult{
		Result: &xaiv1.BatchResult_Response{Response: &xaiv1.BatchResultData{
			Response: &xaiv1.BatchResultData_ImageResponse{
				ImageResponse: &xaiv1.ImageResponse{
					Images: []*xaiv1.GeneratedImage{{
						Image: &xaiv1.GeneratedImage_Url{Url: "http://i"}, RespectModeration: true,
					}},
				},
			},
		}},
	})
	ir, okImg := img.ImageResponse()
	if !okImg {
		t.Fatal("image")
	}
	u, err := ir.URL()
	if err != nil || u != "http://i" {
		t.Fatalf("url=%q err=%v", u, err)
	}

	vid := batch.NewResult(&xaiv1.BatchResult{
		Result: &xaiv1.BatchResult_Response{Response: &xaiv1.BatchResultData{
			Response: &xaiv1.BatchResultData_VideoResponse{
				VideoResponse: &xaiv1.VideoResponse{
					Video: &xaiv1.GeneratedVideo{Url: "http://v", RespectModeration: true},
				},
			},
		}},
	})
	vr, okVid := vid.VideoResponse()
	if !okVid {
		t.Fatal("video")
	}
	vu, err := vr.URL()
	if err != nil || vu != "http://v" {
		t.Fatalf("url=%q err=%v", vu, err)
	}

	all := []*batch.Result{ok, fail, img}
	if len(batch.FilterSucceeded(all)) != 2 {
		t.Fatal(batch.FilterSucceeded(all))
	}
	if len(batch.FilterFailed(all)) != 1 {
		t.Fatal(batch.FilterFailed(all))
	}
}
