// Minimal chat Sample example.
//
//	export XAI_API_KEY=xai-...
//	go run ./examples/chat
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/chat"
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

	model := envOr("XAI_MODEL", "grok-3")
	c := client.Chat.Create(model, chat.WithMessages(chat.System("You are helpful.")))
	_ = c.Append(chat.User("Say hello in one short sentence."))
	resp, err := c.Sample(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Content())
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
