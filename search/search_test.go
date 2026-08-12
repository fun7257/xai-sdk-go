package search_test

import (
	"testing"

	"github.com/fun7257/xai-sdk-go/search"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

func TestParametersProto(t *testing.T) {
	web, err := search.WebSource("US", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	news, err := search.NewsSource("US", nil)
	if err != nil {
		t.Fatal(err)
	}
	x, err := search.XSource([]string{"xai"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	p := search.Parameters{
		Sources: []*xaiv1.Source{web, news, x, search.RSSSource([]string{"https://x.com/rss"})},
		Mode:    search.ModeOn,
	}
	pb := p.Proto()
	if pb.Mode != xaiv1.SearchMode_ON_SEARCH_MODE {
		t.Fatal(pb.Mode)
	}
	if len(pb.Sources) != 4 {
		t.Fatal(len(pb.Sources))
	}
	if !pb.ReturnCitations {
		t.Fatal("citations default")
	}
}

func TestSafeSearchAndXThresholds(t *testing.T) {
	ws, err := search.WebSource("US", nil, nil, search.WithSafeSearch(false))
	if err != nil {
		t.Fatal(err)
	}
	if ws.GetWeb().SafeSearch {
		t.Fatal("expected safe_search false")
	}
	xs, err := search.XSource([]string{"xai"}, nil, search.WithPostFavoriteCount(10), search.WithPostViewCount(100))
	if err != nil {
		t.Fatal(err)
	}
	if xs.GetX().GetPostFavoriteCount() != 10 || xs.GetX().GetPostViewCount() != 100 {
		t.Fatalf("%+v", xs.GetX())
	}
	_, err = search.WebSource("US", []string{"a.com"}, []string{"b.com"})
	if err == nil {
		t.Fatal("expected mutex error")
	}
}

func TestXSourceHandleMutex(t *testing.T) {
	if _, err := search.XSource([]string{"a"}, []string{"b"}); err == nil {
		t.Fatal("expected included/excluded mutex error")
	}
	if src := search.UncheckedXSource([]string{"a"}, []string{"b"}); src == nil || src.GetX() == nil {
		t.Fatal("unchecked variant must build the source")
	}
}
