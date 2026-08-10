package main

import (
	"context"
	"fmt"
	"log"
	"os"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/chat"
	"github.com/fun7257/xai-sdk-go/tools"
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
	ctx := context.Background()
	cid := os.Getenv("XAI_COLLECTION_ID")
	if cid == "" {
		cid = "collection-id"
	}
	tool, err := tools.CollectionsSearchOpts([]string{cid}, tools.WithCollectionsLimit(3))
	if err != nil {
		log.Fatal(err)
	}
	c := client.Chat.Create(types.ModelGrok3, chat.WithTools(tool), chat.WithMessages(chat.User("search docs")))
	resp, err := c.Sample(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Content())
}
