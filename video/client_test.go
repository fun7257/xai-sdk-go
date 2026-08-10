package video_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/fun7257/xai-sdk-go/files"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	"github.com/fun7257/xai-sdk-go/video"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

type mockVideo struct {
	xaiv1.UnimplementedVideoServer
	lastGen   *xaiv1.GenerateVideoRequest
	lastExt   *xaiv1.ExtendVideoRequest
	pollCount int
	failOnce  bool
	forceFail bool
	emptyDone bool
}

func (m *mockVideo) GenerateVideo(ctx context.Context, req *xaiv1.GenerateVideoRequest) (*xaiv1.StartDeferredResponse, error) {
	m.lastGen = req
	return &xaiv1.StartDeferredResponse{RequestId: "req_1"}, nil
}

func (m *mockVideo) ExtendVideo(ctx context.Context, req *xaiv1.ExtendVideoRequest) (*xaiv1.StartDeferredResponse, error) {
	m.lastExt = req
	return &xaiv1.StartDeferredResponse{RequestId: "req_ext"}, nil
}

func (m *mockVideo) GetDeferredVideo(ctx context.Context, req *xaiv1.GetDeferredVideoRequest) (*xaiv1.GetDeferredVideoResponse, error) {
	m.pollCount++
	if m.forceFail {
		return &xaiv1.GetDeferredVideoResponse{
			Status: xaiv1.DeferredStatus_FAILED,
			Response: &xaiv1.VideoResponse{
				Error: &xaiv1.VideoError{Code: "bad", Message: "boom"},
			},
		}, nil
	}
	if m.emptyDone {
		return &xaiv1.GetDeferredVideoResponse{Status: xaiv1.DeferredStatus_DONE}, nil
	}
	if m.failOnce && m.pollCount == 1 {
		return &xaiv1.GetDeferredVideoResponse{Status: xaiv1.DeferredStatus_PENDING}, nil
	}
	pub := "https://cdn.example/v.mp4"
	return &xaiv1.GetDeferredVideoResponse{
		Status: xaiv1.DeferredStatus_DONE,
		Response: &xaiv1.VideoResponse{
			Model: "grok-imagine-video",
			Video: &xaiv1.GeneratedVideo{
				Url: "https://vid.example/1.mp4", Duration: 1, RespectModeration: true,
				FileOutput: &xaiv1.FileOutput{FileId: "file_v", PublicUrl: &pub},
			},
		},
	}, nil
}

func TestGenerateFileIDStorageAndPoll(t *testing.T) {
	mock := &mockVideo{failOnce: true}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterVideoServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	cli := video.New(srv.Conn)
	exp := time.Hour
	resp, err := cli.Generate(context.Background(), "a cat runs", "grok-imagine-video",
		video.WithImageFileID("file_img"),
		video.WithReferenceImageFileIDs("file_ref1"),
		video.WithStorage(files.StorageOptions{Filename: "v.mp4", ExpiresAfter: &exp}),
		video.WithPollInterval(time.Millisecond),
		video.WithPollTimeout(time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}
	if mock.lastGen.GetImage().GetFileId() != "file_img" {
		t.Fatalf("image file_id=%q", mock.lastGen.GetImage().GetFileId())
	}
	if len(mock.lastGen.ReferenceImages) != 1 || mock.lastGen.ReferenceImages[0].GetFileId() != "file_ref1" {
		t.Fatalf("refs=%v", mock.lastGen.ReferenceImages)
	}
	if mock.lastGen.StorageOptions == nil || mock.lastGen.StorageOptions.Filename != "v.mp4" {
		t.Fatalf("storage=%v", mock.lastGen.StorageOptions)
	}
	if mock.pollCount < 2 {
		t.Fatalf("pollCount=%d", mock.pollCount)
	}
	u, err := resp.URL()
	if err != nil || u == "" {
		t.Fatalf("url=%q err=%v", u, err)
	}
	if resp.FileOutput() == nil || resp.PublicURL() == "" {
		t.Fatalf("storage fields missing")
	}
}

func TestGenerateDoneEmptyResponse(t *testing.T) {
	mock := &mockVideo{emptyDone: true}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterVideoServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := video.New(srv.Conn)
	_, err = cli.Generate(context.Background(), "x", "m",
		video.WithPollInterval(time.Millisecond),
		video.WithPollTimeout(time.Second),
	)
	if err == nil || err.Error() != "video generation completed with empty response" {
		t.Fatalf("err=%v", err)
	}
}

func TestUnknownAspectRatio(t *testing.T) {
	cli := video.New(nil)
	_, err := cli.Prepare("p", "m", video.WithAspectRatio("not-a-ratio"))
	if err == nil {
		t.Fatal("expected aspect error")
	}
}

func TestGenerationFailed(t *testing.T) {
	mock := &mockVideo{forceFail: true}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterVideoServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := video.New(srv.Conn)
	_, err = cli.Generate(context.Background(), "x", "m",
		video.WithPollInterval(time.Millisecond),
		video.WithPollTimeout(time.Second),
	)
	var ge *video.GenerationError
	if !errors.As(err, &ge) {
		t.Fatalf("want GenerationError got %v", err)
	}
	if ge.Code != "bad" || ge.Message != "boom" {
		t.Fatalf("%+v", ge)
	}
}

func TestMutualExclusionAndExtend(t *testing.T) {
	cli := video.New(nil)
	_, err := cli.Prepare("p", "m", video.WithImageURL("http://a"), video.WithImageFileID("f"))
	if err == nil {
		t.Fatal("expected image url/file mutex")
	}
	_, err = cli.Prepare("p", "m", video.WithVideoURL("http://v"), video.WithVideoFileID("vf"))
	if err == nil {
		t.Fatal("expected video url/file mutex")
	}
	_, err = cli.PrepareExtension("p", "m", "")
	if err == nil {
		t.Fatal("expected extend requires video")
	}

	// Zero Response defaults RespectModeration to true (no video payload).
	if !(&video.Response{}).RespectModeration() {
		t.Fatal("expected default RespectModeration true on empty Response")
	}

	mock := &mockVideo{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterVideoServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli = video.New(srv.Conn)

	// ExtendStart returns request id without polling.
	id, err := cli.ExtendStart(context.Background(), "continue", "m", "https://v.example/in.mp4")
	if err != nil {
		t.Fatal(err)
	}
	if id != "req_ext" {
		t.Fatalf("id=%q", id)
	}
	if mock.lastExt == nil || mock.lastExt.GetVideo().GetUrl() != "https://v.example/in.mp4" {
		t.Fatalf("extend start wire=%v", mock.lastExt)
	}

	_, err = cli.ExtendWith(context.Background(), "continue", "m", "", "file_vid",
		video.WithPollInterval(time.Millisecond),
		video.WithPollTimeout(time.Second),
		video.WithStorage(files.StorageOptions{Filename: "ext.mp4"}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if mock.lastExt.GetVideo().GetFileId() != "file_vid" {
		t.Fatalf("extend file_id=%q", mock.lastExt.GetVideo().GetFileId())
	}
	if mock.lastExt.StorageOptions == nil {
		t.Fatal("expected storage on extend")
	}
}
