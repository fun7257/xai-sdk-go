// Opt-in OTEL setup skeleton (no OTLP exporter wired; spans go to configured provider).
//
//	export XAI_API_KEY=xai-...
//	go run ./examples/telemetry
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/chat"
	"github.com/fun7257/xai-sdk-go/telemetry"
)

func main() {
	if os.Getenv("XAI_API_KEY") == "" {
		log.Fatal("XAI_API_KEY is required")
	}

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	defer func() { _ = tp.Shutdown(context.Background()) }()
	telemetry.Setup(telemetry.SetupOptions{TracerProvider: tp})

	client, err := xai.NewClient()
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	model := "grok-3"
	if v := os.Getenv("XAI_MODEL"); v != "" {
		model = v
	}
	c := client.Chat.Create(model)
	_ = c.Append(chat.User("Reply with the single word: pong"))
	resp, err := c.Sample(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Content())
	fmt.Printf("recorded spans: %d\n", len(exporter.GetSpans()))
}
