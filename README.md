# xAI Go SDK

[**English**](README.md) · [**中文**](README.zh-CN.md)

Idiomatic **Go gRPC** client for the [xAI API](https://docs.x.ai).  
**Go call UX first** — product capability without cloning Python API shapes.

<p align="left">
  <a href="https://pkg.go.dev/github.com/fun7257/xai-sdk-go"><img src="https://pkg.go.dev/badge/github.com/fun7257/xai-sdk-go.svg" alt="Go Reference"></a>
  <a href="https://github.com/fun7257/xai-sdk-go/actions/workflows/ci.yaml"><img src="https://github.com/fun7257/xai-sdk-go/actions/workflows/ci.yaml/badge.svg" alt="CI"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License"></a>
  <a href="https://go.dev/dl/"><img src="https://img.shields.io/badge/go-1.26%2B-00ADD8?logo=go" alt="Go 1.26+"></a>
</p>

```text
module   github.com/fun7257/xai-sdk-go
release  go get …@vX.Y.Z          # immutable tag
dev tip  go get …@dev             # latest green main (overwritten)
```

**[Install](#install)** ·
**[Quick start](#quick-start)** ·
**[Patterns](#core-patterns)** ·
**[Packages](#packages)** ·
**[Docs](#documentation)** ·
**[Examples](#examples)** ·
**[CI](#versioning--ci)** ·
**[Contributing](#contributing)**

---

## Install

```bash
go get github.com/fun7257/xai-sdk-go@v0.1.1
```

<details>
<summary><strong>Credentials</strong> — environment variables & options</summary>

<br>

| | Environment | Option | Used for |
|--|-------------|--------|----------|
| **Business** | `XAI_API_KEY` | `WithAPIKey` | Chat · Image · Video · Files · Batch · Models · Tokenize · Auth · Search |
| **Management** | `XAI_MANAGEMENT_KEY` | `WithManagementAPIKey` | Collections CRUD / indexing |

- RPC header: `Authorization: Bearer <key>`
- Prefer environment variables; never commit secrets
- Sentinel errors: `ErrNoAPIKey` · `ErrEmptyAPIKey` · `ErrNoManagementKey` (`errors.Is`)

</details>

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
	"github.com/fun7257/xai-sdk-go/types"
)

func main() {
	client, err := xai.NewClient() // XAI_API_KEY
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	c := client.Chat.Create(types.ModelGrok45Latest,
		chat.WithMessages(chat.System("You are helpful.")),
	)
	_ = c.Append(chat.User("Hello!"))

	resp, err := c.Sample(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(resp.Content())
}
```

```bash
export XAI_API_KEY=xai-...
go run ./examples/complete    # annotated end-to-end walkthrough
```

---

## Core patterns

| | Prefer | Notes |
|--|--------|--------|
| **One completion** | `Sample(ctx)` | Use `Samples` when `n > 1` |
| **Multi completion** | `Samples(ctx, chat.WithN(k))` | Never silently drop results |
| **Streaming** | `StreamReader` → `Recv` / `Close` | Always `Close` |
| **Structured** | `Parse(ctx, schema, &dest)` | JSON Schema |
| **Tools** | `tools.WebSearch(...) (*Tool, error)` | Validate on primary path · `Unchecked*` escape hatch |
| **Image / video** | `Image.Sample` · `Video.Generate` | Multi-image: `Samples` + `WithN` |

> Design principles and full preferred shapes → [**PARITY.md**](docs/PARITY.md).

---

## Packages

Access via `client.<Field>` (or each package’s `New`).

| `Client` field | Package | Capability |
|----------------|---------|------------|
| `Chat` | `chat` | Multi-turn · stream · tools · Defer · Parse · store / compact |
| `Image` · `Video` | `image` · `video` | Generate · extend · poll · storage options |
| `Files` | `files` | Upload · list · public URL · `ContentWriter` |
| `Batch` | `batch` | Batch jobs · typed `Result` |
| `Collections` | `collections` | Collections / docs / search (CRUD needs management key) |
| `Auth` · `Models` · `Tokenize` | same name | Key info · model catalog · tokenize |
| — | `tools` · `search` · `telemetry` · `types` | Tools · Live Search · OTEL · constants |

```text
Out of scope   Embed client · OpenAI REST · OAuth · aio dual-stack
Telemetry      telemetry.Setup(...)     off: XAI_SDK_DISABLE_TRACING=1
```

---

## Documentation

| Need | Document |
|------|----------|
| Full API catalogue | [**API.md**](docs/API.md) |
| Parameter tables (EN + 中文) & walkthrough | [**GUIDE.zh-en.md**](docs/GUIDE.zh-en.md) |
| Design principles · preferred APIs | [**PARITY.md**](docs/PARITY.md) |
| Differences vs Python SDK | [**DIFF.md**](docs/DIFF.md) |
| Stability · protobuf | [COMPATIBILITY](docs/COMPATIBILITY.md) · [PROTO](docs/PROTO.md) |
| Security reporting | [SECURITY.md](docs/SECURITY.md) |
| Release · CI · `dev` tip | [RELEASE](docs/RELEASE.md) · [CI](docs/CI.md) |
| Changelog | [CHANGELOG.md](CHANGELOG.md) |

```bash
go doc github.com/fun7257/xai-sdk-go
go doc github.com/fun7257/xai-sdk-go/chat.Chat.Sample
```

---

## Examples

| | |
|--|--|
| [**examples/complete**](examples/complete/) | Annotated end-to-end (start here) |
| [**examples/README.md**](examples/README.md) | Full catalogue |
| [**examples/smoke**](examples/smoke/) | Low-cost live smoke |

```bash
export XAI_API_KEY=xai-...
go run ./examples/chat
go build ./examples/...
```

---

## Versioning & CI

| Channel | Command | Notes |
|---------|---------|--------|
| **Release** | `go get …@vX.Y.Z` | Immutable git tag |
| **Dev tip** | `go get …@dev` | Force-moved after green `main` CI |
| **Runtime** | `xai.Version()` | From module graph — not a hand-maintained constant |

CI runs on PRs, `main`, `v*` tags, and a daily offline suite; optional live integration. Details → [**CI.md**](docs/CI.md).

---

## Contributing

See [**CONTRIBUTING.md**](CONTRIBUTING.md). Security issues: [**SECURITY.md**](docs/SECURITY.md) (do not file public issues with secrets).

```bash
make check && make race && make lint
```
