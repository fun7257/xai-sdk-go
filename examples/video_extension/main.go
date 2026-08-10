package main

import (
	"context"
	"fmt"
	"log"
	"os"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/types"
)

func main() {
	if os.Getenv("XAI_API_KEY") == "" {
		log.Fatal("XAI_API_KEY is required")
	}
	url := os.Getenv("XAI_VIDEO_URL")
	if url == "" {
		log.Fatal("XAI_VIDEO_URL is required")
	}
	client, err := xai.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	id, err := client.Video.ExtendStart(context.Background(), "continue the scene", types.ModelImagineVideo, url)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("request_id", id)
}
