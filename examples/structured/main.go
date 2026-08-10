// Structured JSON output via Chat.Parse.
//
//	export XAI_API_KEY=xai-...
//	go run ./examples/structured
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/chat"
)

type person struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func main() {
	if os.Getenv("XAI_API_KEY") == "" {
		log.Fatal("XAI_API_KEY is required")
	}
	client, err := xai.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	schema := []byte(`{
  "type": "object",
  "properties": {
    "name": {"type": "string"},
    "age": {"type": "integer"}
  },
  "required": ["name", "age"]
}`)

	model := envOr("XAI_MODEL", "grok-3")
	c := client.Chat.Create(model)
	_ = c.Append(chat.User("Extract: Alice is 30 years old."))
	var out person
	resp, err := c.Parse(context.Background(), schema, &out)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("parsed: %+v\nraw: %s\n", out, resp.Content())
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
