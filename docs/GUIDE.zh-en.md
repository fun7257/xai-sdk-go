# Complete usage guide / 完整调用指南

**EN** Parameter reference and call patterns for `github.com/fun7257/xai-sdk-go`.  
**中文** 本模块的参数说明与调用形态参考。

Landing / 入口: [`../README.md`](../README.md) · Design / 设计: [`PARITY.md`](PARITY.md) · Diffs: [`DIFF.md`](DIFF.md)

| | EN | 中文 |
|--|----|------|
| Runnable example | [`examples/complete/`](../examples/complete/) | 可运行完整示例 |
| Stability | [`COMPATIBILITY.md`](COMPATIBILITY.md) | API 稳定性 |
| Contribute / gates | [`../CONTRIBUTING.md`](../CONTRIBUTING.md) | 贡献与质量门禁 |

---

## 1. Install / 安装

```bash
go get github.com/fun7257/xai-sdk-go   # Go 1.26+
```

Quick start also on the [root README](../README.md).

---

## 2. Credentials / 凭证

| Kind / 类型 | Environment / 环境变量 | Option / 选项 | Used for / 用途 |
|-------------|------------------------|---------------|-----------------|
| Business API key / 业务密钥 | `XAI_API_KEY` | `xai.WithAPIKey` | Chat, Image, Video, Files, Batch, Models, Tokenize, Auth, document search |
| Management key / 管理密钥 | `XAI_MANAGEMENT_KEY` | `xai.WithManagementAPIKey` | Collections CRUD / indexing |

Every RPC sends `authorization: Bearer <key>`.  
每次 RPC 都会附带 `authorization: Bearer <密钥>`。

**EN** Never commit keys. Prefer env vars.  
**中文** 切勿把密钥写进仓库；优先使用环境变量。

Sentinel errors (usable with `errors.Is`):  
可用 `errors.Is` 判断的哨兵错误：

| Error | When / 何时 |
|-------|-------------|
| `xai.ErrNoAPIKey` | No key from options or env / 选项与环境中都没有密钥 |
| `xai.ErrEmptyAPIKey` | Empty key used for dial / 拨号时密钥为空 |
| `xai.ErrNoManagementKey` | Collections CRUD without management channel / 集合 CRUD 未配置管理通道 |

---

## 3. Client construction / 客户端构造

### Minimal / 最简

```go
client, err := xai.NewClient() // reads XAI_API_KEY / 读取环境变量
if err != nil { /* handle */ }
defer client.Close()
```

### Explicit options / 显式选项

```go
client, err := xai.NewClient(
    xai.WithAPIKey("xai-..."),                 // overrides env / 覆盖环境变量
    xai.WithManagementAPIKey("xai-..."),       // optional / 可选
    xai.WithTimeout(5*time.Minute),            // unary default timeout / 一元 RPC 默认超时
    // xai.WithAPIHost("api.x.ai:443"),
    // xai.WithInsecure(),                     // local dial only / 仅本地调试
)
```

### Client options reference / 客户端参数一览

| Option | Type | Default / 默认 | Description EN | 说明 中文 |
|--------|------|----------------|----------------|-----------|
| `WithAPIKey(key)` | `string` | env `XAI_API_KEY` | Business API key | 业务 API 密钥 |
| `WithManagementAPIKey(key)` | `string` | env `XAI_MANAGEMENT_KEY` | Management API key | 管理 API 密钥 |
| `WithAPIHost(host)` | `string` | `api.x.ai:443` | gRPC target for business API | 业务 API 拨号地址 |
| `WithManagementAPIHost(host)` | `string` | `management-api.x.ai:443` | Management API target | 管理 API 拨号地址 |
| `WithTimeout(d)` | `time.Duration` | ~27m | Default per-RPC timeout if ctx has no deadline | 上下文无截止时间时的默认 RPC 超时 |
| `WithInsecure()` | — | TLS on | Dial without TLS (Bearer still sent) | 明文拨号（仍带 Bearer） |
| `WithMetadata(k,v,...)` | pairs | — | Extra gRPC metadata | 额外 gRPC 元数据 |
| `WithDialOptions(...)` | `grpc.DialOption` | — | Raw dial options | 原始拨号选项 |
| `WithAPIConn(cc)` | `grpc.ClientConnInterface` | — | Inject connection (tests) | 注入连接（测试） |
| `WithManagementConn(cc)` | same | — | Inject management connection | 注入管理连接 |
| `WithoutEnv()` | — | env on | Do not read keys from environment | 不从环境变量读密钥 |

**Domain fields / 域客户端字段:**  
`client.Auth` · `Chat` · `Image` · `Video` · `Files` · `Batch` · `Collections` · `Models` · `Tokenize`

---

## 4. Chat — complete flow / 对话完整流程

### Preferred call shapes / 推荐调用形态

| Goal / 目标 | API | Notes / 说明 |
|-------------|-----|--------------|
| One completion / 单次补全 | `Sample(ctx)` | Rejects n>1 / n>1 会报错 |
| Multiple / 多次补全 | `Samples(ctx, chat.WithN(k))` | Returns all / 返回全部结果 |
| Streaming / 流式 | `StreamReader(ctx)` + `Recv` / `Close` | Primary stream DX / 主流式 API |
| Structured JSON / 结构化 | `Parse(ctx, schema, &dest)` | Unmarshals into dest / 反序列化到目标 |
| Deferred / 延迟 | `Defer` / `Defers(..., WithDeferN)` | Poll until done / 轮询至完成 |

### Session create options / 会话创建选项

Used in `client.Chat.Create(model, opts...)`.  
用于 `client.Chat.Create(model, opts...)`。

| Option | Parameter / 参数 | Description EN | 说明 中文 |
|--------|------------------|----------------|-----------|
| `WithMessages(msgs...)` | `*xaiv1.Message` | Initial messages | 初始消息列表 |
| `WithUser(u)` | `string` | End-user id for abuse tracking | 终端用户 ID（滥用追踪） |
| `WithMaxTokens(n)` | `int32` | Max completion tokens | 最大生成 token 数 |
| `WithTemperature(t)` | `float32` | Sampling temperature (higher → more random) | 采样温度（越高越随机） |
| `WithTopP(v)` | `float32` | Nucleus sampling | 核采样阈值 |
| `WithSeed(n)` | `int32` | Deterministic seed when supported | 可复现种子（若服务支持） |
| `WithStop(s...)` | `string` | Stop sequences | 停止序列 |
| `WithLogprobs(v)` | `bool` | Request log probabilities | 请求 logprobs |
| `WithTopLogprobs(n)` | `int32` | Top-N logprobs | Top-N logprobs |
| `WithFrequencyPenalty(v)` | `float32` | Frequency penalty | 频率惩罚 |
| `WithPresencePenalty(v)` | `float32` | Presence penalty | 存在惩罚 |
| `WithReasoningEffort(s)` | `none\|low\|medium\|high` | Reasoning depth | 推理强度 |
| `WithTools(t...)` | `*xaiv1.Tool` | Server/client tools | 服务端/客户端工具 |
| `WithToolChoice(tc)` | `*xaiv1.ToolChoice` | Force / auto / none | 工具选择策略 |
| `WithParallelToolCalls(v)` | `bool` | Allow parallel tool calls | 是否允许并行工具调用 |
| `WithResponseFormatText()` | — | Plain text | 纯文本输出 |
| `WithResponseFormatJSONObject()` | — | JSON object (no schema) | JSON 对象（无 schema） |
| `WithResponseFormatJSONSchema(b)` | `[]byte` | JSON Schema | JSON Schema 结构化 |
| `WithSearchParameters(p)` | `search.Parameters` | Live Search params | 实时搜索参数 |
| `WithStoreMessages(v)` | `bool` | Persist for later retrieve | 是否存储消息 |
| `WithPreviousResponseID(id)` | `string` | Continue from stored id | 从已存响应续写 |
| `WithUseEncryptedContent(v)` | `bool` | Encrypted content path | 加密内容路径 |
| `WithMaxTurns(n)` | `int32` | Max agentic turns | 最大代理轮次 |
| `WithInclude(opts...)` | string names | Extra include flags (e.g. `inline_citations`) | 额外 include 标志 |
| `WithAgentCount(n)` | `4` or `16` | Parallel agents | 并行代理数 |
| `WithServiceTier(tier)` | `default\|priority` | Service tier | 服务档位 |
| `WithConversationID(id)` | `string` | **Client-side** correlation only | **仅客户端**会话关联 ID |
| `WithBatchRequestID(id)` | `string` | **Client-side** batch bookkeeping | **仅客户端**批处理记账 ID |

### Call-time options / 单次调用选项

| Option | Parameter | Description EN | 说明 中文 |
|--------|-----------|----------------|-----------|
| `chat.WithN(n)` | `int32` ≥ 1 | Number of completions for `Samples` / stream multi-index | `Samples` 或多路流的补全数量 |

### Message factories / 消息工厂

| Function | Panic? | Error path / 错误路径 | Description |
|----------|--------|----------------------|-------------|
| `chat.User(parts...)` | Yes on illegal part | `chat.NewUser` | User message / 用户消息 |
| `chat.System(parts...)` | Yes | `chat.NewSystem` | System / 系统 |
| `chat.Assistant(parts...)` | Yes | `chat.NewAssistant` | Assistant / 助手 |
| `chat.Developer(parts...)` | Yes | `chat.NewDeveloper` | Developer / 开发者 |
| `chat.ToolResult(text, callID)` | No | — | Tool result / 工具结果 |
| `chat.Text(s)` | No | — | Text content part / 文本 part |
| `chat.Image(url, detail)` | No | — | Image URL; detail `auto\|low\|high` |
| `chat.FileByID` / `FileByData` / `FileByURL` | No | — | File attachments / 文件附件 |

**EN** Use panic factories only with literal safe parts; use `New*` when parts are dynamic.  
**中文** 字面量安全 part 可用会 panic 的工厂；动态输入请用 `New*`。

### Response helpers / 响应访问

| Method | EN | 中文 |
|--------|----|------|
| `Content()` | Assistant text | 助手文本 |
| `ReasoningContent()` | Reasoning trace | 推理内容 |
| `ToolCalls()` | Tool calls for index | 工具调用 |
| `Usage()` / `CostUSD()` | Token usage / cost | 用量 / 费用 |
| `Proto()` | Escape hatch to protobuf | 底层 protobuf |

---

## 5. Tools / 工具

| Constructor | Validates? | Description EN | 说明 中文 |
|-------------|------------|----------------|-----------|
| `tools.WebSearch(opts...)` | Yes → `error` | Web search tool | 网页搜索 |
| `tools.UncheckedWebSearch` | No | Escape hatch | 跳过校验 |
| `tools.XSearch(opts...)` | Yes | X (Twitter) search | X 搜索 |
| `tools.CodeExecution()` | — | Code execution | 代码执行 |
| `tools.Function(name, desc, params)` | marshal | Client function tool | 客户端函数工具 |
| `tools.CollectionsSearchOpts(...)` | mode | RAG collections tool | 集合检索工具 |
| `tools.MCP(url, opts...)` | — | Remote MCP | 远程 MCP |

WebSearch options examples / 示例：  
`WithAllowedDomains` · `WithExcludedDomains` (mutually exclusive / **互斥**) · `WithUserLocation` · `WithImageSearch`

---

## 6. Image / Video (brief) / 图像与视频（简）

**Image / 图像**

```go
img, err := client.Image.Sample(ctx, "a cat", "grok-imagine-image")
// multi: client.Image.Samples(ctx, prompt, model, image.WithN(2))
url, err := img.URL()
```

Common options: `WithN`, aspect ratio, resolution, image URL / file_id references (see package godoc).  
常用：`WithN`、宽高比、分辨率、参考图 URL / file_id（见包文档）。

**Video / 视频**

```go
// Generate + poll pattern — see examples/video
```

---

## 7. End-to-end example / 端到端示例

See the full annotated program:

**EN / 中文:** [`examples/complete/main.go`](../examples/complete/main.go)

```bash
export XAI_API_KEY=xai-...
# optional / 可选:
# export XAI_MODEL=grok-3
# export COMPLETE_SKIP_STREAM=1
# export COMPLETE_SKIP_TOOLS=1
go run ./examples/complete
```

What it demonstrates / 演示内容:

1. Client construction with documented options / 带参数说明的客户端构造  
2. Chat `Sample` with temperature, max tokens, system+user / 带温度与 token 限制的 Sample  
3. Streaming via `StreamReader` / 流式 `StreamReader`  
4. Optional server-side `WebSearch` tool / 可选服务端网页搜索工具  
5. Reading `Content`, usage, version / 读取内容、用量、SDK 版本  

---

## 8. Quality gates & version / 质量门禁与版本

Offline gates / 离线门禁: [`../CONTRIBUTING.md`](../CONTRIBUTING.md) (`make check`, lint, vuln, race).  
Live smoke / 真网可选: `make integration` (needs `XAI_API_KEY`).  
Version / 版本: **`xai.Version`** only — [`RELEASE.md`](RELEASE.md).
