package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

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
	// Deferred completions poll (default up to 10m); bound the wait.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	c := client.Chat.Create(types.ModelGrok3, chat.WithMessages(chat.User("hi")))
	resp, err := c.Defer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Content())
}
