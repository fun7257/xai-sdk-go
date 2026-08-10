# xAI Go SDK

Idiomatic Go gRPC client for the [xAI API](https://docs.x.ai). **Go call UX first** — product capability without cloning Python API shapes.

- Design summary: [`docs/PARITY.md`](docs/PARITY.md)
- Full Go vs Python differences: [`docs/DIFF.md`](docs/DIFF.md)
- API stability (v0 / protobuf): [`docs/COMPATIBILITY.md`](docs/COMPATIBILITY.md)
- Security reporting: [`docs/SECURITY.md`](docs/SECURITY.md)
- Proto residual: [`docs/PROTO.md`](docs/PROTO.md)
- OSS roadmap: [`docs/ROADMAP_OSS.md`](docs/ROADMAP_OSS.md)
- Release checklist: [`docs/RELEASE.md`](docs/RELEASE.md)
- **Complete bilingual usage guide / 中英双语完整调用指南:** [`docs/GUIDE.zh-en.md`](docs/GUIDE.zh-en.md)
- Annotated example / 带注释示例: [`examples/complete/`](examples/complete/)

**Module:** `github.com/fun7257/xai-sdk-go` · **Go:** 1.26+ · **License:** Apache-2.0

## Install

```bash
go get github.com/fun7257/xai-sdk-go
```

## Auth

| Key | Env | Option | Used for |
|-----|-----|--------|----------|
| Business API key | `XAI_API_KEY` | `WithAPIKey` | Chat, Image, Video, Files, Batch, Models, Tokenize, Auth, Documents search |
| Management API key | `XAI_MANAGEMENT_KEY` | `WithManagementAPIKey` | Collections CRUD / indexing |

Every RPC sends `authorization: Bearer <key>`. Optional `WithInsecure()` for local dial (Bearer still attached).

## Quick start

```go
package main

import (
	"context"
	"fmt"
	"log"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/chat"
)

func main() {
	client, err := xai.NewClient() // reads XAI_API_KEY
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	c := client.Chat.Create("grok-3", chat.WithMessages(chat.System("You are helpful.")))
	_ = c.Append(chat.User("Hello!"))
	resp, err := c.Sample(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Content())
}
```

Streaming (primary DX):

```go
sr, err := c.StreamReader(ctx) // optional: chat.WithN(k)
if err != nil {
	log.Fatal(err)
}
defer sr.Close()
for {
	ev, err := sr.Recv()
	if err == io.EOF {
		break
	}
	// use ev.Chunk / ev.Response
}
```

Multi-completion:

```go
// chat
resps, err := c.Samples(ctx, chat.WithN(3))
// image
imgs, err := client.Image.Samples(ctx, "a cat", "grok-imagine-image", image.WithN(2))
```

Tools validate on construct:

```go
web, err := tools.WebSearch(tools.WithAllowedDomains("example.com"))
// illegal allowed+excluded → error
```

## Domains

| Domain | Package / field | Notes |
|--------|-----------------|-------|
| Auth | `client.Auth` | API key info |
| Chat | `client.Chat` | Sample, StreamReader, Parse, Defer, stored, compact |
| Image | `client.Image` | file_id + storage options |
| Video | `client.Video` | Generate/Extend + poll |
| Files | `client.Files` | chunked upload, list filter/sort, public URL |
| Batch | `client.Batch` | typed `batch.Result` |
| Collections | `client.Collections` | management key; UploadDocument |
| Models | `client.Models` | language / image gen models |
| Tokenize | `client.Tokenize` | TokenizeText |

## Capability matrix (high level)

| Capability | Status |
|------------|--------|
| Chat multi-turn Sample/Stream | yes |
| Tools (server + function) | yes |
| Structured Parse | yes |
| Deferred / stored / compact | yes |
| Image + Video gen | yes |
| Files + public URL | yes |
| Batch typed results | yes |
| Collections RAG | yes (management key) |
| OTEL telemetry | opt-in (`telemetry.Setup`) |
| Embed client | **out of scope** |
| OpenAI REST | **out of scope** |

## Differences from Python SDKs (by design)

| Topic | Go choice |
|-------|-----------|
| Concurrency | One client + `context` / goroutines (no aio dual stack) |
| Config | Functional options |
| Multi results | `Samples` / `WithN` — never silent single-index drop |
| Validation | Primary constructors return `error` on illegal combos |
| Responses | Explicit methods + `Proto()` escape hatch |
| Telemetry | Opt-in `telemetry.Setup` |

## Examples

See [`examples/README.md`](examples/README.md). Live smoke:

```bash
export XAI_API_KEY=xai-...
go run ./examples/smoke
SMOKE_SKIP_VIDEO=1 go run ./examples/smoke
```

## Tests and quality gates

```bash
make check   # vet + test + examples build (offline)
make race    # race on concurrent packages
make lint    # golangci-lint
make vuln    # govulncheck
```

CI runs the same offline gates (see `.github/workflows/ci.yaml`).

Optional live smoke (skips without `XAI_API_KEY`; not required for PRs):

```bash
export XAI_API_KEY=xai-...
make integration
```

## Proto

Generated from [xai-org/xai-proto](https://github.com/xai-org/xai-proto) under `third_party/`, plus residual synthesized fields for storage/file_id/public URL/collections. Details: [`docs/PROTO.md`](docs/PROTO.md).

```bash
make proto
```

## Telemetry

```go
import "github.com/fun7257/xai-sdk-go/telemetry"

telemetry.Setup(telemetry.SetupOptions{}) // uses otel.GetTracerProvider()
// or pass TracerProvider in SetupOptions
```

Disable: `XAI_SDK_DISABLE_TRACING=1`  
Omit sensitive attributes: `XAI_SDK_DISABLE_SENSITIVE_TELEMETRY_ATTRIBUTES=1`

## Contributing

See [`CONTRIBUTING.md`](CONTRIBUTING.md). Stability and security: [`docs/COMPATIBILITY.md`](docs/COMPATIBILITY.md), [`docs/SECURITY.md`](docs/SECURITY.md).
