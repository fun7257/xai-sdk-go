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
	web, err := tools.WebSearch()
	if err != nil {
		log.Fatal(err)
	}
	c := client.Chat.Create(types.ModelGrok3,
		chat.WithTools(web, tools.CodeExecution()),
		chat.WithMessages(chat.User("What is the weather in SF?")),
	)
	resp, err := c.Sample(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Content(), resp.ServerSideToolUsage())
}
