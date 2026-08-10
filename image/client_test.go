package image_test

import (
	"context"
	"strconv"
	"testing"
	"time"

	"google.golang.org/grpc"

	"github.com/fun7257/xai-sdk-go/files"
	"github.com/fun7257/xai-sdk-go/image"
	"github.com/fun7257/xai-sdk-go/internal/testutil"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

type captureImage struct {
	xaiv1.UnimplementedImageServer
	last *xaiv1.GenerateImageRequest
}

func (m *captureImage) GenerateImage(ctx context.Context, req *xaiv1.GenerateImageRequest) (*xaiv1.ImageResponse, error) {
	m.last = req
	pub := "https://cdn.example/public.png"
	return &xaiv1.ImageResponse{
		Model: req.Model,
		Images: []*xaiv1.GeneratedImage{{
			RespectModeration: true,
			Image:             &xaiv1.GeneratedImage_Url{Url: "https://img.example/1.png"},
			FileOutput: &xaiv1.FileOutput{
				FileId: "file_stored", Filename: "out.png", PublicUrl: &pub,
			},
		}},
	}, nil
}

func TestDefaultFormatURL(t *testing.T) {
	mock := &captureImage{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterImageServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	cli := image.New(srv.Conn)
	if _, err := cli.Sample(context.Background(), "cat", "grok-imagine-image"); err != nil {
		t.Fatal(err)
	}
	if mock.last == nil || mock.last.Format != xaiv1.ImageFormat_IMG_FORMAT_URL {
		t.Fatalf("Sample default format=%v want URL", mock.last.Format)
	}

	req, err := cli.Prepare("dog", "grok-imagine-image")
	if err != nil {
		t.Fatal(err)
	}
	if req.Format != xaiv1.ImageFormat_IMG_FORMAT_URL {
		t.Fatalf("Prepare default format=%v want URL", req.Format)
	}

	if _, err := cli.Sample(context.Background(), "x", "m", image.WithFormatBase64()); err != nil {
		t.Fatal(err)
	}
	if mock.last.Format != xaiv1.ImageFormat_IMG_FORMAT_BASE64 {
		t.Fatalf("got %v", mock.last.Format)
	}
}

func TestFileIDAndStorageRequest(t *testing.T) {
	mock := &captureImage{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterImageServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()

	cli := image.New(srv.Conn)
	exp := time.Hour
	resp, err := cli.Sample(context.Background(), "edit", "grok-imagine-image",
		image.WithImageFileID("file_abc"),
		image.WithStorage(files.StorageOptions{Filename: "out.png", ExpiresAfter: &exp, PublicURL: true}),
	)
	if err != nil {
		t.Fatal(err)
	}
	if mock.last.GetImage().GetFileId() != "file_abc" {
		t.Fatalf("file_id=%q", mock.last.GetImage().GetFileId())
	}
	if mock.last.StorageOptions == nil || mock.last.StorageOptions.Filename != "out.png" {
		t.Fatalf("storage=%v", mock.last.StorageOptions)
	}
	if mock.last.StorageOptions.PublicUrl == nil {
		t.Fatal("expected public_url options")
	}
	if resp.FileOutput() == nil || resp.FileOutput().FileId != "file_stored" {
		t.Fatalf("file_output=%v", resp.FileOutput())
	}
	if resp.PublicURL() != "https://cdn.example/public.png" {
		t.Fatalf("public_url=%q", resp.PublicURL())
	}
}

func TestMutualExclusion(t *testing.T) {
	cli := image.New(nil)
	_, err := cli.Prepare("p", "m", image.WithImageURL("http://a"), image.WithImageFileID("f1"))
	if err == nil {
		t.Fatal("expected mutex error url+file_id")
	}
	_, err = cli.Prepare("p", "m", image.WithImageURL("http://a"), image.WithImageURLs("http://b"))
	if err == nil {
		t.Fatal("expected mutex error single+multi")
	}
	_, err = cli.Prepare("p", "m", image.WithImageFileID("f1"), image.WithImageFileIDs("f2"))
	if err == nil {
		t.Fatal("expected mutex error single file + multi files")
	}
}

func TestUnknownAspectRatio(t *testing.T) {
	cli := image.New(nil)
	_, err := cli.Prepare("p", "m", image.WithAspectRatio("nope"))
	if err == nil {
		t.Fatal("expected aspect error")
	}
}

func TestMultiFileIDs(t *testing.T) {
	mock := &captureImage{}
	srv, err := testutil.Start(func(s *grpc.Server) { xaiv1.RegisterImageServer(s, mock) })
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := image.New(srv.Conn)
	_, err = cli.Sample(context.Background(), "p", "m",
		image.WithImageFileIDs("f1", "f2"),
		image.WithImageURLs("http://a"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(mock.last.Images) != 3 {
		t.Fatalf("images len=%d", len(mock.last.Images))
	}
}

func TestSamplesMultiN(t *testing.T) {
	srv, err := testutil.Start(func(s *grpc.Server) {
		xaiv1.RegisterImageServer(s, &multiImage{n: 2})
	})
	if err != nil {
		t.Fatal(err)
	}
	defer srv.Close()
	cli := image.New(srv.Conn)

	// Sample must reject n>1 (no silent drop)
	_, err = cli.Sample(context.Background(), "p", "m", image.WithN(2))
	if err == nil {
		t.Fatal("expected Sample error for n>1")
	}

	all, err := cli.Samples(context.Background(), "multi-prompt", "m", image.WithN(2))
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("len=%d", len(all))
	}
	u0, err := all[0].URL()
	if err != nil || u0 == "" {
		t.Fatalf("img0: %v %q", err, u0)
	}
	u1, err := all[1].URL()
	if err != nil || u1 == "" || u1 == u0 {
		t.Fatalf("img1: %v %q vs %q", err, u1, u0)
	}
	if all[0].Prompt() != "multi-prompt" {
		t.Fatalf("prompt=%q", all[0].Prompt())
	}
	if !all[0].RespectModeration() {
		t.Fatal("moderation")
	}
}

type multiImage struct {
	xaiv1.UnimplementedImageServer
	n int
}

func (m *multiImage) GenerateImage(ctx context.Context, req *xaiv1.GenerateImageRequest) (*xaiv1.ImageResponse, error) {
	n := m.n
	if req.N != nil && int(*req.N) > 0 {
		n = int(*req.N)
	}
	if n < 1 {
		n = 1
	}
	imgs := make([]*xaiv1.GeneratedImage, n)
	for i := 0; i < n; i++ {
		imgs[i] = &xaiv1.GeneratedImage{
			RespectModeration: true,
			Image:             &xaiv1.GeneratedImage_Url{Url: "https://img.example/" + string(rune('a'+i)) + ".png"},
		}
	}
	// fix URLs without rune issues
	for i := 0; i < n; i++ {
		imgs[i].Image = &xaiv1.GeneratedImage_Url{Url: "https://img.example/" + itoa(i) + ".png"}
	}
	return &xaiv1.ImageResponse{Model: req.Model, Images: imgs}, nil
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
