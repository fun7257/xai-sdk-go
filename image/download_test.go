package image_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fun7257/xai-sdk-go/image"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

func urlResponse(u string) *image.Response {
	return image.NewResponse(&xaiv1.ImageResponse{
		Images: []*xaiv1.GeneratedImage{{
			Image:             &xaiv1.GeneratedImage_Url{Url: u},
			RespectModeration: true,
		}},
	}, 0)
}

func TestDownloadSuccess(t *testing.T) {
	payload := bytes.Repeat([]byte("img"), 1024)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(payload)
	}))
	defer ts.Close()

	data, err := urlResponse(ts.URL).Download(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(data, payload) {
		t.Fatalf("payload mismatch: got %d bytes", len(data))
	}
}

func TestDownloadRejectsNonHTTPScheme(t *testing.T) {
	_, err := urlResponse("ftp://example.com/img.png").Download(context.Background())
	if err == nil || !strings.Contains(err.Error(), "unsupported url scheme") {
		t.Fatalf("expected scheme rejection, got: %v", err)
	}
}

func TestDownloadRejectsRedirectToNonHTTPScheme(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "ftp://internal.host/secret", http.StatusFound)
	}))
	defer ts.Close()

	_, err := urlResponse(ts.URL).Download(context.Background())
	if err == nil || !strings.Contains(err.Error(), "redirect to unsupported scheme") {
		t.Fatalf("expected redirect scheme rejection, got: %v", err)
	}
}

func TestDownloadRejectsTooManyRedirects(t *testing.T) {
	var ts *httptest.Server
	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, ts.URL+"/loop", http.StatusFound)
	}))
	defer ts.Close()

	_, err := urlResponse(ts.URL).Download(context.Background())
	if err == nil || !strings.Contains(err.Error(), "too many redirects") {
		t.Fatalf("expected redirect limit, got: %v", err)
	}
}

func TestDownloadRejectsNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	_, err := urlResponse(ts.URL).Download(context.Background())
	if err == nil || !strings.Contains(err.Error(), "HTTP 404") {
		t.Fatalf("expected HTTP status error, got: %v", err)
	}
}

// zeroReader yields n zero bytes without allocating them all at once.
type zeroReader struct{ n int64 }

func (z *zeroReader) Read(p []byte) (int, error) {
	if z.n <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > z.n {
		p = p[:z.n]
	}
	for i := range p {
		p[i] = 0
	}
	z.n -= int64(len(p))
	return len(p), nil
}

func TestDownloadEnforcesSizeCap(t *testing.T) {
	if testing.Short() {
		t.Skip("streams >100MiB; skipped with -short")
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.Copy(w, &zeroReader{n: image.MaxDownloadBytes + 1})
	}))
	defer ts.Close()

	_, err := urlResponse(ts.URL).Download(context.Background())
	if err == nil || !strings.Contains(err.Error(), "exceeds") {
		t.Fatalf("expected size cap error, got: %v", err)
	}
}

func TestDownloadNoURL(t *testing.T) {
	r := image.NewResponse(&xaiv1.ImageResponse{
		Images: []*xaiv1.GeneratedImage{{RespectModeration: true}},
	}, 0)
	if _, err := r.Download(context.Background()); err == nil {
		t.Fatal("expected error when the response has no URL")
	}
}
