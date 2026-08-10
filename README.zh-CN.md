# xAI Go SDK

[**English**](README.md) · [**中文**](README.zh-CN.md)

[xAI API](https://docs.x.ai) 的惯用 **Go gRPC** 客户端。  
**优先 Go 调用体验** — 对齐产品能力，而不是克隆 Python API 形态。

<p align="left">
  <a href="https://pkg.go.dev/github.com/fun7257/xai-sdk-go"><img src="https://pkg.go.dev/badge/github.com/fun7257/xai-sdk-go.svg" alt="Go Reference"></a>
  <a href="https://github.com/fun7257/xai-sdk-go/actions/workflows/ci.yaml"><img src="https://github.com/fun7257/xai-sdk-go/actions/workflows/ci.yaml/badge.svg" alt="CI"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-Apache%202.0-blue.svg" alt="License"></a>
  <a href="https://go.dev/dl/"><img src="https://img.shields.io/badge/go-1.26%2B-00ADD8?logo=go" alt="Go 1.26+"></a>
</p>

```text
module   github.com/fun7257/xai-sdk-go
正式版   go get …@vX.Y.Z          # 不可变 tag
开发 tip go get …@dev             # 最新绿 main（会覆盖）
```

**[安装](#安装)** ·
**[快速开始](#快速开始)** ·
**[核心形态](#核心形态)** ·
**[包一览](#包一览)** ·
**[文档](#文档)** ·
**[示例](#示例)** ·
**[版本与 CI](#版本与-ci)** ·
**[贡献](#贡献)**

---

## 安装

```bash
go get github.com/fun7257/xai-sdk-go@v0.1.1
```

<details>
<summary><strong>凭证</strong> — 环境变量与选项</summary>

<br>

| | 环境变量 | 选项 | 用途 |
|--|----------|------|------|
| **业务密钥** | `XAI_API_KEY` | `WithAPIKey` | Chat · Image · Video · Files · Batch · Models · Tokenize · Auth · Search |
| **管理密钥** | `XAI_MANAGEMENT_KEY` | `WithManagementAPIKey` | Collections CRUD / 索引 |

- RPC 头：`Authorization: Bearer <key>`
- 优先环境变量，勿把密钥提交进仓库
- 可判定错误：`ErrNoAPIKey` · `ErrEmptyAPIKey` · `ErrNoManagementKey`（`errors.Is`）

</details>

---

## 快速开始

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
	client, err := xai.NewClient() // 读取 XAI_API_KEY
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

## 核心形态

| | 优先使用 | 说明 |
|--|----------|------|
| **单次补全** | `Sample(ctx)` | `n > 1` 请用 `Samples` |
| **多次补全** | `Samples(ctx, chat.WithN(k))` | 不静默丢弃结果 |
| **流式** | `StreamReader` → `Recv` / `Close` | 务必 `Close` |
| **结构化** | `Parse(ctx, schema, &dest)` | JSON Schema |
| **工具** | `tools.WebSearch(...) (*Tool, error)` | 主路径校验 · `Unchecked*` 逃生舱 |
| **图像 / 视频** | `Image.Sample` · `Video.Generate` | 多图：`Samples` + `WithN` |

> 设计原则与完整形态表 → [**PARITY.md**](docs/PARITY.md)。

---

## 包一览

访问：`client.<字段>`（或各包 `New`）。

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
非目标   Embed 客户端 · OpenAI REST · OAuth · aio 双栈
遥测     telemetry.Setup(...)     关闭: XAI_SDK_DISABLE_TRACING=1
```

---

## 文档

| 需求 | 文档 |
|------|------|
| 函数 / 类型完备说明 | [**API.md**](docs/API.md) |
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

## 示例

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

## 版本与 CI

| 通道 | 命令 | 说明 |
|------|------|------|
| **正式版** | `go get …@vX.Y.Z` | 不可变 git tag |
| **开发 tip** | `go get …@dev` | main CI 全绿后覆盖 |
| **运行时** | `xai.Version()` | 来自模块图，非手写常量 |

CI：PR · `main` · tag `v*` · 每日离线套件 · 可选真网 integration → 详见 [**CI.md**](docs/CI.md)。

---

## 贡献

见 [**CONTRIBUTING.md**](CONTRIBUTING.md)。安全问题请走 [**SECURITY.md**](docs/SECURITY.md)，勿在公开 Issue 中粘贴密钥。

```bash
make check && make race && make lint
```
