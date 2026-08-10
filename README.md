# xAI Go SDK

Idiomatic Go gRPC client for the [xAI API](https://docs.x.ai).  
**Go call UX first** — product capability without cloning Python API shapes.

**Module:** `github.com/fun7257/xai-sdk-go` · **Go:** 1.26+ · **License:** Apache-2.0 · **Release:** git tags (`vX.Y.Z`); runtime `xai.Version()`
---

## Documentation map

| Role | Start here | Details |
|------|------------|---------|
| **New user** | [Install & quick start](#install) below | Full params (EN+中文): [`docs/GUIDE.zh-en.md`](docs/GUIDE.zh-en.md) · Runnable: [`examples/complete/`](examples/complete/) |
| **API 参考** | [`docs/API.md`](docs/API.md)（包/函数完备说明） | `go doc` / pkg.go.dev |
| **API design** | [`docs/PARITY.md`](docs/PARITY.md) (principles + preferred shapes) | Domain vs Python: [`docs/DIFF.md`](docs/DIFF.md) |
| **Stability / pb** | [`docs/COMPATIBILITY.md`](docs/COMPATIBILITY.md) | Residual protos: [`docs/PROTO.md`](docs/PROTO.md) |
| **Security** | [`docs/SECURITY.md`](docs/SECURITY.md) | — |
| **Contributor** | [`CONTRIBUTING.md`](CONTRIBUTING.md) | Examples index: [`examples/README.md`](examples/README.md) |
| **Maintainer** | [`docs/RELEASE.md`](docs/RELEASE.md) | CI triggers / `dev` tag: [`docs/CI.md`](docs/CI.md) · OSS status: [`docs/ROADMAP_OSS.md`](docs/ROADMAP_OSS.md) |
| **Changelog** | [`CHANGELOG.md`](CHANGELOG.md) | — |

---

## Install

```bash
go get github.com/fun7257/xai-sdk-go
```

### Credentials

| Key | Env | Option | Used for |
|-----|-----|--------|----------|
| Business | `XAI_API_KEY` | `WithAPIKey` | Chat, Image, Video, Files, Batch, Models, Tokenize, Auth, search |
| Management | `XAI_MANAGEMENT_KEY` | `WithManagementAPIKey` | Collections CRUD / indexing |

RPCs send `authorization: Bearer <key>`. Prefer env vars; never commit secrets.  
Sentinels: `ErrNoAPIKey`, `ErrEmptyAPIKey`, `ErrNoManagementKey` (`errors.Is`).  
Full option tables → [`docs/GUIDE.zh-en.md`](docs/GUIDE.zh-en.md).

---

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
	client, err := xai.NewClient() // XAI_API_KEY
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

**Preferred shapes** (see [`docs/PARITY.md`](docs/PARITY.md)):

| Goal | API |
|------|-----|
| One completion | `Sample(ctx)` |
| Multi | `Samples(ctx, chat.WithN(k))` — never silent drop |
| Stream | `StreamReader` + `Recv` / `Close` |
| Tools | `tools.WebSearch(...)` → `(*Tool, error)` |

```bash
export XAI_API_KEY=xai-...
go run ./examples/complete   # annotated end-to-end
```

---

## Domains

| Field | Package | Notes |
|-------|---------|-------|
| `Auth` | `auth` | API key info |
| `Chat` | `chat` | Sample, StreamReader, Parse, Defer, stored, compact |
| `Image` / `Video` | `image` / `video` | Gen + poll; file_id / storage where applicable |
| `Files` | `files` | Upload, list, public URL, `ContentWriter` |
| `Batch` | `batch` | Typed `batch.Result` |
| `Collections` | `collections` | Management key for CRUD |
| `Models` / `Tokenize` | `models` / `tokenize` | Metadata / tokenize |
| — | `tools`, `search`, `telemetry`, `types` | Tools, live search, opt-in OTEL, constants |

**Out of scope:** Embed client, OpenAI REST, OAuth, aio dual stack.

**Telemetry (opt-in):** `telemetry.Setup(...)`. Disable: `XAI_SDK_DISABLE_TRACING=1`.

---

## Develop & release

| Topic | Doc |
|-------|-----|
| Offline gates (local) | [`CONTRIBUTING.md`](CONTRIBUTING.md) |
| **CI mode** (PR / main / **`v*` tags** / **daily tests** / floating **`dev`**) | [`docs/CI.md`](docs/CI.md) |
| Release tags + `dev` channel | [`docs/RELEASE.md`](docs/RELEASE.md) |
| Security | [`docs/SECURITY.md`](docs/SECURITY.md) |

```bash
go get github.com/fun7257/xai-sdk-go@vX.Y.Z   # release (immutable)
go get github.com/fun7257/xai-sdk-go@dev      # latest green main (overwritten)
```
