package video_test

import (
	"strings"
	"testing"

	"github.com/fun7257/xai-sdk-go/video"
)

// Options without ExtendVideoRequest fields must be rejected, not silently
// ignored (parity with the strict-input convention).
func TestPrepareExtensionRejectsUnsupportedOptions(t *testing.T) {
	cli := video.New(nil)
	cases := []struct {
		name string
		opt  video.GenerateOption
		want string
	}{
		{"aspect", video.WithAspectRatio("16:9"), "WithAspectRatio"},
		{"resolution", video.WithResolution("720p"), "WithResolution"},
		{"image url", video.WithImageURL("https://x/im.png"), "first-frame image"},
		{"image file", video.WithImageFileID("f1"), "first-frame image"},
		{"reference images", video.WithReferenceImages("https://x/r.png"), "reference images"},
		{"video url option", video.WithVideoURL("https://x/v.mp4"), "argument"},
		{"video file option", video.WithVideoFileID("f2"), "argument"},
	}
	for _, tc := range cases {
		_, err := cli.PrepareExtension("p", "m", "https://x/src.mp4", tc.opt)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: want error containing %q, got %v", tc.name, tc.want, err)
		}
	}

	// Supported options still work.
	req, err := cli.PrepareExtension("p", "m", "https://x/src.mp4", video.WithDuration(3))
	if err != nil {
		t.Fatal(err)
	}
	if req.GetDuration() != 3 || req.GetVideo().GetUrl() != "https://x/src.mp4" {
		t.Fatalf("%+v", req)
	}
}
