// Package video generates and extends videos via the xAI Video API.
//
// Generate starts a deferred job and polls until completion. Inputs support
// URL and Files API file ids with mutual exclusion validation. StorageOptions
// are wired into generate/extend requests when set.
package video

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"

	"github.com/fun7257/xai-sdk-go/files"
	"github.com/fun7257/xai-sdk-go/internal/cost"
	"github.com/fun7257/xai-sdk-go/internal/poll"
	"github.com/fun7257/xai-sdk-go/telemetry"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// GenerationError is returned when video generation fails.
type GenerationError struct {
	Code    string
	Message string
}

// Error implements the error interface.
func (e *GenerationError) Error() string {
	return fmt.Sprintf("video generation failed [%s]: %s", e.Code, e.Message)
}

// Client generates and extends videos.
type Client struct{ stub xaiv1.VideoClient }

// New creates a Video client.
func New(cc grpc.ClientConnInterface) *Client {
	return &Client{stub: xaiv1.NewVideoClient(cc)}
}

// GenerateOption configures video generation.
type GenerateOption func(*genCfg)
type genCfg struct {
	duration     *int32
	imageURL     string
	imageFileID  string
	videoURL     string
	videoFileID  string
	refs         []string
	refFileIDs   []string
	aspectStr    string
	aspectSet    bool
	resStr       string
	resSet       bool
	storage      *files.StorageOptions
	pollTimeout  time.Duration
	pollInterval time.Duration
}

// WithDuration sets video duration in seconds.
func WithDuration(s int32) GenerateOption { return func(c *genCfg) { c.duration = &s } }

// WithImageURL sets an I2V first-frame image URL.
func WithImageURL(u string) GenerateOption { return func(c *genCfg) { c.imageURL = u } }

// WithImageFileID sets an I2V first-frame image file id.
func WithImageFileID(id string) GenerateOption {
	return func(c *genCfg) { c.imageFileID = id }
}

// WithVideoURL sets a video URL (edit path).
func WithVideoURL(u string) GenerateOption { return func(c *genCfg) { c.videoURL = u } }

// WithVideoFileID sets a video file id (edit path).
func WithVideoFileID(id string) GenerateOption {
	return func(c *genCfg) { c.videoFileID = id }
}

// WithReferenceImages sets R2V reference image URLs.
func WithReferenceImages(urls ...string) GenerateOption {
	return func(c *genCfg) { c.refs = urls }
}

// WithReferenceImageFileIDs sets R2V reference image file ids.
func WithReferenceImageFileIDs(ids ...string) GenerateOption {
	return func(c *genCfg) { c.refFileIDs = ids }
}

// WithAspectRatio sets aspect ratio (e.g. "16:9", "1:1", "9:16", "4:3", "3:4", "3:2", "2:3").
// Unknown values cause Prepare/Generate to return an error (no silent fallback).
func WithAspectRatio(s string) GenerateOption {
	return func(c *genCfg) {
		c.aspectStr = s
		c.aspectSet = true
	}
}

// WithResolution sets resolution ("480p" or "720p").
// Unknown values cause Generate/Prepare to return an error (no silent fallback).
func WithResolution(s string) GenerateOption {
	return func(c *genCfg) {
		c.resStr = s
		c.resSet = true
	}
}

// WithStorage sets Files API storage options.
func WithStorage(s files.StorageOptions) GenerateOption {
	return func(c *genCfg) { c.storage = &s }
}

// WithPollTimeout overrides the default 10m poll timeout for Generate/Extend.
func WithPollTimeout(d time.Duration) GenerateOption {
	return func(c *genCfg) { c.pollTimeout = d }
}

// WithPollInterval overrides the default 1s poll interval.
func WithPollInterval(d time.Duration) GenerateOption {
	return func(c *genCfg) { c.pollInterval = d }
}

func videoAspect(s string) (xaiv1.VideoAspectRatio, error) {
	m := map[string]xaiv1.VideoAspectRatio{
		"1:1":  xaiv1.VideoAspectRatio_VIDEO_ASPECT_RATIO_1_1,
		"16:9": xaiv1.VideoAspectRatio_VIDEO_ASPECT_RATIO_16_9,
		"9:16": xaiv1.VideoAspectRatio_VIDEO_ASPECT_RATIO_9_16,
		"4:3":  xaiv1.VideoAspectRatio_VIDEO_ASPECT_RATIO_4_3,
		"3:4":  xaiv1.VideoAspectRatio_VIDEO_ASPECT_RATIO_3_4,
		"3:2":  xaiv1.VideoAspectRatio_VIDEO_ASPECT_RATIO_3_2,
		"2:3":  xaiv1.VideoAspectRatio_VIDEO_ASPECT_RATIO_2_3,
	}
	if v, ok := m[s]; ok {
		return v, nil
	}
	return 0, fmt.Errorf("unknown video aspect ratio %q", s)
}

func videoRes(s string) (xaiv1.VideoResolution, error) {
	switch s {
	case "480p":
		return xaiv1.VideoResolution_VIDEO_RESOLUTION_480P, nil
	case "720p":
		return xaiv1.VideoResolution_VIDEO_RESOLUTION_720P, nil
	default:
		return 0, fmt.Errorf("unknown video resolution %q (want 480p or 720p)", s)
	}
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

func videoURLContent(url string) *xaiv1.VideoUrlContent {
	return &xaiv1.VideoUrlContent{Source: &xaiv1.VideoUrlContent_Url{Url: url}}
}

func videoFileContent(id string) *xaiv1.VideoUrlContent {
	return &xaiv1.VideoUrlContent{Source: &xaiv1.VideoUrlContent_FileId{FileId: id}}
}

func (c *Client) buildGenerate(prompt, model string, cfg *genCfg) (*xaiv1.GenerateVideoRequest, error) {
	if cfg.imageURL != "" && cfg.imageFileID != "" {
		return nil, fmt.Errorf("only one of image_url or image_file_id can be set for a request")
	}
	if cfg.videoURL != "" && cfg.videoFileID != "" {
		return nil, fmt.Errorf("only one of video_url or video_file_id can be set for a request")
	}
	// I2V vs R2V: image and reference images are distinct product paths;
	// Python allows them separately — keep both if set.

	var aspect *xaiv1.VideoAspectRatio
	if cfg.aspectSet {
		v, err := videoAspect(cfg.aspectStr)
		if err != nil {
			return nil, err
		}
		aspect = &v
	}
	var res *xaiv1.VideoResolution
	if cfg.resSet {
		v, err := videoRes(cfg.resStr)
		if err != nil {
			return nil, err
		}
		res = &v
	}

	req := &xaiv1.GenerateVideoRequest{
		Prompt: prompt, Model: model, Duration: cfg.duration,
		AspectRatio: aspect, Resolution: res,
	}
	if cfg.imageURL != "" {
		req.Image = imageURLContent(cfg.imageURL)
	} else if cfg.imageFileID != "" {
		req.Image = imageFileContent(cfg.imageFileID)
	}
	if cfg.videoURL != "" {
		req.Video = videoURLContent(cfg.videoURL)
	} else if cfg.videoFileID != "" {
		req.Video = videoFileContent(cfg.videoFileID)
	}
	for _, u := range cfg.refs {
		req.ReferenceImages = append(req.ReferenceImages, imageURLContent(u))
	}
	for _, id := range cfg.refFileIDs {
		req.ReferenceImages = append(req.ReferenceImages, imageFileContent(id))
	}
	if cfg.storage != nil {
		req.StorageOptions = cfg.storage.Proto()
	}
	return req, nil
}

func (c *Client) buildExtend(prompt, model, videoURL, videoFileID string, cfg *genCfg) (*xaiv1.ExtendVideoRequest, error) {
	if videoURL != "" && videoFileID != "" {
		return nil, fmt.Errorf("only one of video_url or video_file_id can be set for a request")
	}
	if videoURL == "" && videoFileID == "" {
		return nil, fmt.Errorf("one of video_url or video_file_id must be set for a request")
	}
	req := &xaiv1.ExtendVideoRequest{
		Prompt: prompt, Model: model, Duration: cfg.duration,
	}
	if videoURL != "" {
		req.Video = videoURLContent(videoURL)
	} else {
		req.Video = videoFileContent(videoFileID)
	}
	if cfg.storage != nil {
		req.StorageOptions = cfg.storage.Proto()
	}
	return req, nil
}

// Start begins deferred generation and returns request id.
func (c *Client) Start(ctx context.Context, prompt, model string, opts ...GenerateOption) (string, error) {
	cfg := &genCfg{}
	for _, o := range opts {
		o(cfg)
	}
	req, err := c.buildGenerate(prompt, model, cfg)
	if err != nil {
		return "", err
	}
	r, err := c.stub.GenerateVideo(ctx, req)
	if err != nil {
		return "", err
	}
	return r.RequestId, nil
}

// Get polls deferred video status once.
func (c *Client) Get(ctx context.Context, requestID string) (*xaiv1.GetDeferredVideoResponse, error) {
	return c.stub.GetDeferredVideo(ctx, &xaiv1.GetDeferredVideoRequest{RequestId: requestID})
}

// Generate starts generation and polls until complete.
func (c *Client) Generate(ctx context.Context, prompt, model string, opts ...GenerateOption) (*Response, error) {
	ctx, span := telemetry.StartSpan(ctx, telemetry.SpanVideoGenerate,
		attribute.String("gen_ai.system", "xai"),
		attribute.String("gen_ai.request.model", model),
	)
	var err error
	defer func() { telemetry.EndSpan(span, err) }()

	cfg := &genCfg{pollTimeout: 10 * time.Minute, pollInterval: time.Second}
	for _, o := range opts {
		o(cfg)
	}
	var id string
	id, err = c.Start(ctx, prompt, model, opts...)
	if err != nil {
		return nil, err
	}
	var resp *Response
	resp, err = c.wait(ctx, id, cfg.pollTimeout, cfg.pollInterval)
	return resp, err
}

func (c *Client) wait(ctx context.Context, id string, timeout, interval time.Duration) (*Response, error) {
	cfg := poll.Config{Timeout: timeout, Interval: interval, Context: "waiting for video generation"}
	var final *xaiv1.VideoResponse
	err := poll.Wait(ctx, cfg, func(ctx context.Context) (bool, error) {
		r, err := c.Get(ctx, id)
		if err != nil {
			return false, err
		}
		switch r.Status {
		case xaiv1.DeferredStatus_DONE:
			final = r.GetResponse()
			if final == nil {
				return false, fmt.Errorf("video generation completed with empty response")
			}
			return true, nil
		case xaiv1.DeferredStatus_EXPIRED:
			return false, fmt.Errorf("video request expired")
		case xaiv1.DeferredStatus_FAILED:
			resp := r.GetResponse()
			if resp != nil && resp.Error != nil {
				return false, &GenerationError{Code: resp.Error.GetCode(), Message: resp.Error.GetMessage()}
			}
			return false, &GenerationError{Code: "failed", Message: "video generation failed"}
		default:
			return false, nil
		}
	})
	if err != nil {
		return nil, err
	}
	if final == nil {
		return nil, fmt.Errorf("video generation completed with empty response")
	}
	return &Response{proto: final}, nil
}

// ExtendStart begins extend without polling and returns the deferred request id (VID-01).
func (c *Client) ExtendStart(ctx context.Context, prompt, model, videoURL string, opts ...GenerateOption) (string, error) {
	return c.ExtendStartWith(ctx, prompt, model, videoURL, "", opts...)
}

// ExtendStartWith begins extend (URL or file id) without polling; pair with Get/wait.
func (c *Client) ExtendStartWith(ctx context.Context, prompt, model, videoURL, videoFileID string, opts ...GenerateOption) (string, error) {
	cfg := &genCfg{}
	for _, o := range opts {
		o(cfg)
	}
	req, err := c.buildExtend(prompt, model, videoURL, videoFileID, cfg)
	if err != nil {
		return "", err
	}
	start, err := c.stub.ExtendVideo(ctx, req)
	if err != nil {
		return "", err
	}
	return start.RequestId, nil
}

// Extend extends an existing video with a prompt (video URL form).
func (c *Client) Extend(ctx context.Context, prompt, model, videoURL string, opts ...GenerateOption) (*Response, error) {
	return c.ExtendWith(ctx, prompt, model, videoURL, "", opts...)
}

// ExtendWith extends a video using either URL or file id.
func (c *Client) ExtendWith(ctx context.Context, prompt, model, videoURL, videoFileID string, opts ...GenerateOption) (*Response, error) {
	ctx, span := telemetry.StartSpan(ctx, telemetry.SpanVideoExtend,
		attribute.String("gen_ai.system", "xai"),
		attribute.String("gen_ai.request.model", model),
	)
	var err error
	defer func() { telemetry.EndSpan(span, err) }()

	cfg := &genCfg{pollTimeout: 10 * time.Minute, pollInterval: time.Second}
	for _, o := range opts {
		o(cfg)
	}
	var req *xaiv1.ExtendVideoRequest
	req, err = c.buildExtend(prompt, model, videoURL, videoFileID, cfg)
	if err != nil {
		return nil, err
	}
	var start *xaiv1.StartDeferredResponse
	start, err = c.stub.ExtendVideo(ctx, req)
	if err != nil {
		return nil, err
	}
	var resp *Response
	resp, err = c.wait(ctx, start.RequestId, cfg.pollTimeout, cfg.pollInterval)
	return resp, err
}

// Prepare builds a GenerateVideoRequest for batch.
func (c *Client) Prepare(prompt, model string, opts ...GenerateOption) (*xaiv1.GenerateVideoRequest, error) {
	cfg := &genCfg{}
	for _, o := range opts {
		o(cfg)
	}
	return c.buildGenerate(prompt, model, cfg)
}

// PrepareExtension builds an ExtendVideoRequest for batch (URL form).
func (c *Client) PrepareExtension(prompt, model, videoURL string, opts ...GenerateOption) (*xaiv1.ExtendVideoRequest, error) {
	return c.PrepareExtensionWith(prompt, model, videoURL, "", opts...)
}

// PrepareExtensionWith builds an ExtendVideoRequest for batch.
func (c *Client) PrepareExtensionWith(prompt, model, videoURL, videoFileID string, opts ...GenerateOption) (*xaiv1.ExtendVideoRequest, error) {
	cfg := &genCfg{}
	for _, o := range opts {
		o(cfg)
	}
	return c.buildExtend(prompt, model, videoURL, videoFileID, cfg)
}

// Response wraps a completed video response.
type Response struct{ proto *xaiv1.VideoResponse }

// NewResponse wraps a proto video response.
func NewResponse(proto *xaiv1.VideoResponse) *Response {
	return &Response{proto: proto}
}

// Proto returns the underlying proto.
func (r *Response) Proto() *xaiv1.VideoResponse { return r.proto }

// URL returns the generated video URL.
func (r *Response) URL() (string, error) {
	if r.proto == nil || r.proto.Video == nil {
		return "", fmt.Errorf("no video")
	}
	if u := r.proto.Video.GetUrl(); u != "" {
		return u, nil
	}
	if !r.proto.Video.RespectModeration {
		return "", fmt.Errorf("video did not respect moderation rules; URL is not available")
	}
	return "", fmt.Errorf("video URL missing from response")
}

// RespectModeration reports whether the video passed moderation (VID-02).
func (r *Response) RespectModeration() bool {
	if r.proto == nil || r.proto.Video == nil {
		return true
	}
	return r.proto.Video.RespectModeration
}

// Duration returns the generated video duration in seconds.
func (r *Response) Duration() int32 {
	if r.proto == nil || r.proto.Video == nil {
		return 0
	}
	return r.proto.Video.Duration
}

// Model returns the model name.
func (r *Response) Model() string {
	if r.proto == nil {
		return ""
	}
	return r.proto.Model
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
	if r.proto == nil || r.proto.Video == nil {
		return nil
	}
	return r.proto.Video.GetFileOutput()
}

// StorageError returns a storage failure message if present.
func (r *Response) StorageError() string {
	if r.proto == nil || r.proto.Video == nil {
		return ""
	}
	return r.proto.Video.GetStorageError()
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
