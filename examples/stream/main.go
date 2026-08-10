// Streaming chat with StreamReader.Recv.
//
//	export XAI_API_KEY=xai-...
//	go run ./examples/stream
package main

import (
	"context"
	"fmt"
	"io"
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
	c := client.Chat.Create(model)
	_ = c.Append(chat.User("Count from 1 to 5, one number per line."))

	sr, err := c.StreamReader(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer sr.Close()
	for {
		ev, err := sr.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		if ev.Chunk != nil {
			fmt.Print(ev.Chunk.Content())
		}
	}
	fmt.Println()
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
