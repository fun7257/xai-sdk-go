# xAI Go SDK

Idiomatic **Go gRPC** client for the [xAI API](https://docs.x.ai).  
优先 **Go 调用体验** — 对齐产品能力，而不是克隆 Python API 形态。

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
<summary><strong>Credentials</strong> — 环境变量与选项</summary>

<br>

| | Environment | Option | Used for |
|--|-------------|--------|----------|
| **Business** | `XAI_API_KEY` | `WithAPIKey` | Chat · Image · Video · Files · Batch · Models · Tokenize · Auth · Search |
| **Management** | `XAI_MANAGEMENT_KEY` | `WithManagementAPIKey` | Collections CRUD / indexing |

- RPC 头：`Authorization: Bearer <key>`
- 勿把密钥写进仓库；优先环境变量
- 可判定错误：`ErrNoAPIKey` · `ErrEmptyAPIKey` · `ErrNoManagementKey`（`errors.Is`）

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
go run ./examples/complete    # 中英注释 · 端到端
```

---

## Core patterns

| | Prefer | Notes |
|--|--------|--------|
| **单次补全** | `Sample(ctx)` | `n > 1` 用 `Samples` |
| **多次补全** | `Samples(ctx, chat.WithN(k))` | 不静默丢弃 |
| **流式** | `StreamReader` → `Recv` / `Close` | 务必 `Close` |
| **结构化** | `Parse(ctx, schema, &dest)` | JSON Schema |
| **工具** | `tools.WebSearch(...) (*Tool, error)` | 主路径校验 · `Unchecked*` 逃生 |
| **图像 / 视频** | `Image.Sample` · `Video.Generate` | 多图：`Samples` + `WithN` |

> 设计原则与完整形态表见 [**PARITY.md**](docs/PARITY.md)。

---

## Packages

访问方式：`client.<Field>`（或各包 `New`）。

| `Client` 字段 | 包 | 能力 |
|---------------|-----|------|
| `Chat` | `chat` | 多轮 · 流式 · 工具 · Defer · Parse · 存储 / 压缩 |
| `Image` · `Video` | `image` · `video` | 生成 · 延长 · 轮询 · 存储选项 |
| `Files` | `files` | 上传 · 列表 · 公开 URL · `ContentWriter` |
| `Batch` | `batch` | 批任务 · 类型化 `Result` |
| `Collections` | `collections` | 集合 / 文档 / 检索（CRUD 需 management key） |
| `Auth` · `Models` · `Tokenize` | 同名 | 密钥信息 · 模型目录 · 分词 |
| — | `tools` · `search` · `telemetry` · `types` | 工具 · Live Search · OTEL · 常量 |

```text
非目标   Embed client · OpenAI REST · OAuth · aio dual-stack
遥测     telemetry.Setup(...)     关闭: XAI_SDK_DISABLE_TRACING=1
```

---

## Documentation

| 需求 | 文档 |
|------|------|
| 函数 / 类型说明 | [**API.md**](docs/API.md) |
| 参数表（中英）与完整调用 | [**GUIDE.zh-en.md**](docs/GUIDE.zh-en.md) |
| 设计原则 · 推荐 API | [**PARITY.md**](docs/PARITY.md) |
| 与 Python SDK 差异 | [**DIFF.md**](docs/DIFF.md) |
| 稳定性 · protobuf | [COMPATIBILITY](docs/COMPATIBILITY.md) · [PROTO](docs/PROTO.md) |
| 安全披露 | [SECURITY.md](docs/SECURITY.md) |
| 发版 · CI · `dev` tip | [RELEASE](docs/RELEASE.md) · [CI](docs/CI.md) |
| 变更记录 | [CHANGELOG.md](CHANGELOG.md) |

```bash
go doc github.com/fun7257/xai-sdk-go
go doc github.com/fun7257/xai-sdk-go/chat.Chat.Sample
```

---

## Examples

| | |
|--|--|
| [**examples/complete**](examples/complete/) | 参数注释端到端（推荐先看） |
| [**examples/README.md**](examples/README.md) | 全目录索引 |
| [**examples/smoke**](examples/smoke/) | 真网低成本冒烟 |

```bash
export XAI_API_KEY=xai-...
go run ./examples/chat
go build ./examples/...
```

---

## Versioning & CI

| 通道 | 命令 | 说明 |
|------|------|------|
| **Release** | `go get …@vX.Y.Z` | 不可变 git tag |
| **Dev tip** | `go get …@dev` | main CI 全绿后覆盖 |
| **Runtime** | `xai.Version()` | 模块图解析，非手写常量 |

CI：PR · `main` · tag `v*` · 每日离线套件 · 可选真网 integration → 详见 [**CI.md**](docs/CI.md)。

---

## Contributing

见 [**CONTRIBUTING.md**](CONTRIBUTING.md)。安全问题请走 [**SECURITY.md**](docs/SECURITY.md)，勿公开 Issue 贴密钥。

```bash
make check && make race && make lint
```
