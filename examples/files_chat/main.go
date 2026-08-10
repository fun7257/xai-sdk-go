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
	fid := os.Getenv("XAI_FILE_ID")
	if fid == "" {
		log.Fatal("XAI_FILE_ID is required")
	}
	client, err := xai.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()
	c := client.Chat.Create(types.ModelGrok3, chat.WithMessages(
		chat.User("summarize", chat.FileByID(fid)),
	))
	resp, err := c.Sample(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Content())
}
