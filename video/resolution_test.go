package video_test

import (
	"testing"

	"github.com/fun7257/xai-sdk-go/video"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

// Unknown resolutions must error like unknown aspect ratios (no silent 480p fallback).
func TestPrepareRejectsUnknownResolution(t *testing.T) {
	cli := video.New(nil)
	if _, err := cli.Prepare("p", "m", video.WithResolution("1080p")); err == nil {
		t.Fatal("expected error for unknown resolution")
	}
	req, err := cli.Prepare("p", "m", video.WithResolution("720p"))
	if err != nil {
		t.Fatal(err)
	}
	if req.GetResolution() != xaiv1.VideoResolution_VIDEO_RESOLUTION_720P {
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
