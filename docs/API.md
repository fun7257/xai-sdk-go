# API 参考 / Public API Reference

模块：`github.com/fun7257/xai-sdk-go`  
稳定性：[`COMPATIBILITY.md`](COMPATIBILITY.md) · 参数中英表：[`GUIDE.zh-en.md`](GUIDE.zh-en.md) · 设计：[`PARITY.md`](PARITY.md)

> 本文档覆盖**惯用公开 API**（非生成包 `xai/api/v1` 的完整 pb 清单）。  
> 生成类型以 `xaiv1` 形式出现在签名中时，字段级说明见 protos / `go doc xaiv1`。

---

## 1. 包一览

| Import 路径 | 职责 |
|-------------|------|
| `github.com/fun7257/xai-sdk-go` | 根入口 `NewClient`、选项、哨兵错误、`Version` |
| `.../chat` | 对话会话、补全、流式、工具结果、延迟、结构化 |
| `.../tools` | 服务端/客户端工具构造 |
| `.../search` | Live Search 参数与 Source |
| `.../image` | 图像生成 |
| `.../video` | 视频生成/延长 |
| `.../files` | 文件上传/下载/公开链接 |
| `.../batch` | 批处理任务与结果包装 |
| `.../collections` | 集合管理与文档检索（需 management key） |
| `.../auth` | API Key 元信息 |
| `.../models` | 模型列表/查询 |
| `.../tokenize` | 文本分词 |
| `.../telemetry` | 可选 OpenTelemetry |
| `.../types` | 稳定字符串常量（模型名等） |
| `.../xai/api/v1` | **生成** protobuf / gRPC（`xaiv1`） |

---

## 2. 根包 `xai`

### 2.1 构造与生命周期

| 符号 | 签名（概要） | 说明 |
|------|----------------|------|
| `NewClient` | `func NewClient(opts ...Option) (*Client, error)` | 创建客户端。默认读 `XAI_API_KEY`；可注入 conn/密钥/主机。 |
| `(*Client).Close` | `func (c *Client) Close() error` | 关闭**本 Client 拥有**的 gRPC 连接（注入的 conn 不关）。 |
| `Version` | `func Version() string` | 模块版本（git tag / 模块图）；本地未打 tag 多为 `"devel"`。 |

### 2.2 `Client` 域字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `Auth` | `*auth.Client` | 密钥信息 |
| `Chat` | `*chat.Client` | 对话 |
| `Image` | `*image.Client` | 图像 |
| `Video` | `*video.Client` | 视频 |
| `Files` | `*files.Client` | 文件 |
| `Batch` | `*batch.Client` | 批处理 |
| `Collections` | `*collections.Client` | 集合（CRUD 需 management 连接） |
| `Models` | `*models.Client` | 模型目录 |
| `Tokenize` | `*tokenize.Client` | 分词 |

### 2.3 选项 `Option`

| 函数 | 参数 | 说明 |
|------|------|------|
| `WithAPIKey` | `key string` | 业务密钥，覆盖 `XAI_API_KEY` |
| `WithManagementAPIKey` | `key string` | 管理密钥，覆盖 `XAI_MANAGEMENT_KEY` |
| `WithAPIHost` | `host string` | 业务 API，默认 `api.x.ai:443` |
| `WithManagementAPIHost` | `host string` | 管理 API，默认 `management-api.x.ai:443` |
| `WithTimeout` | `d time.Duration` | 无 deadline 时一元 RPC 默认超时（默认很长） |
| `WithInsecure` | — | 非 TLS 拨号（Bearer 仍发送） |
| `WithMetadata` | `kv ...string` | 额外 metadata，`k1,v1,k2,v2…`；不可覆盖 Authorization |
| `WithDialOptions` | `...grpc.DialOption` | 透传 gRPC 拨号选项 |
| `WithAPIConn` | `grpc.ClientConnInterface` | 注入业务连接（测试） |
| `WithManagementConn` | `grpc.ClientConnInterface` | 注入管理连接 |
| `WithoutEnv` | — | 不读环境变量中的密钥 |

### 2.4 哨兵错误

| 变量 | 何时 |
|------|------|
| `ErrNoAPIKey` | 无 API key（选项与环境皆无） |
| `ErrEmptyAPIKey` | 拨号使用空 key |
| `ErrNoManagementKey` | Collections CRUD 无 management 通道 |

使用：`errors.Is(err, xai.ErrNoAPIKey)`。

---

## 3. `chat`

### 3.1 客户端与会话

| 符号 | 说明 |
|------|------|
| `chat.New(cc)` | 底层 gRPC 包装（通常用 `client.Chat`） |
| `(*Client).Create(model, opts...)` | 创建多轮 `*Chat`（**不发 RPC**） |
| `(*Chat).Append(msg)` | 追加 `*Message` / 支持的消息形态 |
| `(*Chat).Messages()` | 当前消息快照 |
| `(*Chat).Sample(ctx)` | **单次**补全；`n>1` 应改用 `Samples` |
| `(*Chat).Samples(ctx, WithN(k))` | 多次补全，返回全部结果 |
| `(*Chat).StreamReader(ctx, opts...)` | 主流式：`Recv` 直到 `io.EOF`，务必 `Close` |
| `(*StreamReader).Recv` | 下一事件 `StreamEvent{Chunk, Response}` |
| `(*StreamReader).Close` | 结束流 / 取消 RPC |
| `(*StreamReader).Response` / `Responses` | 结束后聚合结果 |
| `(*Chat).Stream(ctx)` | channel 形态流式（次优选） |
| `(*Chat).Parse(ctx, schemaJSON, dest)` | JSON Schema 结构化输出并反序列化到 `dest` |
| `(*Chat).Defer` / `Defers` | 延迟补全；`WithDeferN` / 超时 / 轮询间隔 |
| `(*Chat).Compact` | 上下文压缩 |
| `(*Client).GetStoredCompletion` | 取已存响应 |
| `(*Client).DeleteStoredCompletion` | 删除已存响应 |
| `(*Client).CompactContext` | 独立压缩 RPC |
| `(*Chat).PrepareBatchRequest` | 构造批处理用 completion 请求（不发送） |
| `(*Chat).ConversationID` / `BatchRequestID` | 客户端侧关联 ID（不一定上线） |

### 3.2 会话选项 `chat.Option`（Create）

| 选项 | 说明 |
|------|------|
| `WithMessages` | 初始消息 |
| `WithUser` | 终端用户 ID |
| `WithMaxTokens` | 最大生成 token |
| `WithTemperature` / `WithTopP` | 采样 |
| `WithSeed` / `WithStop` | 种子 / 停止串 |
| `WithLogprobs` / `WithTopLogprobs` | 对数概率 |
| `WithFrequencyPenalty` / `WithPresencePenalty` | 惩罚项 |
| `WithReasoningEffort` | `none\|low\|medium\|high` |
| `WithTools` / `WithToolChoice` / `WithParallelToolCalls` | 工具 |
| `WithResponseFormat*` | text / json_object / json_schema |
| `WithSearchParameters` | Live Search（`search.Parameters`） |
| `WithStoreMessages` / `WithPreviousResponseID` | 存储与续写 |
| `WithUseEncryptedContent` / `WithMaxTurns` / `WithInclude` | 高级 |
| `WithAgentCount` | `4` 或 `16` |
| `WithServiceTier` | `default\|priority` |
| `WithConversationID` / `WithBatchRequestID` | 客户端记账 |

调用期：`chat.WithN(n)`（`Samples` / 多路流）。

延迟：`WithDeferTimeout` / `WithDeferInterval` / `WithDeferN`。

### 3.3 消息工厂

| 函数 | 错误行为 | 说明 |
|------|----------|------|
| `User` / `System` / `Assistant` / `Developer` | **非法 part panic** | 字面量 path；动态用 `New*` |
| `NewUser` / `NewSystem` / `NewAssistant` / `NewDeveloper` | 返回 `error` | 推荐生产动态内容 |
| `ToolResult(text, callID)` | — | 工具角色消息 |
| `Text` / `Image(url, detail)` | — | Content part；`detail`: auto/low/high |
| `FileByID` / `FileByData` / `FileByURL` | — | 文件 part |
| `File(...)` | `error` | 三选一：fileID / data / url |

### 3.4 响应

**`Response`**：`Content` / `ReasoningContent` / `EncryptedContent` / `FinishReason` / `ToolCalls` / `ToolOutputs` / `Citations` / `InlineCitations` / `Usage` / `CostUSD` / `ID` / `Proto` 等。

**`Chunk`**：流式 delta 对应方法。

**`CompactContextResponse`**：压缩结果字段访问。

---

## 4. `tools`

| 函数 | 返回 | 说明 |
|------|------|------|
| `WebSearch(opts...)` | `(*Tool, error)` | **主路径**；allowed/excluded 域名互斥 |
| `UncheckedWebSearch` | `*Tool` | 跳过校验 |
| `XSearch(opts...)` | `(*Tool, error)` | X 搜索；handle 互斥 |
| `UncheckedXSearch` | `*Tool` | 跳过校验 |
| `WebSearchChecked` / `XSearchChecked` | 同主路径 | **Deprecated** 别名 |
| `CodeExecution()` | `*Tool` | 代码执行 |
| `Function(name, desc, params)` | `(*Tool, error)` | 客户端函数；params 为 JSON 对象/string/bytes |
| `CollectionsSearch` / `CollectionsSearchOpts` | 工具 / error | 集合检索工具 |
| `MCP(url, opts...)` | `*Tool` | 远程 MCP |
| `RequiredTool(name)` | `*ToolChoice` | 强制某函数 |
| `Mode("auto\|none\|required")` | `*ToolChoice` | 工具模式 |
| `CallType(tc)` | `string` | 工具调用类型短名 |
| `ValidateWebSearchDomains` / `ValidateXSearchHandles` | `error` | 校验辅助 |

**WebSearch 选项**：`WithAllowedDomains` / `WithExcludedDomains` / `WithImageUnderstanding` / `WithImageSearch` / `WithUserLocation`。

**XSearch 选项**：`WithAllowedXHandles` / `WithExcludedXHandles` / `WithXDateRange` / `WithXMediaUnderstanding`。

**CollectionsSearch 选项**：`WithCollectionsLimit` / `WithCollectionsInstructions` / `WithCollectionsRetrievalMode`（hybrid/semantic/keyword）。

**MCP 选项**：`WithMCPLabel` / `Description` / `Auth` / `AllowedTools` / `Headers`。

---

## 5. `search`

| 符号 | 说明 |
|------|------|
| `Parameters` | Live Search 参数；`Proto()` → `*xaiv1.SearchParameters` |
| `Mode` | `ModeAuto` / `ModeOn` / `ModeOff` |
| `WebSource` / `NewsSource` / `XSource` | 校验源（域名/handle 互斥），返回 error |
| `UncheckedWebSource` / `UncheckedNewsSource` / `UncheckedXSource` | 跳过校验 |
| `RSSSource` | RSS 源 |
| `ValidateWebsites` / `ValidateXHandles` | 域名 / X handle 规则 |
| `MaxWebsites` | 最大网站数常量 `5` |
| `WithSafeSearch` 等 | Source 选项 |

---

## 6. `image`

| 方法 | 说明 |
|------|------|
| `Sample(ctx, prompt, model, opts...)` | 单图；`n>1` 请用 `Samples` |
| `Samples(...)` | 多图 |
| `Prepare(...)` | 只构造请求，不 RPC |

**选项**：`WithN` / `WithUser` / `WithImageURL(s)` / `WithImageFileID(s)` / `WithFormatURL` / `WithFormatBase64` / `WithAspectRatio` / `WithResolution` / `WithStorage`。

**Response**：`URL` / `Base64` / `Model` / `Prompt` / `RespectModeration` / `Usage` / `CostUSD` / `FileOutput` / `PublicURL` / `Download`（http(s)、体积上限 `MaxDownloadBytes`）。

---

## 7. `video`

| 方法 | 说明 |
|------|------|
| `Generate` | 文生视频并轮询至完成 |
| `Start` / `Get` | 异步：提交拿 request id / 查询状态 |
| `Extend` / `ExtendWith` | 延长视频（URL 或 file_id） |
| `ExtendStart` / `ExtendStartWith` | 仅提交延长 |
| `Prepare` / `PrepareExtension*` | 构造请求不发送 |

**选项**：`WithDuration` / `WithAspectRatio` / `WithResolution` / `WithImageURL` / `WithImageFileID` / `WithVideoURL` / `WithVideoFileID` / `WithReferenceImages` / `WithReferenceImageFileIDs` / `WithStorage` / `WithPollTimeout` / `WithPollInterval`。

**Response**：`URL` / `Duration` / `Model` / `Usage` / `CostUSD` / 存储与公开 URL 字段。  
**GenerationError**：失败时可用 `errors.As`。

---

## 8. `files`

| 方法 | 说明 |
|------|------|
| `Upload` / `UploadPath` | 分块上传（本地校验文件名与 TTL） |
| `List` / `ListAll` / `Get` / `Delete` | 元数据（`ListAll` 自动翻页） |
| `Content` | 下载到 `[]byte`（有 `MaxContentBytes`） |
| `ContentWriter` | 流式写入 `io.Writer`（大文件） |
| `CreatePublicURL` / `RevokePublicURL` | 公开链接 |
| `BatchUpload` / `BatchUploadWithOptions` / `BatchUploadItems` | 并发批量上传 |

**上传选项**：`WithExpiresAfter` / `WithProgress`。  
**列表选项**：`WithListLimit` / `PaginationToken` / `Order` / `SortBy` / `Filter`。  
**下载选项**：`WithContentFormat`（`original` / `text`）。  
**批量选项**：`WithBatchConcurrency` / `WithBatchOnComplete` / `WithBatchUploadOptions`。  
**StorageOptions**：图像/视频存储联动；`Proto` / `StorageFromProto`。

---

## 9. `batch`

| 方法 | 说明 |
|------|------|
| `Create` / `Add` / `Get` / `Cancel` / `List` / `ListAll` | 批任务生命周期（`ListAll` 自动翻页） |
| `ListBatchRequests` / `ListBatchResults` / `GetBatchRequestResult` | 明细与结果 |
| `ChatBatchRequest` / `ImageBatchRequest` / `VideoBatchRequest` | 包装单条请求 |
| `NewResult` / `WrapResults` | 结果包装 |
| `(*Result).Succeeded` / `Failed` / `Error` | 状态 |
| `ChatResponse` / `ImageResponse` / `VideoResponse` | 类型化结果 |
| `FilterSucceeded` / `FilterFailed` | 过滤 |

---

## 10. `collections`（管理密钥）

需 `WithManagementAPIKey` 或 `WithManagementConn`，否则 CRUD → `ErrNoManagementKey`。

| 方法 | 说明 |
|------|------|
| `Create` / `List` / `ListAll` / `Get` / `Delete` / `Update` | 集合（`ListAll` 自动翻页） |
| `GenerateDescription` | 生成描述 |
| `UploadDocument` / `AddExistingDocument` / `RemoveDocument` | 文档入出集 |
| `UpdateDocument` / `GetDocument` / `ListDocuments` / `ListAllDocuments` / `BatchGetDocuments` | 文档元数据（`ListAllDocuments` 自动翻页） |
| `ReindexDocument` / `WaitForIndexing` | 索引 |
| `Search` / `SearchSimple` | 检索（Documents 走业务 API） |

辅助：`FieldDefinition` / `AddFieldDefinition` / `DeleteFieldDefinition` / `ValidateChunkConfiguration`。  
Create 选项：`WithDescription` / `WithIndexModel` / `WithChunkConfiguration` / `WithMetric` / `WithFieldDefinitions`。  
Upload 选项：`WithDocumentFields` / `WithWaitForIndexing` / `WithUploadOptions` / `WithDeleteOnAddFailure`。  
Documents 列表选项：`WithDocumentsFilter` / `WithDocumentsName` / `WithDocumentsLimit` / `WithDocumentsPaginationToken`。  
Search 选项：`WithSearchLimit` / `WithSearchInstructions` / `WithRetrievalMode`。

---

## 11. `auth` / `models` / `tokenize`

| 包 | 方法 | 说明 |
|----|------|------|
| auth | `GetAPIKeyInfo(ctx)` | 当前连接密钥元数据 |
| models | `ListLanguageModels` / `GetLanguageModel` | 语言模型 |
| models | `ListEmbeddingModels` / `GetEmbeddingModel` | 嵌入**模型目录**（非 Embed RPC） |
| models | `ListImageGenerationModels` / `GetImageGenerationModel` | 图像生成模型 |
| tokenize | `TokenizeText(ctx, text, model)` | 返回 `[]*xaiv1.Token` |

---

## 12. `telemetry`

| 符号 | 说明 |
|------|------|
| `Setup(SetupOptions)` | 接入全局 TracerProvider（opt-in） |
| `Reset` / `Enabled` / `SetTracer` / `Tracer` | 状态 |
| `StartSpan` / `EndSpan` | 跨度 |
| `ChatSampleAttrs` / `ChatRequestAttrs` / `AttrPrompt` | 属性辅助 |
| `TracingDisabled` | 环境 `XAI_SDK_DISABLE_TRACING` |
| `SensitiveAttributesDisabled` | 环境 `XAI_SDK_DISABLE_SENSITIVE_TELEMETRY_ATTRIBUTES` |
| `SpanChatSample` 等常量 | span 名称 |

---

## 13. `types` 常量（常用）

| 类别 | 常量示例 |
|------|----------|
| Chat 模型 | `ModelGrok45`, `ModelGrok45Latest`, `ModelGrok420`, `ModelGrok3` |
| 图像 | `ModelImagineImage`, `ModelImagineImagePro`, `ModelImagineImageQuality` |
| 视频 | `ModelImagineVideo`, `ModelImagineVideo15`, `ModelImagineVideo15Preview` |
| 推理 | `ReasoningNone/Low/Medium/High` |
| 档位 | `ServiceTierDefault`, `ServiceTierPriority` |
| 图格式/比例 | `ImageFormatURL/Base64`, `Aspect1_1`, `Aspect16_9`, … |
| 视频 | `VideoAspect16_9`, `VideoRes480p`, `VideoRes720p` |
| 工具/搜索 | `ToolModeAuto/None/Required`, `SearchModeAuto/On/Off` |
| 图像细节 | `ImageDetailAuto/Low/High` |

---

## 14. 推荐调用路径（速查）

```text
NewClient → Chat.Create(model, WithMessages(...))
         → Sample / Samples(WithN) / StreamReader / Parse / Defer

NewClient → Image.Sample | Video.Generate
NewClient → Files.Upload | ContentWriter
NewClient → tools.WebSearch → Chat.WithTools
```

更完整参数与中英示例：[`GUIDE.zh-en.md`](GUIDE.zh-en.md) · [`examples/complete`](../examples/complete)。

---

## 15. 引用与 godoc

```bash
go doc github.com/fun7257/xai-sdk-go
go doc github.com/fun7257/xai-sdk-go/chat
go doc github.com/fun7257/xai-sdk-go/chat.Chat.Sample
# 生成类型
go doc github.com/fun7257/xai-sdk-go/xai/api/v1
```

pkg.go.dev（发版 tag 后）：  
`https://pkg.go.dev/github.com/fun7257/xai-sdk-go@v0.2.0`
