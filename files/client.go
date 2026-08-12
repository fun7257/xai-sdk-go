// Package files implements the xAI Files API (upload, list, public URLs).
//
// Uploads use 3 MiB chunks. List supports order, sort_by, and optional filter
// expressions. BatchUpload runs concurrent path uploads with an optional
// per-file completion callback.
//
// Progress hooks use func(delta, total int64). ExpiresAfter on upload sets a
// TTL measured from upload time (server range typically 1h–30d).
package files

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"google.golang.org/grpc"

	"github.com/fun7257/xai-sdk-go/telemetry"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

const chunkSize = 3 << 20 // 3 MiB

// MaxContentBytes is the maximum size Content will buffer (512 MiB, matching API file cap).
const MaxContentBytes = 512 << 20

// Client manages Files API operations.
type Client struct{ stub xaiv1.FilesClient }

// New creates a Files client on the given connection.
func New(cc grpc.ClientConnInterface) *Client {
	return &Client{stub: xaiv1.NewFilesClient(cc)}
}

// UploadOption configures upload.
type UploadOption func(*uploadCfg)
type uploadCfg struct {
	expiresAfter *int64
	progress     func(delta, total int64)
}

// MinExpiresAfter and MaxExpiresAfter bound the upload TTL accepted by the
// API (UploadFileInit.expires_after must be in [1h, 30d] inclusive; 0 and
// out-of-range values are rejected server-side). Upload validates locally so
// misconfigured TTLs fail fast instead of after streaming the whole file.
const (
	MinExpiresAfter = time.Hour
	MaxExpiresAfter = 30 * 24 * time.Hour
)

// WithExpiresAfter sets the file TTL measured from upload time.
// The API accepts TTLs in [MinExpiresAfter, MaxExpiresAfter]; Upload returns
// an error for out-of-range values before sending any data.
func WithExpiresAfter(d time.Duration) UploadOption {
	return func(c *uploadCfg) {
		s := int64(d.Seconds())
		c.expiresAfter = &s
	}
}

func validateExpiresAfter(seconds int64) error {
	if seconds < int64(MinExpiresAfter.Seconds()) || seconds > int64(MaxExpiresAfter.Seconds()) {
		return fmt.Errorf("files: expires_after %ds out of range [%s, %s]",
			seconds, MinExpiresAfter, MaxExpiresAfter)
	}
	return nil
}

// WithProgress registers a progress hook called after each chunk with
// (bytes_in_chunk, total_size). total may be 0 if unknown.
func WithProgress(fn func(delta, total int64)) UploadOption {
	return func(c *uploadCfg) { c.progress = fn }
}

// Upload streams a file to the API. name is the remote filename.
func (c *Client) Upload(ctx context.Context, name string, r io.Reader, size int64, opts ...UploadOption) (*xaiv1.File, error) {
	ctx, span := telemetry.StartSpan(ctx, telemetry.SpanFilesUpload,
		attribute.String("gen_ai.system", "xai"),
		attribute.String("file.name", name),
	)
	var err error
	defer func() { telemetry.EndSpan(span, err) }()

	cfg := &uploadCfg{}
	for _, o := range opts {
		o(cfg)
	}
	if cfg.expiresAfter != nil {
		if err = validateExpiresAfter(*cfg.expiresAfter); err != nil {
			return nil, err
		}
	}
	// Child context so abandon/error paths cancel the client stream promptly.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var stream xaiv1.Files_UploadFileClient
	stream, err = c.stub.UploadFile(ctx)
	if err != nil {
		return nil, err
	}
	// Per gRPC client-stream semantics, Send returns io.EOF when the server has
	// already terminated the stream; the real status (e.g. quota or size
	// rejection) is only available from CloseAndRecv. sendAll maps that case to
	// errServerClosedStream so it is never surfaced as a bare EOF.
	sendErr := c.sendAll(stream, name, r, size, cfg)
	if sendErr != nil && !errors.Is(sendErr, errServerClosedStream) {
		err = sendErr
		return nil, err
	}
	var f *xaiv1.File
	f, err = stream.CloseAndRecv()
	if err != nil {
		return nil, err
	}
	return f, nil
}

// errServerClosedStream signals that the server ended the upload stream before
// the client finished sending; the final status comes from CloseAndRecv.
var errServerClosedStream = errors.New("server closed upload stream early")

// sendAll streams init + data chunks. It returns nil when the reader is
// drained, errServerClosedStream when the server closed the stream early, or
// the failing read/send error otherwise.
func (c *Client) sendAll(stream xaiv1.Files_UploadFileClient, name string, r io.Reader, size int64, cfg *uploadCfg) error {
	send := func(chunk *xaiv1.UploadFileChunk) error {
		err := stream.Send(chunk)
		if err == io.EOF {
			return errServerClosedStream
		}
		return err
	}
	init := &xaiv1.UploadFileInit{Name: name}
	if cfg.expiresAfter != nil {
		init.ExpiresAfter = cfg.expiresAfter
	}
	if err := send(&xaiv1.UploadFileChunk{Chunk: &xaiv1.UploadFileChunk_Init{Init: init}}); err != nil {
		return err
	}
	buf := make([]byte, chunkSize)
	for {
		n, readErr := r.Read(buf)
		if n > 0 {
			data := make([]byte, n)
			copy(data, buf[:n])
			if err := send(&xaiv1.UploadFileChunk{Chunk: &xaiv1.UploadFileChunk_Data{Data: data}}); err != nil {
				return err
			}
			if cfg.progress != nil {
				cfg.progress(int64(n), size)
			}
		}
		if readErr == io.EOF {
			return nil
		}
		if readErr != nil {
			return fmt.Errorf("read upload source: %w", readErr)
		}
	}
}

// UploadPath uploads a local path.
func (c *Client) UploadPath(ctx context.Context, path string, opts ...UploadOption) (*xaiv1.File, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	return c.Upload(ctx, filepath.Base(path), f, st.Size(), opts...)
}

// ListOption configures listing.
type ListOption func(*xaiv1.ListFilesRequest)

// WithListLimit sets the page size (server clamps to [1, 100]).
func WithListLimit(n int32) ListOption {
	return func(r *xaiv1.ListFilesRequest) { r.Limit = n }
}

// WithListPaginationToken resumes listing from a prior page token.
func WithListPaginationToken(t string) ListOption {
	return func(r *xaiv1.ListFilesRequest) { r.PaginationToken = &t }
}

// WithListOrder sets ascending (true) or descending (false) order.
func WithListOrder(asc bool) ListOption {
	return func(r *xaiv1.ListFilesRequest) {
		if asc {
			r.Order = xaiv1.Ordering_ASCENDING
		} else {
			r.Order = xaiv1.Ordering_DESCENDING
		}
	}
}

// WithListSortBy sets the sort field: "created_at" (default), "filename", or "size".
// Unknown values fall back to "created_at" (the option type cannot return an error).
func WithListSortBy(sortBy string) ListOption {
	return func(r *xaiv1.ListFilesRequest) {
		var s xaiv1.FilesSortBy
		switch sortBy {
		case "filename", "name":
			s = xaiv1.FilesSortBy_FILES_SORT_BY_FILENAME
		case "size", "size_bytes":
			s = xaiv1.FilesSortBy_FILES_SORT_BY_SIZE
		default:
			s = xaiv1.FilesSortBy_FILES_SORT_BY_CREATED_AT
		}
		r.SortBy = &s
	}
}

// WithListFilter sets a filter expression (e.g. content_type = "application/pdf").
// Wired to ListFilesRequest.filter when the field is available.
func WithListFilter(filter string) ListOption {
	return func(r *xaiv1.ListFilesRequest) {
		if filter != "" {
			r.Filter = &filter
		}
	}
}

// List returns file metadata pages.
func (c *Client) List(ctx context.Context, opts ...ListOption) (*xaiv1.ListFilesResponse, error) {
	req := &xaiv1.ListFilesRequest{}
	for _, o := range opts {
		o(req)
	}
	return c.stub.ListFiles(ctx, req)
}

// Get retrieves file metadata.
func (c *Client) Get(ctx context.Context, fileID string) (*xaiv1.File, error) {
	return c.stub.RetrieveFile(ctx, &xaiv1.RetrieveFileRequest{FileId: fileID})
}

// Delete deletes a file.
func (c *Client) Delete(ctx context.Context, fileID string) (*xaiv1.DeleteFileResponse, error) {
	return c.stub.DeleteFile(ctx, &xaiv1.DeleteFileRequest{FileId: fileID})
}

// Content downloads file bytes into memory, capped at MaxContentBytes (512 MiB).
// For unbounded/large objects use ContentWriter. On error no partial data is
// returned.
func (c *Client) Content(ctx context.Context, fileID string) ([]byte, error) {
	var buf []byte
	err := c.ContentWriter(ctx, fileID, writerFunc(func(p []byte) (int, error) {
		if int64(len(buf))+int64(len(p)) > MaxContentBytes {
			return 0, fmt.Errorf("file content exceeds max download size of %d bytes", MaxContentBytes)
		}
		buf = append(buf, p...)
		return len(p), nil
	}))
	if err != nil {
		return nil, err
	}
	return buf, nil
}

// writerFunc adapts a function to io.Writer.
type writerFunc func([]byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }

// ContentWriter streams file content chunks to w without buffering the full file
// in the SDK (FILE-04). Callers own backpressure and total size limits.
func (c *Client) ContentWriter(ctx context.Context, fileID string, w io.Writer) error {
	if w == nil {
		return fmt.Errorf("nil writer")
	}
	// Child context so early returns (e.g. writer errors) cancel the server
	// stream promptly instead of holding it until the caller's ctx ends.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream, err := c.stub.RetrieveFileContent(ctx, &xaiv1.RetrieveFileContentRequest{FileId: fileID})
	if err != nil {
		return err
	}
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		data := chunk.GetData()
		if len(data) == 0 {
			continue
		}
		if _, err := w.Write(data); err != nil {
			return err
		}
	}
}

// CreatePublicURL creates a public URL for a file.
// expiresAfter is optional; when set, it must be at least one second (the wire
// value has seconds granularity) and the public URL expires after that duration.
func (c *Client) CreatePublicURL(ctx context.Context, fileID string, expiresAfter *time.Duration) (*xaiv1.CreatePublicUrlResponse, error) {
	req := &xaiv1.CreatePublicUrlRequest{FileId: fileID}
	if expiresAfter != nil {
		s := int64(expiresAfter.Seconds())
		if s <= 0 {
			return nil, fmt.Errorf("files: public URL expires_after must be at least 1s, got %s", *expiresAfter)
		}
		req.ExpiresAfter = &s
	}
	return c.stub.CreatePublicUrl(ctx, req)
}

// RevokePublicURL revokes a public URL.
func (c *Client) RevokePublicURL(ctx context.Context, fileID string) (*xaiv1.RevokePublicUrlResponse, error) {
	return c.stub.RevokePublicUrl(ctx, &xaiv1.RevokePublicUrlRequest{FileId: fileID})
}

// BatchUploadCallback is invoked after each item finishes (success or error).
// path is the source path or synthetic name for reader items.
type BatchUploadCallback func(index int, path string, file *xaiv1.File, err error)

// BatchItem is a path or in-memory reader for batch upload (FILE-01).
// Exactly one of Path or Reader must be set; setting both (or neither) yields
// an error for that item. When Reader is set, Name is required and Size should
// be the known size (or -1 if unknown; Upload may still work for some backends).
type BatchItem struct {
	Path   string
	Reader io.Reader
	Name   string
	Size   int64
}

// BatchUploadOption configures BatchUpload.
type BatchUploadOption func(*batchCfg)
type batchCfg struct {
	uploadOpts  []UploadOption
	onComplete  BatchUploadCallback
	concurrency int
}

// WithBatchConcurrency sets max concurrent uploads (default 4).
func WithBatchConcurrency(n int) BatchUploadOption {
	return func(c *batchCfg) { c.concurrency = n }
}

// WithBatchOnComplete registers a per-file completion callback.
func WithBatchOnComplete(fn BatchUploadCallback) BatchUploadOption {
	return func(c *batchCfg) { c.onComplete = fn }
}

// WithBatchUploadOptions applies UploadOptions to every file in the batch.
func WithBatchUploadOptions(opts ...UploadOption) BatchUploadOption {
	return func(c *batchCfg) { c.uploadOpts = append(c.uploadOpts, opts...) }
}

// BatchUpload uploads multiple local paths with limited concurrency.
// Returns parallel slices of files and errors (same length as paths).
// Prefer BatchUploadWithOptions for callbacks; this signature keeps the
// legacy concurrency int for call-site compatibility.
func (c *Client) BatchUpload(ctx context.Context, paths []string, concurrency int, opts ...UploadOption) ([]*xaiv1.File, []error) {
	return c.BatchUploadWithOptions(ctx, paths,
		WithBatchConcurrency(concurrency),
		WithBatchUploadOptions(opts...),
	)
}

// BatchUploadWithOptions uploads paths with options (concurrency, callback).
func (c *Client) BatchUploadWithOptions(ctx context.Context, paths []string, opts ...BatchUploadOption) ([]*xaiv1.File, []error) {
	items := make([]BatchItem, len(paths))
	for i, p := range paths {
		items[i] = BatchItem{Path: p}
	}
	return c.BatchUploadItems(ctx, items, opts...)
}

// BatchUploadItems uploads a mix of paths and in-memory readers (FILE-01).
//
// At most concurrency (default 4) worker goroutines run; items are processed
// in order of submission. A panic from an item upload or the onComplete
// callback is recovered into that item's error instead of crashing the
// process (the panic happens on an SDK-owned goroutine the caller could not
// otherwise recover from).
func (c *Client) BatchUploadItems(ctx context.Context, items []BatchItem, opts ...BatchUploadOption) ([]*xaiv1.File, []error) {
	cfg := &batchCfg{concurrency: 4}
	for _, o := range opts {
		o(cfg)
	}
	if cfg.concurrency < 1 {
		cfg.concurrency = 4
	}
	filesOut := make([]*xaiv1.File, len(items))
	errs := make([]error, len(items))
	workers := cfg.concurrency
	if workers > len(items) {
		workers = len(items)
	}
	idxCh := make(chan int)
	var wg sync.WaitGroup
	var cbMu sync.Mutex
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range idxCh {
				label, f, err := c.uploadBatchItem(ctx, i, items[i], cfg.uploadOpts)
				if cfg.onComplete != nil {
					if cbErr := runBatchCallback(&cbMu, cfg.onComplete, i, label, f, err); cbErr != nil {
						err = errors.Join(err, cbErr)
					}
				}
				filesOut[i], errs[i] = f, err
			}
		}()
	}
	for i := range items {
		idxCh <- i
	}
	close(idxCh)
	wg.Wait()
	return filesOut, errs
}

// uploadBatchItem uploads one batch item, converting panics into errors.
func (c *Client) uploadBatchItem(ctx context.Context, i int, it BatchItem, uploadOpts []UploadOption) (label string, f *xaiv1.File, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("batch item %d: panic during upload: %v", i, r)
		}
	}()
	label = it.Path
	switch {
	case it.Path != "" && it.Reader != nil:
		err = fmt.Errorf("batch item %d: exactly one of Path or Reader must be set", i)
	case it.Path != "":
		f, err = c.UploadPath(ctx, it.Path, uploadOpts...)
	case it.Reader != nil:
		if it.Name == "" {
			err = fmt.Errorf("batch item %d: Name required when using Reader", i)
		} else {
			label = it.Name
			f, err = c.Upload(ctx, it.Name, it.Reader, it.Size, uploadOpts...)
		}
	default:
		err = fmt.Errorf("batch item %d: Path or Reader required", i)
	}
	return label, f, err
}

// runBatchCallback invokes the user callback under the shared mutex,
// converting a callback panic into an error for that item.
func runBatchCallback(mu *sync.Mutex, cb BatchUploadCallback, i int, label string, f *xaiv1.File, err error) (cbErr error) {
	mu.Lock()
	defer mu.Unlock()
	defer func() {
		if r := recover(); r != nil {
			cbErr = fmt.Errorf("batch item %d: onComplete callback panicked: %v", i, r)
		}
	}()
	cb(i, label, f, err)
	return nil
}
