// Package collections implements Collections management and document search.
//
// CRUD and document indexing require a management API key (management channel).
// Search uses the business API Documents service.
package collections

import (
	"context"
	"errors"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"github.com/fun7257/xai-sdk-go/files"
	"github.com/fun7257/xai-sdk-go/internal/poll"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// ErrNoManagementKey is returned when Collections CRUD is used without a
// management connection / XAI_MANAGEMENT_KEY. Re-exported as xai.ErrNoManagementKey.
var ErrNoManagementKey = errors.New("xai: management API key not provided")

// Client manages collections (management channel required for CRUD).
type Client struct {
	mgmt    xaiv1.CollectionsClient
	docs    xaiv1.DocumentsClient
	files   *files.Client
	hasMgmt bool
}

// New creates a Collections client. management may be nil (CRUD will error).
// The api connection is used for Documents.Search; files uploads use api as well.
func New(api, management grpc.ClientConnInterface) *Client {
	c := &Client{}
	if api != nil {
		c.docs = xaiv1.NewDocumentsClient(api)
		c.files = files.New(api)
	}
	if management != nil {
		c.mgmt = xaiv1.NewCollectionsClient(management)
		c.hasMgmt = true
	}
	return c
}

func (c *Client) requireMgmt() error {
	if !c.hasMgmt || c.mgmt == nil {
		return fmt.Errorf("please provide a management API key: %w", ErrNoManagementKey)
	}
	return nil
}

// Create creates a collection.
func (c *Client) Create(ctx context.Context, name string, opts ...CreateOption) (*xaiv1.CollectionMetadata, error) {
	if err := c.requireMgmt(); err != nil {
		return nil, err
	}
	cfg := &createCfg{}
	for _, o := range opts {
		o(cfg)
	}
	if err := ValidateChunkConfiguration(cfg.chunk); err != nil {
		return nil, err
	}
	metric, err := hnswMetric(cfg.metricStr, cfg.metricSet)
	if err != nil {
		return nil, err
	}
	return c.mgmt.CreateCollection(ctx, &xaiv1.CreateCollectionRequest{
		CollectionName:        name,
		CollectionDescription: cfg.description,
		IndexConfiguration:    cfg.index,
		ChunkConfiguration:    cfg.chunk,
		MetricSpace:           metric,
		FieldDefinitions:      cfg.fields,
	})
}

// CreateOption configures Create.
type CreateOption func(*createCfg)
type createCfg struct {
	description string
	index       *xaiv1.IndexConfiguration
	chunk       *xaiv1.ChunkConfiguration
	metricStr   string
	metricSet   bool
	fields      []*xaiv1.FieldDefinition
}

// WithDescription sets the collection description.
func WithDescription(s string) CreateOption { return func(c *createCfg) { c.description = s } }

// WithIndexModel sets the embedding model name.
func WithIndexModel(model string) CreateOption {
	return func(c *createCfg) { c.index = &xaiv1.IndexConfiguration{ModelName: model} }
}

// WithChunkConfiguration sets chunking config (exactly one of chars/tokens/bytes).
func WithChunkConfiguration(ch *xaiv1.ChunkConfiguration) CreateOption {
	return func(c *createCfg) { c.chunk = ch }
}

// WithMetric sets HNSW metric: "cosine" (default), "euclidean", or "inner_product".
// Unknown values cause Create to return an error (no silent fallback).
func WithMetric(m string) CreateOption {
	return func(c *createCfg) {
		c.metricStr = m
		c.metricSet = true
	}
}

func hnswMetric(s string, set bool) (xaiv1.HNSWMetric, error) {
	if !set {
		return xaiv1.HNSWMetric_HNSW_METRIC_UNKNOWN, nil
	}
	switch s {
	case "", "cosine":
		return xaiv1.HNSWMetric_HNSW_METRIC_COSINE, nil
	case "euclidean":
		return xaiv1.HNSWMetric_HNSW_METRIC_EUCLIDEAN, nil
	case "inner_product":
		return xaiv1.HNSWMetric_HNSW_METRIC_INNER_PRODUCT, nil
	default:
		return 0, fmt.Errorf("collections: unknown metric %q (want cosine, euclidean, or inner_product)", s)
	}
}

// WithFieldDefinitions sets initial field definitions.
func WithFieldDefinitions(fs ...*xaiv1.FieldDefinition) CreateOption {
	return func(c *createCfg) { c.fields = fs }
}

// ValidateChunkConfiguration requires exactly one of chars/tokens/bytes when cfg is non-nil.
// When cfg is nil, returns nil.
func ValidateChunkConfiguration(cfg *xaiv1.ChunkConfiguration) error {
	if cfg == nil {
		return nil
	}
	n := 0
	if cfg.CharsConfiguration != nil {
		n++
	}
	if cfg.TokensConfiguration != nil {
		n++
	}
	if cfg.BytesConfiguration != nil {
		n++
	}
	if n == 0 {
		return fmt.Errorf("chunk configuration requires exactly one of chars, tokens, or bytes configuration")
	}
	if n > 1 {
		return fmt.Errorf("chunk configuration allows only one of chars, tokens, or bytes configuration")
	}
	return nil
}

// FieldDefinition builds a field definition.
func FieldDefinition(key string, required, injectIntoChunk, unique bool, description string) *xaiv1.FieldDefinition {
	return &xaiv1.FieldDefinition{
		Key: key, Required: required, InjectIntoChunk: injectIntoChunk,
		Unique: unique, Description: description,
	}
}

// AddFieldDefinition builds a FieldDefinitionUpdate with ADD operation.
func AddFieldDefinition(fd *xaiv1.FieldDefinition) *xaiv1.FieldDefinitionUpdate {
	return &xaiv1.FieldDefinitionUpdate{
		FieldDefinition: fd,
		Operation:       xaiv1.FieldDefinitionOperation_FIELD_DEFINITION_ADD,
	}
}

// DeleteFieldDefinition builds a FieldDefinitionUpdate with DELETE operation.
func DeleteFieldDefinition(key string) *xaiv1.FieldDefinitionUpdate {
	return &xaiv1.FieldDefinitionUpdate{
		FieldDefinition: &xaiv1.FieldDefinition{Key: key},
		Operation:       xaiv1.FieldDefinitionOperation_FIELD_DEFINITION_DELETE,
	}
}

// ListOption configures ListCollections.
type ListOption func(*xaiv1.ListCollectionsRequest)

// WithListLimit sets page size.
func WithListLimit(n int32) ListOption {
	return func(r *xaiv1.ListCollectionsRequest) { r.Limit = n }
}

// WithListFilter sets a filter expression.
func WithListFilter(filter string) ListOption {
	return func(r *xaiv1.ListCollectionsRequest) { r.Filter = filter }
}

// WithListPaginationToken resumes listing.
func WithListPaginationToken(t string) ListOption {
	return func(r *xaiv1.ListCollectionsRequest) { r.PaginationToken = t }
}

// List lists collections (management key required).
func (c *Client) List(ctx context.Context, opts ...ListOption) (*xaiv1.ListCollectionsResponse, error) {
	if err := c.requireMgmt(); err != nil {
		return nil, err
	}
	req := &xaiv1.ListCollectionsRequest{}
	for _, o := range opts {
		o(req)
	}
	return c.mgmt.ListCollections(ctx, req)
}

// Get returns collection metadata.
func (c *Client) Get(ctx context.Context, collectionID string) (*xaiv1.CollectionMetadata, error) {
	if err := c.requireMgmt(); err != nil {
		return nil, err
	}
	return c.mgmt.GetCollectionMetadata(ctx, &xaiv1.GetCollectionMetadataRequest{CollectionId: collectionID})
}

// Delete deletes a collection.
func (c *Client) Delete(ctx context.Context, collectionID string) error {
	if err := c.requireMgmt(); err != nil {
		return err
	}
	_, err := c.mgmt.DeleteCollection(ctx, &xaiv1.DeleteCollectionRequest{CollectionId: collectionID})
	return err
}

// Update updates a collection.
func (c *Client) Update(ctx context.Context, collectionID string, req *xaiv1.UpdateCollectionRequest) (*xaiv1.CollectionMetadata, error) {
	if err := c.requireMgmt(); err != nil {
		return nil, err
	}
	if req == nil {
		req = &xaiv1.UpdateCollectionRequest{}
	}
	if err := ValidateChunkConfiguration(req.ChunkConfiguration); err != nil {
		return nil, err
	}
	req.CollectionId = collectionID
	return c.mgmt.UpdateCollection(ctx, req)
}

// GenerateDescription generates a description for a collection.
func (c *Client) GenerateDescription(ctx context.Context, collectionID string) (string, error) {
	if err := c.requireMgmt(); err != nil {
		return "", err
	}
	r, err := c.mgmt.GenerateCollectionDescription(ctx, &xaiv1.GenerateCollectionDescriptionRequest{CollectionId: collectionID})
	if err != nil {
		return "", err
	}
	return r.Description, nil
}

// AddExistingDocument adds an already-uploaded file to a collection.
func (c *Client) AddExistingDocument(ctx context.Context, collectionID, fileID string, fields map[string]string) error {
	if err := c.requireMgmt(); err != nil {
		return err
	}
	_, err := c.mgmt.AddDocumentToCollection(ctx, &xaiv1.AddDocumentToCollectionRequest{
		CollectionId: collectionID, FileId: fileID, Fields: fields,
	})
	return err
}

// UploadDocumentOption configures UploadDocument.
type UploadDocumentOption func(*uploadDocCfg)
type uploadDocCfg struct {
	fields             map[string]string
	wait               bool
	waitTimeout        time.Duration
	waitInterval       time.Duration
	uploadOpts         []files.UploadOption
	deleteOnAddFailure bool
}

// WithDocumentFields sets metadata fields on the document.
func WithDocumentFields(fields map[string]string) UploadDocumentOption {
	return func(c *uploadDocCfg) { c.fields = fields }
}

// WithWaitForIndexing waits until the document is processed (or fails).
func WithWaitForIndexing(timeout, interval time.Duration) UploadDocumentOption {
	return func(c *uploadDocCfg) {
		c.wait = true
		c.waitTimeout = timeout
		c.waitInterval = interval
	}
}

// WithUploadOptions applies Files upload options.
func WithUploadOptions(opts ...files.UploadOption) UploadDocumentOption {
	return func(c *uploadDocCfg) { c.uploadOpts = opts }
}

// WithDeleteOnAddFailure deletes the uploaded Files object if AddExistingDocument
// fails after a successful upload (BHV-04). Default is to leave the orphan file.
func WithDeleteOnAddFailure() UploadDocumentOption {
	return func(c *uploadDocCfg) { c.deleteOnAddFailure = true }
}

// UploadDocument uploads a local file via Files API, adds it to the collection,
// and optionally waits for indexing.
//
// On AddExistingDocument failure after a successful upload, the uploaded file is
// left in place by default (returned so callers can Delete). Pass
// WithDeleteOnAddFailure to automatically delete the orphaned file (best-effort).
// Indexing wait failures leave file and collection membership as-is after a
// successful add.
func (c *Client) UploadDocument(ctx context.Context, collectionID, path string, opts ...UploadDocumentOption) (*xaiv1.File, *xaiv1.DocumentMetadata, error) {
	if err := c.requireMgmt(); err != nil {
		return nil, nil, err
	}
	if c.files == nil {
		return nil, nil, fmt.Errorf("files client unavailable")
	}
	cfg := &uploadDocCfg{}
	for _, o := range opts {
		o(cfg)
	}
	f, err := c.files.UploadPath(ctx, path, cfg.uploadOpts...)
	if err != nil {
		return nil, nil, err
	}
	if err := c.AddExistingDocument(ctx, collectionID, f.Id, cfg.fields); err != nil {
		if cfg.deleteOnAddFailure {
			// Detach from ctx so cleanup still runs when the add failed due to
			// cancellation/deadline; bound it so it cannot hang.
			cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
			defer cancel()
			_, _ = c.files.Delete(cleanupCtx, f.Id)
		}
		return f, nil, err
	}
	if !cfg.wait {
		return f, nil, nil
	}
	doc, err := c.WaitForIndexing(ctx, collectionID, f.Id, cfg.waitTimeout, cfg.waitInterval)
	return f, doc, err
}

// RemoveDocument removes a document from a collection.
func (c *Client) RemoveDocument(ctx context.Context, collectionID, fileID string) error {
	if err := c.requireMgmt(); err != nil {
		return err
	}
	_, err := c.mgmt.RemoveDocumentFromCollection(ctx, &xaiv1.RemoveDocumentFromCollectionRequest{
		CollectionId: collectionID, FileId: fileID,
	})
	return err
}

// UpdateDocumentOption configures UpdateDocument (COL-01).
type UpdateDocumentOption func(*updateDocCfg)
type updateDocCfg struct {
	name        string
	data        []byte
	contentType string
	fields      map[string]string
}

// WithUpdateName sets the document display name.
func WithUpdateName(name string) UpdateDocumentOption {
	return func(c *updateDocCfg) { c.name = name }
}

// WithUpdateData replaces document bytes.
func WithUpdateData(data []byte) UpdateDocumentOption {
	return func(c *updateDocCfg) { c.data = data }
}

// WithUpdateContentType sets MIME content type.
func WithUpdateContentType(ct string) UpdateDocumentOption {
	return func(c *updateDocCfg) { c.contentType = ct }
}

// WithUpdateFields sets/replaces metadata fields.
func WithUpdateFields(fields map[string]string) UpdateDocumentOption {
	return func(c *updateDocCfg) { c.fields = fields }
}

// UpdateDocument updates a document's data and/or metadata (management API).
func (c *Client) UpdateDocument(ctx context.Context, collectionID, fileID string, opts ...UpdateDocumentOption) (*xaiv1.DocumentMetadata, error) {
	if err := c.requireMgmt(); err != nil {
		return nil, err
	}
	cfg := &updateDocCfg{}
	for _, o := range opts {
		o(cfg)
	}
	return c.mgmt.UpdateDocument(ctx, &xaiv1.UpdateDocumentRequest{
		CollectionId: collectionID,
		FileId:       fileID,
		Name:         cfg.name,
		Data:         cfg.data,
		ContentType:  cfg.contentType,
		Fields:       cfg.fields,
	})
}

// GetDocument returns document metadata.
func (c *Client) GetDocument(ctx context.Context, collectionID, fileID string) (*xaiv1.DocumentMetadata, error) {
	if err := c.requireMgmt(); err != nil {
		return nil, err
	}
	return c.mgmt.GetDocumentMetadata(ctx, &xaiv1.GetDocumentMetadataRequest{
		CollectionId: collectionID, FileId: fileID,
	})
}

// ListDocumentsOption configures ListDocuments.
type ListDocumentsOption func(*xaiv1.ListDocumentsRequest)

// WithDocumentsFilter sets a filter expression.
func WithDocumentsFilter(filter string) ListDocumentsOption {
	return func(r *xaiv1.ListDocumentsRequest) { r.Filter = filter }
}

// WithDocumentsName filters by document name.
func WithDocumentsName(name string) ListDocumentsOption {
	return func(r *xaiv1.ListDocumentsRequest) { r.Name = name }
}

// WithDocumentsLimit sets page size.
func WithDocumentsLimit(n int32) ListDocumentsOption {
	return func(r *xaiv1.ListDocumentsRequest) { r.Limit = n }
}

// ListDocuments lists documents in a collection.
func (c *Client) ListDocuments(ctx context.Context, collectionID string, opts ...ListDocumentsOption) (*xaiv1.ListDocumentsResponse, error) {
	if err := c.requireMgmt(); err != nil {
		return nil, err
	}
	req := &xaiv1.ListDocumentsRequest{CollectionId: collectionID}
	for _, o := range opts {
		o(req)
	}
	return c.mgmt.ListDocuments(ctx, req)
}

// BatchGetDocuments retrieves multiple documents by id.
func (c *Client) BatchGetDocuments(ctx context.Context, collectionID string, fileIDs []string) (*xaiv1.BatchGetDocumentsResponse, error) {
	if err := c.requireMgmt(); err != nil {
		return nil, err
	}
	return c.mgmt.BatchGetDocuments(ctx, &xaiv1.BatchGetDocumentsRequest{
		CollectionId: collectionID, FileIds: fileIDs,
	})
}

// ReindexDocument re-indexes a document.
func (c *Client) ReindexDocument(ctx context.Context, collectionID, fileID string) error {
	if err := c.requireMgmt(); err != nil {
		return err
	}
	_, err := c.mgmt.ReIndexDocument(ctx, &xaiv1.ReIndexDocumentRequest{
		CollectionId: collectionID, FileId: fileID,
	})
	return err
}

// SearchOption configures document Search.
type SearchOption func(*searchCfg)
type searchCfg struct {
	limit        *int32
	instructions string
	// retrieval: hybrid (default), semantic, keyword
	mode string
}

// WithSearchLimit sets the max number of chunks returned.
func WithSearchLimit(n int32) SearchOption {
	return func(c *searchCfg) { c.limit = &n }
}

// WithSearchInstructions sets user-defined search instructions.
func WithSearchInstructions(s string) SearchOption {
	return func(c *searchCfg) { c.instructions = s }
}

// WithRetrievalMode sets retrieval mode: "hybrid" (default), "semantic", or "keyword".
func WithRetrievalMode(mode string) SearchOption {
	return func(c *searchCfg) { c.mode = mode }
}

// Search searches documents across collections via the Documents API.
func (c *Client) Search(ctx context.Context, query string, collectionIDs []string, opts ...SearchOption) (*xaiv1.SearchResponse, error) {
	if c.docs == nil {
		return nil, fmt.Errorf("documents client unavailable")
	}
	cfg := &searchCfg{}
	for _, o := range opts {
		o(cfg)
	}
	req := &xaiv1.SearchRequest{
		Query:  query,
		Source: &xaiv1.DocumentsSource{CollectionIds: collectionIDs},
	}
	if cfg.limit != nil {
		req.Limit = cfg.limit
	}
	if cfg.instructions != "" {
		req.Instructions = &cfg.instructions
	}
	switch cfg.mode {
	case "semantic":
		req.RetrievalMode = &xaiv1.SearchRequest_SemanticRetrieval{SemanticRetrieval: &xaiv1.SemanticRetrieval{}}
	case "keyword":
		req.RetrievalMode = &xaiv1.SearchRequest_KeywordRetrieval{KeywordRetrieval: &xaiv1.KeywordRetrieval{}}
	case "hybrid", "":
		if cfg.mode == "hybrid" {
			req.RetrievalMode = &xaiv1.SearchRequest_HybridRetrieval{HybridRetrieval: &xaiv1.HybridRetrieval{}}
		}
	default:
		return nil, fmt.Errorf("unknown retrieval_mode %q (want hybrid, semantic, or keyword)", cfg.mode)
	}
	// Legacy signature convenience: Search(ctx, query, ids, limit, instructions)
	return c.docs.Search(ctx, req)
}

// SearchSimple is a compatibility wrapper matching the prior limit/instructions signature.
func (c *Client) SearchSimple(ctx context.Context, query string, collectionIDs []string, limit *int32, instructions string) (*xaiv1.SearchResponse, error) {
	var opts []SearchOption
	if limit != nil {
		opts = append(opts, WithSearchLimit(*limit))
	}
	if instructions != "" {
		opts = append(opts, WithSearchInstructions(instructions))
	}
	return c.Search(ctx, query, collectionIDs, opts...)
}

// WaitForIndexing polls until a document is processed or failed.
func (c *Client) WaitForIndexing(ctx context.Context, collectionID, fileID string, timeout, interval time.Duration) (*xaiv1.DocumentMetadata, error) {
	if timeout <= 0 {
		timeout = 2 * time.Minute
	}
	if interval <= 0 {
		interval = 10 * time.Second
	}
	var doc *xaiv1.DocumentMetadata
	err := poll.Wait(ctx, poll.Config{Timeout: timeout, Interval: interval, Context: "waiting for document to be indexed"}, func(ctx context.Context) (bool, error) {
		d, err := c.GetDocument(ctx, collectionID, fileID)
		if err != nil {
			return false, err
		}
		doc = d
		switch d.Status {
		case xaiv1.DocumentStatus_DOCUMENT_STATUS_PROCESSED:
			return true, nil
		case xaiv1.DocumentStatus_DOCUMENT_STATUS_FAILED:
			msg := d.ErrorMessage
			if msg == "" {
				msg = "document indexing failed"
			}
			return false, fmt.Errorf("%s", msg)
		case xaiv1.DocumentStatus_DOCUMENT_STATUS_PROCESSING,
			xaiv1.DocumentStatus_DOCUMENT_STATUS_CHUNKED,
			xaiv1.DocumentStatus_DOCUMENT_STATUS_EMBEDDING,
			xaiv1.DocumentStatus_DOCUMENT_STATUS_WRITING,
			xaiv1.DocumentStatus_DOCUMENT_STATUS_UNKNOWN:
			return false, nil
		default:
			return false, nil
		}
	})
	return doc, err
}
