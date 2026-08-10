// Client-side function tool example (single round).
//
//	export XAI_API_KEY=xai-...
//	go run ./examples/function_calling
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/chat"
	"github.com/fun7257/xai-sdk-go/tools"
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

	fn, err := tools.Function("get_weather", "Get weather for a city", map[string]any{
		"type": "object",
		"properties": map[string]any{
			"city": map[string]any{"type": "string"},
		},
		"required": []string{"city"},
	})
	if err != nil {
		log.Fatal(err)
	}

	model := envOr("XAI_MODEL", "grok-3")
	c := client.Chat.Create(model, chat.WithTools(fn), chat.WithToolChoice(tools.Mode("auto")))
	_ = c.Append(chat.User("What is the weather in Tokyo?"))
	resp, err := c.Sample(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for _, tc := range resp.ToolCalls() {
		fmt.Printf("tool call: %s args=%s\n", tc.GetFunction().GetName(), tc.GetFunction().GetArguments())
	}
	if content := resp.Content(); content != "" {
		fmt.Println(content)
	}
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
