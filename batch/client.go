// Package batch manages batch jobs and typed batch results.
package batch

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/fun7257/xai-sdk-go/chat"
	"github.com/fun7257/xai-sdk-go/image"
	"github.com/fun7257/xai-sdk-go/video"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// Client manages batch jobs.
type Client struct{ stub xaiv1.BatchMgmtClient }

// New creates a Batch client.
func New(cc grpc.ClientConnInterface) *Client {
	return &Client{stub: xaiv1.NewBatchMgmtClient(cc)}
}

// Create creates a batch. Optional inputFileID for JSONL-backed batches.
func (c *Client) Create(ctx context.Context, name string, inputFileID string) (*xaiv1.Batch, error) {
	return c.stub.CreateBatch(ctx, &xaiv1.CreateBatchRequest{Name: name, InputFileId: inputFileID})
}

// Add appends requests to a batch.
func (c *Client) Add(ctx context.Context, batchID string, requests ...*xaiv1.BatchRequest) error {
	_, err := c.stub.AddBatchRequests(ctx, &xaiv1.AddBatchRequestsRequest{
		BatchId: batchID, BatchRequests: requests,
	})
	return err
}

// Get retrieves a batch.
func (c *Client) Get(ctx context.Context, batchID string) (*xaiv1.Batch, error) {
	return c.stub.GetBatch(ctx, &xaiv1.GetBatchRequest{BatchId: batchID})
}

// Cancel cancels a batch.
func (c *Client) Cancel(ctx context.Context, batchID string) (*xaiv1.Batch, error) {
	return c.stub.CancelBatch(ctx, &xaiv1.CancelBatchRequest{BatchId: batchID})
}

// List lists batches.
func (c *Client) List(ctx context.Context, paginationToken string) (*xaiv1.ListBatchesResponse, error) {
	req := &xaiv1.ListBatchesRequest{}
	if paginationToken != "" {
		req.PaginationToken = &paginationToken
	}
	return c.stub.ListBatches(ctx, req)
}

// ListAll returns all batches, following pagination tokens until the final page.
func (c *Client) ListAll(ctx context.Context) ([]*xaiv1.Batch, error) {
	var out []*xaiv1.Batch
	token := ""
	for {
		resp, err := c.List(ctx, token)
		if err != nil {
			return nil, err
		}
		out = append(out, resp.GetBatches()...)
		next := resp.GetPaginationToken()
		// Stop on the final page or a non-advancing token (defensive).
		if next == "" || next == token {
			return out, nil
		}
		token = next
	}
}

// ListBatchRequests lists request metadata for a batch.
func (c *Client) ListBatchRequests(ctx context.Context, batchID, paginationToken string) (*xaiv1.ListBatchRequestMetadataResponse, error) {
	req := &xaiv1.ListBatchRequestMetadataRequest{BatchId: batchID}
	if paginationToken != "" {
		req.PaginationToken = &paginationToken
	}
	return c.stub.ListBatchRequestMetadata(ctx, req)
}

// ListBatchResults lists results for a batch.
func (c *Client) ListBatchResults(ctx context.Context, batchID, paginationToken string) (*xaiv1.ListBatchResultsResponse, error) {
	req := &xaiv1.ListBatchResultsRequest{BatchId: batchID}
	if paginationToken != "" {
		req.PaginationToken = &paginationToken
	}
	return c.stub.ListBatchResults(ctx, req)
}

// GetBatchRequestResult retrieves a single request and its result.
func (c *Client) GetBatchRequestResult(ctx context.Context, batchID, batchRequestID string) (*xaiv1.GetBatchRequestResultResponse, error) {
	return c.stub.GetBatchRequestResult(ctx, &xaiv1.GetBatchRequestResultRequest{
		BatchId: batchID, BatchRequestId: batchRequestID,
	})
}

// ChatBatchRequest wraps a chat completion for batch.
func ChatBatchRequest(req *xaiv1.GetCompletionsRequest, batchRequestID string) *xaiv1.BatchRequest {
	br := &xaiv1.BatchRequest{Request: &xaiv1.BatchRequest_CompletionRequest{CompletionRequest: req}}
	if batchRequestID != "" {
		br.BatchRequestId = &batchRequestID
	}
	return br
}

// ImageBatchRequest wraps image generation for batch.
func ImageBatchRequest(req *xaiv1.GenerateImageRequest, batchRequestID string) *xaiv1.BatchRequest {
	br := &xaiv1.BatchRequest{Request: &xaiv1.BatchRequest_ImageRequest{ImageRequest: req}}
	if batchRequestID != "" {
		br.BatchRequestId = &batchRequestID
	}
	return br
}

// VideoBatchRequest wraps video generation for batch.
func VideoBatchRequest(req *xaiv1.GenerateVideoRequest, batchRequestID string) *xaiv1.BatchRequest {
	br := &xaiv1.BatchRequest{Request: &xaiv1.BatchRequest_VideoRequest{VideoRequest: req}}
	if batchRequestID != "" {
		br.BatchRequestId = &batchRequestID
	}
	return br
}

// Result wraps a BatchResult with typed accessors.
type Result struct {
	proto *xaiv1.BatchResult
}

// NewResult wraps a proto BatchResult.
func NewResult(p *xaiv1.BatchResult) *Result {
	return &Result{proto: p}
}

// Proto returns the underlying BatchResult.
func (r *Result) Proto() *xaiv1.BatchResult { return r.proto }

// BatchRequestID returns the batch request identifier.
func (r *Result) BatchRequestID() string {
	if r == nil || r.proto == nil {
		return ""
	}
	return r.proto.BatchRequestId
}

// Succeeded reports whether the result contains a successful response.
func (r *Result) Succeeded() bool {
	if r == nil || r.proto == nil {
		return false
	}
	return r.proto.GetResponse() != nil
}

// Failed reports whether the result contains an error status.
func (r *Result) Failed() bool {
	if r == nil || r.proto == nil {
		return false
	}
	return r.proto.GetError() != nil
}

// Error returns a gRPC status error when Failed, else nil.
// The full status proto is preserved, so structured details remain available
// via status.Convert(err).Details().
func (r *Result) Error() error {
	if !r.Failed() {
		return nil
	}
	st := r.proto.GetError()
	if st == nil {
		return nil
	}
	return status.FromProto(st).Err()
}

func (r *Result) responseData() *xaiv1.BatchResultData {
	if r == nil || r.proto == nil {
		return nil
	}
	return r.proto.GetResponse()
}

// ChatResponse returns a chat.Response when the result is a completion.
func (r *Result) ChatResponse() (*chat.Response, bool) {
	data := r.responseData()
	if data == nil {
		return nil, false
	}
	pb := data.GetCompletionResponse()
	if pb == nil {
		return nil, false
	}
	z := 0
	return chat.NewResponse(pb, &z), true
}

// ImageResponse returns an image.Response when the result is image generation.
func (r *Result) ImageResponse() (*image.Response, bool) {
	data := r.responseData()
	if data == nil {
		return nil, false
	}
	pb := data.GetImageResponse()
	if pb == nil {
		return nil, false
	}
	return image.NewResponse(pb, 0), true
}

// VideoResponse returns a video.Response when the result is video generation.
func (r *Result) VideoResponse() (*video.Response, bool) {
	data := r.responseData()
	if data == nil {
		return nil, false
	}
	pb := data.GetVideoResponse()
	if pb == nil {
		return nil, false
	}
	return video.NewResponse(pb), true
}

// WrapResults converts a slice of BatchResult into Result wrappers.
func WrapResults(in []*xaiv1.BatchResult) []*Result {
	out := make([]*Result, 0, len(in))
	for _, p := range in {
		out = append(out, NewResult(p))
	}
	return out
}

// FilterSucceeded returns only successful results.
func FilterSucceeded(in []*Result) []*Result {
	var out []*Result
	for _, r := range in {
		if r.Succeeded() {
			out = append(out, r)
		}
	}
	return out
}

// FilterFailed returns only failed results.
func FilterFailed(in []*Result) []*Result {
	var out []*Result
	for _, r := range in {
		if r.Failed() {
			out = append(out, r)
		}
	}
	return out
}
