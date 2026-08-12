// Package search builds Live Search parameters and sources for chat.
package search

import (
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"

	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// Mode controls when search runs.
type Mode string

const (
	ModeAuto Mode = "auto"
	ModeOn   Mode = "on"
	ModeOff  Mode = "off"
)

// Parameters configures Live Search for chat.
type Parameters struct {
	Sources          []*xaiv1.Source
	Mode             Mode
	FromDate         *time.Time
	ToDate           *time.Time
	ReturnCitations  *bool
	MaxSearchResults *int32
}

// Proto converts to the API message.
func (p Parameters) Proto() *xaiv1.SearchParameters {
	out := &xaiv1.SearchParameters{Sources: p.Sources, Mode: modeToProto(p.Mode)}
	if p.FromDate != nil {
		out.FromDate = timestamppb.New(*p.FromDate)
	}
	if p.ToDate != nil {
		out.ToDate = timestamppb.New(*p.ToDate)
	}
	if p.ReturnCitations != nil {
		out.ReturnCitations = *p.ReturnCitations
	} else {
		out.ReturnCitations = true
	}
	if p.MaxSearchResults != nil {
		out.MaxSearchResults = p.MaxSearchResults
	}
	return out
}

func modeToProto(m Mode) xaiv1.SearchMode {
	switch m {
	case ModeOn:
		return xaiv1.SearchMode_ON_SEARCH_MODE
	case ModeOff:
		return xaiv1.SearchMode_OFF_SEARCH_MODE
	default:
		return xaiv1.SearchMode_AUTO_SEARCH_MODE
	}
}

// MaxWebsites is the maximum allowed/excluded sites per web/news source (Python parity).
const MaxWebsites = 5

// SourceOption configures web/news/X source factories.
type SourceOption func(*sourceCfg)
type sourceCfg struct {
	safeSearch        *bool
	postFavoriteCount *int32
	postViewCount     *int32
	validateDomains   bool
}

// WithSafeSearch sets safe_search for web/news sources (SRCH-01). Default true.
func WithSafeSearch(v bool) SourceOption {
	return func(c *sourceCfg) { c.safeSearch = &v }
}

// WithPostFavoriteCount sets X post favorite threshold (SRCH-02).
func WithPostFavoriteCount(n int32) SourceOption {
	return func(c *sourceCfg) { c.postFavoriteCount = &n }
}

// WithPostViewCount sets X post view threshold (SRCH-02).
func WithPostViewCount(n int32) SourceOption {
	return func(c *sourceCfg) { c.postViewCount = &n }
}

// WithoutDomainValidation skips allowed/excluded mutex and max-5 checks (escape hatch).
// Primary WebSource/NewsSource validate by default.
func WithoutDomainValidation() SourceOption {
	return func(c *sourceCfg) { c.validateDomains = false }
}

// WithDomainValidation is accepted for clarity; validation is already the default.
func WithDomainValidation() SourceOption {
	return func(c *sourceCfg) { c.validateDomains = true }
}

func applySourceOpts(opts []SourceOption) *sourceCfg {
	c := &sourceCfg{validateDomains: true} // fail closed by default
	for _, o := range opts {
		o(c)
	}
	return c
}

// ValidateWebsites returns an error when allowed and excluded are both set, or either exceeds MaxWebsites.
func ValidateWebsites(allowed, excluded []string) error {
	if len(allowed) > 0 && len(excluded) > 0 {
		return fmt.Errorf("search: allowed_websites and excluded_websites are mutually exclusive")
	}
	if len(allowed) > MaxWebsites {
		return fmt.Errorf("search: at most %d allowed websites", MaxWebsites)
	}
	if len(excluded) > MaxWebsites {
		return fmt.Errorf("search: at most %d excluded websites", MaxWebsites)
	}
	return nil
}

// WebSource builds a web source (primary). safe_search defaults to true.
// Validates allowed/excluded mutex and max site count unless WithoutDomainValidation.
func WebSource(country string, excluded, allowed []string, opts ...SourceOption) (*xaiv1.Source, error) {
	cfg := applySourceOpts(opts)
	if cfg.validateDomains {
		if err := ValidateWebsites(allowed, excluded); err != nil {
			return nil, err
		}
	}
	safe := true
	if cfg.safeSearch != nil {
		safe = *cfg.safeSearch
	}
	ws := &xaiv1.WebSource{ExcludedWebsites: excluded, AllowedWebsites: allowed, SafeSearch: safe}
	if country != "" {
		ws.Country = &country
	}
	return &xaiv1.Source{Source: &xaiv1.Source_Web{Web: ws}}, nil
}

// UncheckedWebSource skips domain validation.
func UncheckedWebSource(country string, excluded, allowed []string, opts ...SourceOption) *xaiv1.Source {
	src, _ := WebSource(country, excluded, allowed, append(opts, WithoutDomainValidation())...)
	return src
}

// NewsSource builds a news source (primary). Validates excluded list size by default.
func NewsSource(country string, excluded []string, opts ...SourceOption) (*xaiv1.Source, error) {
	cfg := applySourceOpts(opts)
	if cfg.validateDomains {
		if err := ValidateWebsites(nil, excluded); err != nil {
			return nil, err
		}
	}
	safe := true
	if cfg.safeSearch != nil {
		safe = *cfg.safeSearch
	}
	ns := &xaiv1.NewsSource{ExcludedWebsites: excluded, SafeSearch: safe}
	if country != "" {
		ns.Country = &country
	}
	return &xaiv1.Source{Source: &xaiv1.Source_News{News: ns}}, nil
}

// UncheckedNewsSource skips domain validation.
func UncheckedNewsSource(country string, excluded []string, opts ...SourceOption) *xaiv1.Source {
	src, _ := NewsSource(country, excluded, append(opts, WithoutDomainValidation())...)
	return src
}

// ValidateXHandles returns an error when included and excluded handles are both set.
func ValidateXHandles(included, excluded []string) error {
	if len(included) > 0 && len(excluded) > 0 {
		return fmt.Errorf("search: included_x_handles and excluded_x_handles are mutually exclusive")
	}
	return nil
}

// XSource builds an X source (primary). Optional WithPostFavoriteCount /
// WithPostViewCount. Validates included/excluded handle mutual exclusion
// unless WithoutDomainValidation, matching tools.XSearch behavior.
func XSource(included, excluded []string, opts ...SourceOption) (*xaiv1.Source, error) {
	cfg := applySourceOpts(opts)
	if cfg.validateDomains {
		if err := ValidateXHandles(included, excluded); err != nil {
			return nil, err
		}
	}
	xs := &xaiv1.XSource{
		IncludedXHandles: included, ExcludedXHandles: excluded,
	}
	if cfg.postFavoriteCount != nil {
		xs.PostFavoriteCount = cfg.postFavoriteCount
	}
	if cfg.postViewCount != nil {
		xs.PostViewCount = cfg.postViewCount
	}
	return &xaiv1.Source{Source: &xaiv1.Source_X{X: xs}}, nil
}

// UncheckedXSource skips handle validation.
func UncheckedXSource(included, excluded []string, opts ...SourceOption) *xaiv1.Source {
	src, _ := XSource(included, excluded, append(opts, WithoutDomainValidation())...)
	return src
}

// RSSSource builds an RSS source.
func RSSSource(links []string) *xaiv1.Source {
	return &xaiv1.Source{Source: &xaiv1.Source_Rss{Rss: &xaiv1.RssSource{Links: links}}}
}
