// Collections list skeleton (requires management key).
//
//	export XAI_API_KEY=xai-...
//	export XAI_MANAGEMENT_KEY=xai-...
//	go run ./examples/collections
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/collections"
)

func main() {
	if os.Getenv("XAI_API_KEY") == "" || os.Getenv("XAI_MANAGEMENT_KEY") == "" {
		log.Fatal("XAI_API_KEY and XAI_MANAGEMENT_KEY are required")
	}
	client, err := xai.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	resp, err := client.Collections.List(context.Background(), collections.WithListLimit(20))
	if err != nil {
		log.Fatal(err)
	}
	for _, c := range resp.Collections {
		fmt.Printf("%s\t%s\n", c.CollectionId, c.CollectionName)
	}
}
