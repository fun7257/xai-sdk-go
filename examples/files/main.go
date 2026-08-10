// List files (and optional upload when XAI_UPLOAD_PATH is set).
//
//	export XAI_API_KEY=xai-...
//	go run ./examples/files
//	XAI_UPLOAD_PATH=./note.txt go run ./examples/files
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/files"
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

	ctx := context.Background()
	if p := os.Getenv("XAI_UPLOAD_PATH"); p != "" {
		f, err := client.Files.UploadPath(ctx, p)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("uploaded id=%s name=%s\n", f.Id, f.Filename)
	}

	resp, err := client.Files.List(ctx, files.WithListLimit(10), files.WithListSortBy("created_at"))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range resp.Data {
		fmt.Printf("%s\t%s\t%d\n", f.Id, f.Filename, f.Size)
	}
}
