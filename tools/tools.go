// Package tools builds server-side and client-side tools for chat.
//
// Primary constructors (WebSearch, XSearch) validate mutual exclusion and
// return errors. Use Unchecked* only when you intentionally skip validation.
package tools

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

func bp(v bool) *bool { return &v }
func sp(v string) *string {
	if v == "" {
		return nil
	}
	return &v
}

// Function creates a client-side function tool.
func Function(name, description string, parameters any) (*xaiv1.Tool, error) {
	var raw []byte
	var err error
	switch p := parameters.(type) {
	case string:
		raw = []byte(p)
	case []byte:
		raw = p
	case json.RawMessage:
		raw = p
	default:
		raw, err = json.Marshal(p)
		if err != nil {
			return nil, err
		}
	}
	return &xaiv1.Tool{Tool: &xaiv1.Tool_Function{Function: &xaiv1.Function{
		Name: name, Description: description, Parameters: string(raw),
	}}}, nil
}

// RequiredTool forces a function name.
func RequiredTool(name string) *xaiv1.ToolChoice {
	return &xaiv1.ToolChoice{ToolChoice: &xaiv1.ToolChoice_FunctionName{FunctionName: name}}
}

// Mode returns a tool choice mode string (auto/none/required).
func Mode(mode string) *xaiv1.ToolChoice {
	var m xaiv1.ToolMode
	switch mode {
	case "none":
		m = xaiv1.ToolMode_TOOL_MODE_NONE
	case "required":
		m = xaiv1.ToolMode_TOOL_MODE_REQUIRED
	default:
		m = xaiv1.ToolMode_TOOL_MODE_AUTO
	}
	return &xaiv1.ToolChoice{ToolChoice: &xaiv1.ToolChoice_Mode{Mode: m}}
}

// WebSearchOption configures WebSearch.
type WebSearchOption func(*webSearchCfg)
type webSearchCfg struct {
	excluded, allowed               []string
	imageUnderstanding, imageSearch bool
	country, city, region, tz       string
}

// WebSearch creates a server-side web search tool (primary path).
// Returns an error if allowed and excluded domains are both set.
func WebSearch(opts ...WebSearchOption) (*xaiv1.Tool, error) {
	cfg := &webSearchCfg{}
	for _, o := range opts {
		o(cfg)
	}
	if err := ValidateWebSearchDomains(cfg.allowed, cfg.excluded); err != nil {
		return nil, err
	}
	return buildWebSearch(cfg), nil
}

// UncheckedWebSearch builds a WebSearch tool without domain validation (escape hatch).
func UncheckedWebSearch(opts ...WebSearchOption) *xaiv1.Tool {
	cfg := &webSearchCfg{}
	for _, o := range opts {
		o(cfg)
	}
	return buildWebSearch(cfg)
}

// WebSearchChecked is a deprecated alias of WebSearch.
func WebSearchChecked(opts ...WebSearchOption) (*xaiv1.Tool, error) {
	return WebSearch(opts...)
}

// ValidateWebSearchDomains returns an error when both allowed and excluded are set.
func ValidateWebSearchDomains(allowed, excluded []string) error {
	if len(allowed) > 0 && len(excluded) > 0 {
		return fmt.Errorf("web search: allowed_domains and excluded_domains are mutually exclusive")
	}
	return nil
}

func buildWebSearch(cfg *webSearchCfg) *xaiv1.Tool {
	ws := &xaiv1.WebSearch{ExcludedDomains: cfg.excluded, AllowedDomains: cfg.allowed}
	if cfg.imageUnderstanding {
		ws.EnableImageUnderstanding = bp(true)
	}
	if cfg.imageSearch {
		ws.EnableImageSearch = bp(true)
	}
	if cfg.country != "" || cfg.city != "" || cfg.region != "" || cfg.tz != "" {
		ws.UserLocation = &xaiv1.WebSearchUserLocation{
			Country: sp(cfg.country), City: sp(cfg.city), Region: sp(cfg.region), Timezone: sp(cfg.tz),
		}
	}
	return &xaiv1.Tool{Tool: &xaiv1.Tool_WebSearch{WebSearch: ws}}
}

// WithExcludedDomains restricts web search exclusions.
func WithExcludedDomains(d ...string) WebSearchOption {
	return func(c *webSearchCfg) { c.excluded = d }
}

// WithAllowedDomains restricts web search to domains.
func WithAllowedDomains(d ...string) WebSearchOption {
	return func(c *webSearchCfg) { c.allowed = d }
}

// WithImageUnderstanding enables image understanding.
func WithImageUnderstanding(v bool) WebSearchOption {
	return func(c *webSearchCfg) { c.imageUnderstanding = v }
}

// WithImageSearch enables image search results.
func WithImageSearch(v bool) WebSearchOption {
	return func(c *webSearchCfg) { c.imageSearch = v }
}

// WithUserLocation sets geolocation preference.
func WithUserLocation(country, city, region, timezone string) WebSearchOption {
	return func(c *webSearchCfg) { c.country, c.city, c.region, c.tz = country, city, region, timezone }
}

// XSearchOption configures XSearch.
type XSearchOption func(*xSearchCfg)
type xSearchCfg struct {
	from, to          *time.Time
	allowed, excluded []string
	img, vid          bool
}

// XSearch creates a server-side X search tool (primary path).
// Returns an error if allowed and excluded handles are both set.
func XSearch(opts ...XSearchOption) (*xaiv1.Tool, error) {
	cfg := &xSearchCfg{}
	for _, o := range opts {
		o(cfg)
	}
	if err := ValidateXSearchHandles(cfg.allowed, cfg.excluded); err != nil {
		return nil, err
	}
	return buildXSearch(cfg), nil
}

// UncheckedXSearch builds an XSearch tool without handle validation (escape hatch).
func UncheckedXSearch(opts ...XSearchOption) *xaiv1.Tool {
	cfg := &xSearchCfg{}
	for _, o := range opts {
		o(cfg)
	}
	return buildXSearch(cfg)
}

// XSearchChecked is a deprecated alias of XSearch.
func XSearchChecked(opts ...XSearchOption) (*xaiv1.Tool, error) {
	return XSearch(opts...)
}

// ValidateXSearchHandles returns an error when both allowed and excluded are set.
func ValidateXSearchHandles(allowed, excluded []string) error {
	if len(allowed) > 0 && len(excluded) > 0 {
		return fmt.Errorf("x search: allowed_x_handles and excluded_x_handles are mutually exclusive")
	}
	return nil
}

func buildXSearch(cfg *xSearchCfg) *xaiv1.Tool {
	xs := &xaiv1.XSearch{AllowedXHandles: cfg.allowed, ExcludedXHandles: cfg.excluded}
	if cfg.img {
		xs.EnableImageUnderstanding = bp(true)
	}
	if cfg.vid {
		xs.EnableVideoUnderstanding = bp(true)
	}
	if cfg.from != nil {
		xs.FromDate = timestamppb.New(*cfg.from)
	}
	if cfg.to != nil {
		xs.ToDate = timestamppb.New(*cfg.to)
	}
	return &xaiv1.Tool{Tool: &xaiv1.Tool_XSearch{XSearch: xs}}
}

// WithXDateRange sets X search date bounds.
func WithXDateRange(from, to *time.Time) XSearchOption {
	return func(c *xSearchCfg) { c.from, c.to = from, to }
}

// WithAllowedXHandles limits X handles.
func WithAllowedXHandles(h ...string) XSearchOption {
	return func(c *xSearchCfg) { c.allowed = h }
}

// WithExcludedXHandles excludes X handles.
func WithExcludedXHandles(h ...string) XSearchOption {
	return func(c *xSearchCfg) { c.excluded = h }
}

// WithXMediaUnderstanding enables image/video understanding for X posts.
func WithXMediaUnderstanding(image, video bool) XSearchOption {
	return func(c *xSearchCfg) { c.img, c.vid = image, video }
}

// CodeExecution creates a server-side code execution tool.
func CodeExecution() *xaiv1.Tool {
	return &xaiv1.Tool{Tool: &xaiv1.Tool_CodeExecution{CodeExecution: &xaiv1.CodeExecution{}}}
}

// CollectionsSearchOption configures CollectionsSearch.
type CollectionsSearchOption func(*csCfg)
type csCfg struct {
	limit        *int32
	instructions string
	mode         string // hybrid|semantic|keyword
}

// WithCollectionsLimit sets the search limit.
func WithCollectionsLimit(n int32) CollectionsSearchOption {
	return func(c *csCfg) { c.limit = &n }
}

// WithCollectionsInstructions sets search instructions.
func WithCollectionsInstructions(s string) CollectionsSearchOption {
	return func(c *csCfg) { c.instructions = s }
}

// WithCollectionsRetrievalMode sets retrieval mode: hybrid, semantic, or keyword.
func WithCollectionsRetrievalMode(mode string) CollectionsSearchOption {
	return func(c *csCfg) { c.mode = mode }
}

// CollectionsSearch creates a server-side collections search tool.
// For retrieval_mode support use CollectionsSearchOpts.
//
// Error is ignored because this path never sets retrieval_mode; invalid modes
// cannot occur here. Prefer CollectionsSearchOpts when you need error returns.
func CollectionsSearch(collectionIDs []string, limit *int32, instructions string) *xaiv1.Tool {
	var opts []CollectionsSearchOption
	if limit != nil {
		opts = append(opts, WithCollectionsLimit(*limit))
	}
	if instructions != "" {
		opts = append(opts, WithCollectionsInstructions(instructions))
	}
	t, err := CollectionsSearchOpts(collectionIDs, opts...)
	if err != nil {
		// Unreachable without WithCollectionsRetrievalMode; keep non-error API.
		return nil
	}
	return t
}

// CollectionsSearchOpts creates a CollectionsSearch tool with options.
// Invalid retrieval_mode returns an error.
func CollectionsSearchOpts(collectionIDs []string, opts ...CollectionsSearchOption) (*xaiv1.Tool, error) {
	cfg := &csCfg{}
	for _, o := range opts {
		o(cfg)
	}
	cs := &xaiv1.CollectionsSearch{CollectionIds: collectionIDs}
	if cfg.limit != nil {
		cs.Limit = cfg.limit
	}
	if cfg.instructions != "" {
		cs.Instructions = &cfg.instructions
	}
	switch cfg.mode {
	case "", "hybrid":
		if cfg.mode == "hybrid" {
			cs.RetrievalMode = &xaiv1.CollectionsSearch_HybridRetrieval{HybridRetrieval: &xaiv1.HybridRetrieval{}}
		}
	case "semantic":
		cs.RetrievalMode = &xaiv1.CollectionsSearch_SemanticRetrieval{SemanticRetrieval: &xaiv1.SemanticRetrieval{}}
	case "keyword":
		cs.RetrievalMode = &xaiv1.CollectionsSearch_KeywordRetrieval{KeywordRetrieval: &xaiv1.KeywordRetrieval{}}
	default:
		return nil, fmt.Errorf("collections search: unknown retrieval_mode %q", cfg.mode)
	}
	return &xaiv1.Tool{Tool: &xaiv1.Tool_CollectionsSearch{CollectionsSearch: cs}}, nil
}

// MCPOption configures MCP.
type MCPOption func(*mcpCfg)
type mcpCfg struct {
	url, label, desc, auth string
	allowed                []string
	headers                map[string]string
}

// MCP creates a remote MCP tool.
func MCP(serverURL string, opts ...MCPOption) *xaiv1.Tool {
	cfg := &mcpCfg{url: serverURL}
	for _, o := range opts {
		o(cfg)
	}
	m := &xaiv1.MCP{
		ServerUrl: cfg.url, ServerLabel: cfg.label, ServerDescription: cfg.desc,
		AllowedToolNames: cfg.allowed, ExtraHeaders: cfg.headers,
	}
	if cfg.auth != "" {
		m.Authorization = &cfg.auth
	}
	return &xaiv1.Tool{Tool: &xaiv1.Tool_Mcp{Mcp: m}}
}

// WithMCPLabel sets MCP label.
func WithMCPLabel(s string) MCPOption { return func(c *mcpCfg) { c.label = s } }

// WithMCPDescription sets MCP description.
func WithMCPDescription(s string) MCPOption { return func(c *mcpCfg) { c.desc = s } }

// WithMCPAuth sets MCP authorization header value.
func WithMCPAuth(s string) MCPOption { return func(c *mcpCfg) { c.auth = s } }

// WithMCPAllowedTools limits MCP tool names.
func WithMCPAllowedTools(n ...string) MCPOption { return func(c *mcpCfg) { c.allowed = n } }

// WithMCPHeaders sets extra headers.
func WithMCPHeaders(h map[string]string) MCPOption { return func(c *mcpCfg) { c.headers = h } }

// CallType returns a short tool call type string.
func CallType(tc *xaiv1.ToolCall) string {
	if tc == nil {
		return ""
	}
	name := tc.GetType().String()
	const p = "TOOL_CALL_TYPE_"
	if strings.HasPrefix(name, p) {
		return strings.ToLower(strings.TrimPrefix(name, p))
	}
	return name
}
