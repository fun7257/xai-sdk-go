// Short video generation (poll until done). Expensive — prefer smoke with SMOKE_SKIP_VIDEO for CI.
//
//	export XAI_API_KEY=xai-...
//	go run ./examples/video
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/video"
)

func main() {
	if os.Getenv("XAI_API_KEY") == "" {
		log.Fatal("XAI_API_KEY is required")
	}
	client, err := xai.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := envOr("XAI_VIDEO_MODEL", "grok-imagine-video")
	resp, err := client.Video.Generate(context.Background(), "a cat walks", model,
		video.WithDuration(1),
		video.WithResolution("480p"),
	)
	if err != nil {
		log.Fatal(err)
	}
	u, err := resp.URL()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(u)
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
