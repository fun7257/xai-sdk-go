package main

import (
	"context"
	"fmt"
	"log"
	"os"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/chat"
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
	c := client.Chat.Create(types.ModelGrok3, chat.WithStoreMessages(true), chat.WithMessages(chat.User("remember this")))
	resp, err := c.Sample(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.ID())
	if id := resp.ID(); id != "" {
		_, _ = client.Chat.GetStoredCompletion(ctx, id)
	}
}
