# xAI Go SDK

[xAI API](https://docs.x.ai) 的惯用 Go gRPC 客户端。**优先 Go 调用体验**，对齐产品能力，而非克隆 Python API 形态。

| | |
|--|--|
| **Module** | [`github.com/fun7257/xai-sdk-go`](https://pkg.go.dev/github.com/fun7257/xai-sdk-go) |
| **Go** | 1.26+ |
| **License** | Apache-2.0 |
| **Release** | Git tags `vX.Y.Z` · 开发 tip：`@dev` |

---

## Contents

1. [Install](#install)
2. [Quick start](#quick-start)
3. [Core patterns](#core-patterns)
4. [Packages](#packages)
5. [Documentation](#documentation)
6. [Examples](#examples)
7. [Versioning & CI](#versioning--ci)
8. [Contributing](#contributing)

---

## Install

```bash
go get github.com/fun7257/xai-sdk-go@v0.1.1   # pin a release
# go get github.com/fun7257/xai-sdk-go@dev    # latest green main (mutable)
```

### Credentials

| Key | Environment | Option | Scope |
|-----|-------------|--------|--------|
| Business | `XAI_API_KEY` | `WithAPIKey` | Chat, Image, Video, Files, Batch, Models, Tokenize, Auth, search |
| Management | `XAI_MANAGEMENT_KEY` | `WithManagementAPIKey` | Collections CRUD / indexing |

RPC 使用 `Authorization: Bearer <key>`。密钥走环境变量，勿提交仓库。  
可判定错误：`ErrNoAPIKey` · `ErrEmptyAPIKey` · `ErrNoManagementKey`（`errors.Is`）。

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
	client, err := xai.NewClient() // reads XAI_API_KEY
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
go run ./examples/complete   # 中英注释端到端示例
```

---

## Core patterns

| Goal | Prefer | Notes |
|------|--------|--------|
| 单次补全 | `Sample(ctx)` | `n>1` 请用 `Samples` |
| 多次补全 | `Samples(ctx, chat.WithN(k))` | 不静默丢弃结果 |
| 流式 | `StreamReader` → `Recv` / `Close` | 务必 `Close` |
| 结构化 | `Parse(ctx, schema, &dest)` | JSON Schema |
| 工具 | `tools.WebSearch(...)` → `(*Tool, error)` | 主路径校验；`Unchecked*` 为逃生舱 |
| 图像 / 视频 | `Image.Sample` · `Video.Generate` | 多图用 `Samples` + `WithN` |

设计原则与完整形态表 → [`docs/PARITY.md`](docs/PARITY.md)。

---

## Packages

通过 `client.<Field>` 访问（或各包 `New`）：

| Field | Package | Capability |
|-------|---------|------------|
| `Chat` | [`chat`](docs/API.md#3-chat) | 多轮、流式、工具、Defer、Parse、存储/压缩 |
| `Image` / `Video` | `image` / `video` | 生成、延长、轮询、存储选项 |
| `Files` | `files` | 上传、列表、公开 URL、大文件 `ContentWriter` |
| `Batch` | `batch` | 批任务 + 类型化 `Result` |
| `Collections` | `collections` | 集合/文档/检索（CRUD 需 management key） |
| `Auth` / `Models` / `Tokenize` | 同名包 | 密钥信息、模型目录、分词 |
| — | `tools` · `search` · `telemetry` · `types` | 工具、Live Search、可选 OTEL、常量 |

**非目标：** Embed 客户端 · OpenAI REST · OAuth · aio 双栈。

Telemetry（可选）：`telemetry.Setup(...)` · 关闭：`XAI_SDK_DISABLE_TRACING=1`。

---

## Documentation

| 你想… | 打开 |
|--------|------|
| **函数/类型完备说明** | [`docs/API.md`](docs/API.md) |
| **参数表（中英）+ 完整调用** | [`docs/GUIDE.zh-en.md`](docs/GUIDE.zh-en.md) |
| **设计原则 / 推荐 API** | [`docs/PARITY.md`](docs/PARITY.md) |
| **与 Python SDK 差异** | [`docs/DIFF.md`](docs/DIFF.md) |
| **稳定性 / protobuf 策略** | [`docs/COMPATIBILITY.md`](docs/COMPATIBILITY.md) · [`docs/PROTO.md`](docs/PROTO.md) |
| **安全披露** | [`docs/SECURITY.md`](docs/SECURITY.md) |
| **发版 / CI / 开发 tip** | [`docs/RELEASE.md`](docs/RELEASE.md) · [`docs/CI.md`](docs/CI.md) |
| **变更记录** | [`CHANGELOG.md`](CHANGELOG.md) |

命令行：`go doc github.com/fun7257/xai-sdk-go/...`

---

## Examples

| 路径 | 说明 |
|------|------|
| [`examples/complete`](examples/complete/) | 带参数注释的端到端（推荐先看） |
| [`examples/README.md`](examples/README.md) | 全目录索引（chat / stream / image / …） |
| [`examples/smoke`](examples/smoke/) | 真网低成本冒烟 |

```bash
export XAI_API_KEY=xai-...
go run ./examples/chat
go build ./examples/...
```

---

## Versioning & CI

- **正式版：** 不可变 git tag → `go get …@vX.Y.Z`
- **开发 tip：** 浮动 tag `dev`（main CI 全绿后覆盖）→ `go get …@dev`
- **CI：** PR / main / `v*` tag / 每日离线套件；可选真网 integration  
  细节 → [`docs/CI.md`](docs/CI.md)

运行时版本字符串：`xai.Version()`（来自模块图，非手写常量）。

---

## Contributing

见 [`CONTRIBUTING.md`](CONTRIBUTING.md)（`make check` / lint / race / vuln）。  
问题与安全：Issue 模板 · [`docs/SECURITY.md`](docs/SECURITY.md)。

```bash
make check && make race && make lint
```
