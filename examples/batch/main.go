// Batch helpers + chat PrepareBatchRequest path.
//
//	export XAI_API_KEY=xai-...
//	go run ./examples/batch
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/batch"
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

	// Cheap: list existing batches.
	resp, err := client.Batch.List(ctx, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("batches page: %d items\n", len(resp.Batches))

	// Offline construction: chat session → PrepareBatchRequest + batch_request_id.
	// (Does not submit unless you call Batch.Add with a real batch id.)
	ch := client.Chat.Create(types.ModelGrok3,
		chat.WithMessages(chat.User("batch hello")),
		chat.WithBatchRequestID("example-req-1"),
	)
	req, err := ch.PrepareBatchRequest()
	if err != nil {
		log.Fatal(err)
	}
	br := batch.ChatBatchRequest(req, ch.BatchRequestID())
	id := ""
	if br.BatchRequestId != nil {
		id = *br.BatchRequestId
	}
	fmt.Printf("prepared batch_request_id=%q model=%q messages=%d\n",
		id, req.GetModel(), len(req.GetMessages()))

	if os.Getenv("XAI_BATCH_CREATE") == "1" {
		b, err := client.Batch.Create(ctx, "sdk-example", "")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("created batch_id=%s\n", b.BatchId)
	}
}
