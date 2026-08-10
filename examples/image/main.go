// Image generation (URL format by default).
//
//	export XAI_API_KEY=xai-...
//	go run ./examples/image
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/image"
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

	model := envOr("XAI_IMAGE_MODEL", "grok-imagine-image")
	resp, err := client.Image.Sample(context.Background(), "a watercolor fox", model,
		image.WithFormatURL(),
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
