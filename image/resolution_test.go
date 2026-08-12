package image_test

import (
	"testing"

	"github.com/fun7257/xai-sdk-go/image"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// Unknown resolutions must error like unknown aspect ratios (no silent 1k fallback).
func TestPrepareRejectsUnknownResolution(t *testing.T) {
	cli := image.New(nil)
	if _, err := cli.Prepare("p", "m", image.WithResolution("4k")); err == nil {
		t.Fatal("expected error for unknown resolution")
	}
	if _, err := cli.Prepare("p", "m", image.WithResolution("2K")); err == nil {
		t.Fatal("expected error for wrong-case resolution")
	}
	req, err := cli.Prepare("p", "m", image.WithResolution("2k"))
	if err != nil {
		t.Fatal(err)
	}
	if req.GetResolution() != xaiv1.ImageResolution_IMG_RESOLUTION_2K {
		t.Fatalf("resolution=%v", req.GetResolution())
	}
	req, err = cli.Prepare("p", "m")
	if err != nil {
		t.Fatal(err)
	}
	if req.Resolution != nil {
		t.Fatalf("unset resolution must stay nil, got %v", req.Resolution)
	}
}
