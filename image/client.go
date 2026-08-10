// Package image generates images via the xAI Image API.
//
// Default response format is URL. Inputs support image URL(s) and/or Files API
// file id(s) with mutual exclusion rules. StorageOptions are wired into
// GenerateImageRequest when set.
package image

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/fun7257/xai-sdk-go/files"
	"github.com/fun7257/xai-sdk-go/internal/cost"
	"github.com/fun7257/xai-sdk-go/telemetry"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"
)

// MaxDownloadBytes is the maximum size Download will buffer (100 MiB).
const MaxDownloadBytes = 100 << 20

// Client generates images.
type Client struct{ stub xaiv1.ImageClient }

// New creates an Image client.
func New(cc grpc.ClientConnInterface) *Client {
	return &Client{stub: xaiv1.NewImageClient(cc)}
}

// SampleOption configures image generation.
type SampleOption func(*sampleCfg)
type sampleCfg struct {
	n            *int32
	user         string
	imageURL     string
	imageFileID  string
	imageURLs    []string
	imageFileIDs []string
	format       xaiv1.ImageFormat
	aspectStr    string
	aspectSet    bool
	resolution   *xaiv1.ImageResolution
	storage      *files.StorageOptions
}

// WithN sets the number of images to generate.
func WithN(n int32) SampleOption { return func(c *sampleCfg) { c.n = &n } }

// WithUser sets the opaque user identifier.
func WithUser(u string) SampleOption { return func(c *sampleCfg) { c.user = u } }

// WithImageURL sets a single reference image URL (mutually exclusive with WithImageFileID
// and multi-image options).
func WithImageURL(url string) SampleOption { return func(c *sampleCfg) { c.imageURL = url } }

// WithImageFileID sets a single reference image from a Files API id.
func WithImageFileID(id string) SampleOption {
	return func(c *sampleCfg) { c.imageFileID = id }
}

// WithImageURLs sets multi-reference image URLs (may be combined with WithImageFileIDs).
func WithImageURLs(urls ...string) SampleOption {
	return func(c *sampleCfg) { c.imageURLs = urls }
}

// WithImageFileIDs sets multi-reference image file ids (may be combined with WithImageURLs).
func WithImageFileIDs(ids ...string) SampleOption {
	return func(c *sampleCfg) { c.imageFileIDs = ids }
}

// WithFormatURL requests URL responses (default).
func WithFormatURL() SampleOption {
	return func(c *sampleCfg) { c.format = xaiv1.ImageFormat_IMG_FORMAT_URL }
}

// WithFormatBase64 requests base64 responses.
func WithFormatBase64() SampleOption {
	return func(c *sampleCfg) { c.format = xaiv1.ImageFormat_IMG_FORMAT_BASE64 }
}

// WithAspectRatio sets aspect ratio (e.g. "1:1", "16:9", "9:16", "3:4", "4:3", "2:3", "3:2", "auto").
// Unknown values cause Sample/Prepare to return an error (no silent fallback).
func WithAspectRatio(ar string) SampleOption {
	return func(c *sampleCfg) {
		c.aspectStr = ar
		c.aspectSet = true
	}
}

// WithResolution sets resolution ("1k" or "2k"). Unknown values default to 1k.
func WithResolution(r string) SampleOption {
	return func(c *sampleCfg) {
		v := resolution(r)
		c.resolution = &v
	}
}

// WithStorage sets Files API storage options for the generated asset.
func WithStorage(s files.StorageOptions) SampleOption {
	return func(c *sampleCfg) { c.storage = &s }
}

func aspect(s string) (xaiv1.ImageAspectRatio, error) {
	m := map[string]xaiv1.ImageAspectRatio{
		"1:1":  xaiv1.ImageAspectRatio_IMG_ASPECT_RATIO_1_1,
		"3:4":  xaiv1.ImageAspectRatio_IMG_ASPECT_RATIO_3_4,
		"4:3":  xaiv1.ImageAspectRatio_IMG_ASPECT_RATIO_4_3,
		"9:16": xaiv1.ImageAspectRatio_IMG_ASPECT_RATIO_9_16,
		"16:9": xaiv1.ImageAspectRatio_IMG_ASPECT_RATIO_16_9,
		"2:3":  xaiv1.ImageAspectRatio_IMG_ASPECT_RATIO_2_3,
		"3:2":  xaiv1.ImageAspectRatio_IMG_ASPECT_RATIO_3_2,
		"auto": xaiv1.ImageAspectRatio_IMG_ASPECT_RATIO_AUTO,
	}
	if v, ok := m[s]; ok {
		return v, nil
	}
	return 0, fmt.Errorf("unknown image aspect ratio %q", s)
}

func resolution(s string) xaiv1.ImageResolution {
	if s == "2k" {
		return xaiv1.ImageResolution_IMG_RESOLUTION_2K
	}
	return xaiv1.ImageResolution_IMG_RESOLUTION_1K
}

func imageURLContent(url string) *xaiv1.ImageUrlContent {
	return &xaiv1.ImageUrlContent{
		Source: &xaiv1.ImageUrlContent_ImageUrl{ImageUrl: url},
		Detail: xaiv1.ImageDetail_DETAIL_AUTO,
	}
}

func imageFileContent(id string) *xaiv1.ImageUrlContent {
	return &xaiv1.ImageUrlContent{
		Source: &xaiv1.ImageUrlContent_FileId{FileId: id},
		Detail: xaiv1.ImageDetail_DETAIL_AUTO,
	}
}

func (c *Client) buildRequest(prompt, model string, cfg *sampleCfg) (*xaiv1.GenerateImageRequest, error) {
	if cfg.imageURL != "" && cfg.imageFileID != "" {
		return nil, fmt.Errorf("only one of image_url or image_file_id can be set for a request")
	}
	hasSingle := cfg.imageURL != "" || cfg.imageFileID != ""
	hasMulti := len(cfg.imageURLs) > 0 || len(cfg.imageFileIDs) > 0
	if hasSingle && hasMulti {
		return nil, fmt.Errorf("only one of image_url/image_file_id or image_urls/image_file_ids can be set for a request")
	}

	var aspectRatio *xaiv1.ImageAspectRatio
	if cfg.aspectSet {
		v, err := aspect(cfg.aspectStr)
		if err != nil {
			return nil, err
		}
		aspectRatio = &v
	}

	format := cfg.format
	if format == xaiv1.ImageFormat_IMG_FORMAT_INVALID {
		format = xaiv1.ImageFormat_IMG_FORMAT_URL
	}
	req := &xaiv1.GenerateImageRequest{
		Prompt: prompt, Model: model, User: cfg.user, Format: format,
		N: cfg.n, AspectRatio: aspectRatio, Resolution: cfg.resolution,
	}
	if cfg.imageURL != "" {
		req.Image = imageURLContent(cfg.imageURL)
	} else if cfg.imageFileID != "" {
		req.Image = imageFileContent(cfg.imageFileID)
	}
	for _, u := range cfg.imageURLs {
		req.Images = append(req.Images, imageURLContent(u))
	}
	for _, id := range cfg.imageFileIDs {
		req.Images = append(req.Images, imageFileContent(id))
	}
	if cfg.storage != nil {
		req.StorageOptions = cfg.storage.Proto()
	}
	return req, nil
}

// Sample generates a single image (n=1). Preferred one-shot API.
// For multiple images use Samples with WithN(k).
func (c *Client) Sample(ctx context.Context, prompt, model string, opts ...SampleOption) (*Response, error) {
	cfg := &sampleCfg{format: xaiv1.ImageFormat_IMG_FORMAT_URL}
	for _, o := range opts {
		o(cfg)
	}
	if cfg.n != nil && *cfg.n > 1 {
		return nil, fmt.Errorf("image.Sample: n=%d > 1; use Samples with WithN to receive all images", *cfg.n)
	}
	all, err := c.samples(ctx, prompt, model, cfg)
	if err != nil {
		return nil, err
	}
	if len(all) == 0 {
		return nil, fmt.Errorf("no images in response")
	}
	return all[0], nil
}

// Samples generates one or more images (never silently drops extras).
// Use WithN(k) for k images; default is server/default count or 1.
func (c *Client) Samples(ctx context.Context, prompt, model string, opts ...SampleOption) ([]*Response, error) {
	cfg := &sampleCfg{format: xaiv1.ImageFormat_IMG_FORMAT_URL}
	for _, o := range opts {
		o(cfg)
	}
	return c.samples(ctx, prompt, model, cfg)
}

func (c *Client) samples(ctx context.Context, prompt, model string, cfg *sampleCfg) ([]*Response, error) {
	ctx, span := telemetry.StartSpan(ctx, telemetry.SpanImageSample,
		attribute.String("gen_ai.system", "xai"),
		attribute.String("gen_ai.request.model", model),
	)
	var err error
	defer func() { telemetry.EndSpan(span, err) }()

	var req *xaiv1.GenerateImageRequest
	req, err = c.buildRequest(prompt, model, cfg)
	if err != nil {
		return nil, err
	}
	var pb *xaiv1.ImageResponse
	pb, err = c.stub.GenerateImage(ctx, req)
	if err != nil {
		return nil, err
	}
	if pb == nil || len(pb.Images) == 0 {
		return nil, fmt.Errorf("no images in response")
	}
	out := make([]*Response, len(pb.Images))
	for i := range pb.Images {
		out[i] = &Response{proto: pb, index: i, prompt: prompt}
	}
	return out, nil
}

// Prepare builds a batch request for image generation.
func (c *Client) Prepare(prompt, model string, opts ...SampleOption) (*xaiv1.GenerateImageRequest, error) {
	cfg := &sampleCfg{format: xaiv1.ImageFormat_IMG_FORMAT_URL}
	for _, o := range opts {
		o(cfg)
	}
	return c.buildRequest(prompt, model, cfg)
}

// Response wraps image generation results.
type Response struct {
	proto  *xaiv1.ImageResponse
	index  int
	prompt string // request prompt (API ImageResponse has no prompt field)
}

// NewResponse wraps a proto image response. index selects which generated image (default 0).
func NewResponse(proto *xaiv1.ImageResponse, index int) *Response {
	return &Response{proto: proto, index: index}
}

// Proto returns the underlying proto.
func (r *Response) Proto() *xaiv1.ImageResponse { return r.proto }

func (r *Response) image() *xaiv1.GeneratedImage {
	if r.proto == nil || r.index >= len(r.proto.Images) {
		return nil
	}
	return r.proto.Images[r.index]
}

// URL returns the generated image URL when format was URL.
func (r *Response) URL() (string, error) {
	img := r.image()
	if img == nil {
		return "", fmt.Errorf("no image")
	}
	if u := img.GetUrl(); u != "" {
		return u, nil
	}
	if !img.GetRespectModeration() {
		return "", fmt.Errorf("image did not respect moderation rules; URL is not available")
	}
	return "", fmt.Errorf("image was not returned via URL")
}

// Base64 returns the generated image base64 when format was base64.
func (r *Response) Base64() (string, error) {
	img := r.image()
	if img == nil {
		return "", fmt.Errorf("no image")
	}
	if b := img.GetBase64(); b != "" {
		return b, nil
	}
	if !img.GetRespectModeration() {
		return "", fmt.Errorf("image did not respect moderation rules; base64 is not available")
	}
	return "", fmt.Errorf("image was not returned via base64")
}

// Model returns the model that produced the image.
func (r *Response) Model() string {
	if r.proto == nil {
		return ""
	}
	return r.proto.Model
}

// Prompt returns the request prompt associated with this response (client-side;
// the public ImageResponse proto does not echo prompt).
func (r *Response) Prompt() string {
	if r == nil {
		return ""
	}
	return r.prompt
}

// RespectModeration reports whether the selected image passed moderation.
// Defaults to true when the image is missing (no false-negative on empty).
func (r *Response) RespectModeration() bool {
	img := r.image()
	if img == nil {
		return true
	}
	return img.GetRespectModeration()
}

// Usage returns sampling usage if present.
func (r *Response) Usage() *xaiv1.SamplingUsage {
	if r.proto == nil {
		return nil
	}
	return r.proto.Usage
}

// CostUSD returns estimated cost when reported.
func (r *Response) CostUSD() (float64, bool) {
	u := r.Usage()
	if u == nil || u.CostInUsdTicks == nil {
		return 0, false
	}
	return cost.FromTicks(*u.CostInUsdTicks, true)
}

// FileOutput returns storage metadata when the asset was persisted.
func (r *Response) FileOutput() *xaiv1.FileOutput {
	img := r.image()
	if img == nil {
		return nil
	}
	return img.GetFileOutput()
}

// StorageError returns a storage failure message if present.
func (r *Response) StorageError() string {
	img := r.image()
	if img == nil {
		return ""
	}
	return img.GetStorageError()
}

// PublicURL returns the public URL from FileOutput when available.
func (r *Response) PublicURL() string {
	fo := r.FileOutput()
	if fo == nil {
		return ""
	}
	return fo.GetPublicUrl()
}

// PublicURLError returns the public URL error from FileOutput when available.
func (r *Response) PublicURLError() string {
	fo := r.FileOutput()
	if fo == nil {
		return ""
	}
	return fo.GetPublicUrlError()
}

// Download fetches the image bytes via the response URL using a short-lived HTTP client.
//
// Trust boundary: only call Download on URLs returned by the xAI API for this
// response. The client allows only http/https schemes, does not follow redirects
// to other schemes, and caps the body at MaxDownloadBytes (100 MiB).
func (r *Response) Download(ctx context.Context) ([]byte, error) {
	u, err := r.URL()
	if err != nil {
		return nil, err
	}
	parsed, err := url.Parse(u)
	if err != nil {
		return nil, fmt.Errorf("download image: invalid url: %w", err)
	}
	switch parsed.Scheme {
	case "https", "http":
		// http allowed for local tests; production API URLs are https.
	default:
		return nil, fmt.Errorf("download image: unsupported url scheme %q (only http/https)", parsed.Scheme)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}
	cli := &http.Client{
		Timeout: 60 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if req.URL.Scheme != "https" && req.URL.Scheme != "http" {
				return fmt.Errorf("redirect to unsupported scheme %q", req.URL.Scheme)
			}
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}
	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download image: HTTP %s", resp.Status)
	}
	limited := io.LimitReader(resp.Body, MaxDownloadBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > MaxDownloadBytes {
		return nil, fmt.Errorf("download image: body exceeds %d bytes", MaxDownloadBytes)
	}
	return data, nil
}
