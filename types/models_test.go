package types_test

import (
	"testing"

	"github.com/fun7257/xai-sdk-go/image"
	"github.com/fun7257/xai-sdk-go/types"
	"github.com/fun7257/xai-sdk-go/video"
)

// Every exported aspect/resolution constant must be accepted by its
// validating option, so constants and validators cannot drift apart.
func TestConstantsAcceptedByValidators(t *testing.T) {
	imgCli := image.New(nil)
	for _, ar := range []string{
		types.Aspect1_1, types.Aspect16_9, types.Aspect9_16, types.Aspect3_4,
		types.Aspect4_3, types.Aspect2_3, types.Aspect3_2, types.AspectAuto,
	} {
		if _, err := imgCli.Prepare("p", "m", image.WithAspectRatio(ar)); err != nil {
			t.Errorf("image aspect %q rejected: %v", ar, err)
		}
	}
	for _, res := range []string{types.ImageRes1K, types.ImageRes2K} {
		if _, err := imgCli.Prepare("p", "m", image.WithResolution(res)); err != nil {
			t.Errorf("image resolution %q rejected: %v", res, err)
		}
	}

	vidCli := video.New(nil)
	for _, ar := range []string{
		types.VideoAspect1_1, types.VideoAspect16_9, types.VideoAspect9_16,
		types.VideoAspect4_3, types.VideoAspect3_4, types.VideoAspect3_2, types.VideoAspect2_3,
	} {
		if _, err := vidCli.Prepare("p", "m", video.WithAspectRatio(ar)); err != nil {
			t.Errorf("video aspect %q rejected: %v", ar, err)
		}
	}
	for _, res := range []string{types.VideoRes480p, types.VideoRes720p} {
		if _, err := vidCli.Prepare("p", "m", video.WithResolution(res)); err != nil {
			t.Errorf("video resolution %q rejected: %v", res, err)
		}
	}
}
