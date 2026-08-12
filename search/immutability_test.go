package search_test

import (
	"testing"

	"github.com/fun7257/xai-sdk-go/search"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// Proto must copy the sources slice and scalar pointers so later Parameters
// mutations do not leak into a built request.
func TestProtoCopiesMutableState(t *testing.T) {
	x, err := search.XSource([]string{"xai"}, nil)
	if err != nil {
		t.Fatal(err)
	}
	limit := int32(3)
	p := search.Parameters{
		Sources:          []*xaiv1.Source{x},
		MaxSearchResults: &limit,
	}
	pb := p.Proto()

	p.Sources[0] = nil // caller mutates after building
	limit = 99

	if len(pb.GetSources()) != 1 || pb.GetSources()[0] == nil || pb.GetSources()[0].GetX() == nil {
		t.Fatalf("sources not isolated: %+v", pb.GetSources())
	}
	if pb.GetMaxSearchResults() != 3 {
		t.Fatalf("max results not isolated: %d", pb.GetMaxSearchResults())
	}
}
