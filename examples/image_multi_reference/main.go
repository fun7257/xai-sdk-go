package main

import (
	"context"
	"fmt"
	"log"
	"os"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/image"
	"github.com/fun7257/xai-sdk-go/types"
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
	all, err := client.Image.Samples(context.Background(), "blend", types.ModelImagineImage,
		image.WithImageURLs("https://example.com/a.png", "https://example.com/b.png"),
	)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(all))
}
