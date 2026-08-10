# Go SDK vs Python `xai-sdk` — difference guide

**Audience:** library authors and app developers comparing this module to the official Python SDK.  
**Module:** `github.com/fun7257/xai-sdk-go`  
**Python baseline:** product surface of official `xai-sdk` (sync client; aio is intentionally not mirrored)  
**Related:** design principles in [`PARITY.md`](PARITY.md) · proto residual in [`PROTO.md`](PROTO.md)

---

## 0. How to read this document

Differences fall into **four kinds**. Do not mix them:

| Kind | Meaning | Action |
|------|---------|--------|
| **A. Intentional Go design** | Same product capability, different call shape for better Go UX / safety | Prefer Go form; do not “fix” to look like Python |
| **B. Product capability** | RPC/feature supported in both (or noted as partial) | Use the Go preferred entry point |
| **C. Proto residual** | Go (and Python) may use fields not yet in public `xai-org/xai-proto` | Functional OK; contract risk until upstream catches up |
| **D. Out of scope** | Explicit non-goals for this repo | Will not be implemented |

**Success metric for this SDK:** product capability + idiomatic Go call experience — **not** cloning Python names, dual stacks, or class hierarchies.

---

## 1. Design principles (why Go differs)

1. **Context first** — every RPC takes `context.Context`; cancellation is not a second client type.  
2. **Functional options** — `WithTemperature`, `WithN`, etc., instead of large kwargs.  
3. **One primary name family** — short path for n=1 (`Sample`), plural + options for multi (`Samples` + `WithN`).  
4. **No silent data loss** — multi-result APIs never keep only index 0 without saying so; singular APIs **reject** n>1 with an error pointing to the plural form.  
5. **Validate on the primary path** — illegal domain/handle combos fail when constructing tools/sources; `Unchecked*` is an explicit escape hatch.  
6. **Explicit wrappers** — methods like `Content()`, `URL()`, not dynamic attribute access.  
7. **No global side effects by default** — OTEL only after `telemetry.Setup`.  
8. **Errors as values** — domain errors (`video.GenerationError`) work with `errors.As` / `errors.Is` patterns where applicable.

---

## 2. Quick mapping (Python → Go)

| Python (sync, conceptual) | Preferred Go |
|---------------------------|--------------|
| `Client()` / env keys | `xai.NewClient(opts...)` |
| `client.chat.create(model, **kwargs)` | `client.Chat.Create(model, chat.With...)` |
| `chat.sample()` | `chat.Sample(ctx)` |
| `chat.sample_batch(n)` | `chat.Samples(ctx, chat.WithN(n))` |
| `chat.stream()` / iterate | `StreamReader` + `Recv` / `Close` (or `Stream` channels) |
| `chat.stream_batch(n)` | `StreamReader(ctx, chat.WithN(n))` |
| `chat.defer()` / `defer_batch` | `Defer` / `Defers(..., chat.WithDeferN(n))` |
| `chat.parse(PydanticModel)` | `Parse(ctx, schemaJSON, &dest)` |
| `user()` / `system()` / `image()` / `file()` | `chat.User` / `System` / `Image` / `File` / `FileByID`… |
| `web_search()` / `x_search()` | `tools.WebSearch` / `XSearch` → `(*Tool, error)` |
| `SearchParameters` + sources | `search.Parameters` + `WebSource` / `XSource`… |
| `client.image.sample` / batch | `Image.Sample` / `Image.Samples` + `WithN` |
| `client.video.generate` / `extend` / start | `Video.Generate` / `Extend` / `ExtendStart` + `Get` |
| `client.files.*` | `Files.Upload` / `BatchUploadItems` / `Content` / `ContentWriter` |
| `client.batch.*` | `Batch.*` + `batch.Result` helpers |
| `client.collections.*` | `Collections.*` (management key for CRUD) |
| OTEL (often ambient) | `telemetry.Setup` opt-in |

---

## 3. Client, auth, connection

### Capability

| Topic | Status |
|-------|--------|
| Business API key (`XAI_API_KEY` / `WithAPIKey`) | yes |
| Management key (`XAI_MANAGEMENT_KEY` / `WithManagementAPIKey`) | yes — Collections CRUD |
| Dual hosts (api / management) | yes |
| TLS default; `WithInsecure` | yes |
| Bearer metadata; user metadata cannot clobber Authorization | yes |
| Default timeout / keepalive / UNAVAILABLE retry | yes (internal dial) |
| SDK version/language metadata | yes |
| `Close()` | yes |
| Injected test conns (`WithAPIConn` / `WithManagementConn`) | yes |
| `WithoutEnv` for pure-option construction | yes |

### Intentional differences

| Python | Go | Rationale |
|--------|-----|-----------|
| Often implicit env + constructors | `NewClient` + options; env is default helper | Explicit DI for tests |
| sync Client vs AsyncClient | **One** `Client` | Go concurrency = context + goroutines |
| Channel details hidden | Same; options for dial extras | Keep public API small |

### Out of scope

- OAuth / cookie / password  
- Auto-loading `.env` files  

---

## 4. Chat

### Capability

| Feature | Go |
|---------|----|
| Stateful multi-turn session | `Chat` via `Create` (no RPC until Sample/Stream/…) |
| Create options (temp, tools, formats, store, previous id, include, agent count, service tier, …) | functional `With*` options |
| Append message / response / compaction | `Append` |
| Sample n=1 | `Sample(ctx)` |
| Sample n>1 | `Samples(ctx, WithN(k))` |
| Stream | `StreamReader` (primary), `Stream` (channels) |
| Stream multi | `StreamReader(ctx, WithN(k))` → `Responses()` after EOF |
| Deferred | `Defer` (n=1) / `Defers` + `WithDeferN` |
| Structured output | `Parse(ctx, schemaJSON, dest)` or create-time `WithResponseFormatJSONSchema` |
| Stored completion | client `GetStoredCompletion` / `DeleteStoredCompletion` |
| Compact | `Compact` / client `CompactContext` |
| Response accessors | Content, Reasoning, Encrypted, ToolCalls, ToolOutputs, Citations, InlineCitations, Logprobs, Usage, CostUSD, ServerSideToolUsage, Settings, DebugOutput, … |
| Chunk accessors | Content, Reasoning, ToolCalls, Encrypted, InlineCitations, ToolOutputs, ServerSideToolUsage, DebugOutput, … |
| Not concurrent-safe | documented on `Chat` |

### Preferred call shapes

```go
// Single completion (most common)
resp, err := session.Sample(ctx)

// Multi completion — never drops extras
resps, err := session.Samples(ctx, chat.WithN(3))

// Stream (always Close)
sr, err := session.StreamReader(ctx) // or chat.WithN(2)
defer sr.Close()
for {
    ev, err := sr.Recv()
    if err == io.EOF { break }
    // ev.Chunk / ev.Response
}

// Deferred
resp, err := session.Defer(ctx)
// multi deferred:
resps, err := session.Defers(ctx, chat.WithDeferN(2), chat.WithDeferTimeout(10*time.Minute))
```

### Intentional differences

| Python | Go | Rationale |
|--------|-----|-----------|
| `sample_batch` / `stream_batch` / `defer_batch` names | `Samples` / `StreamReader`+`WithN` / `Defers`+`WithDeferN` | Clear English; avoid deprecated-style `*_batch` |
| Single method often returns one object even for multi | Singular APIs **error** if n>1; plural returns `[]*Response` | **No silent discard** of completion indices |
| `parse(BaseModel)` | schema bytes + `json.Unmarshal` into `dest` | No forced schema library |
| Stream as iterator / async generator | `Recv` loop or channels | Idiomatic Go streaming |
| Dynamic response attributes | Explicit methods + `Proto()` | Static typing, godoc, tooling |
| `conversation_id` often for tracing | `WithConversationID` client-side + OTEL attr; **not** on chat request proto | Matches public chat.proto field set |
| `batch_request_id` on create | Client-side + `PrepareBatchRequest` + `batch.ChatBatchRequest` | Batch id lives on batch envelope |

### Safety rules (Go-specific)

| Call | n>1 behavior |
|------|----------------|
| `Sample` | n/a (always 1) |
| `Samples` | returns all indices |
| `Defer` | **errors** if `WithDeferN(k)` and k>1 → use `Defers` |
| `Defers` | returns all indices |
| `image.Sample` | **errors** if `WithN(k)` and k>1 → use `Samples` |

### Message factories

| Python | Go |
|--------|-----|
| `user(...)` / `system(...)` / … | `chat.User` / `System` / `Assistant` / `Developer` (panic on bad part type) + `NewUser`… returning error |
| `text` / `image` / `file` overloads | `Text`, `Image(url, detail)`, `FileByID` / `FileByData` / `FileByURL`, or unified `File(...)` with exactly-one-mode validation |
| Image detail auto/low/high | same string details |

### Partial / notes

- Schema-from-struct reflection is **not** built-in (callers generate JSON Schema).  
- `ToolOutputs` on multi-index responses: filtered by index when set; nil index merges tool-role outputs (server multi-tool mode).  

---

## 5. Tools

### Capability

| Tool | Go |
|------|-----|
| Client function | `tools.Function(name, desc, parametersJSON/map)` |
| Tool choice | `RequiredTool`, `Mode` |
| Web search | `WebSearch(opts...) (*Tool, error)` |
| X search | `XSearch(opts...) (*Tool, error)` |
| Code execution | `CodeExecution()` |
| Collections search | `CollectionsSearch` / `CollectionsSearchOpts` |
| MCP | `MCP(url, opts...)` |
| Call type helper | `CallType(tc)` |

### Intentional differences

| Python | Go | Rationale |
|--------|-----|-----------|
| Constructor may raise ValueError on domain mutex | Primary API returns `error` | Go style; works in options pipelines |
| Single constructor often “just works” or raises | **Primary validates**; `UncheckedWebSearch` / `UncheckedXSearch` for escape | Fail-closed by default without burying safety behind `*Checked` names |
| `web_search` kwargs | `WithAllowedDomains`, `WithExcludedDomains`, location, image flags | Options composition |

### Deprecated aliases

- `WebSearchChecked` / `XSearchChecked` → aliases of validating `WebSearch` / `XSearch` (prefer the short names).

---

## 6. Search (Live Search parameters)

### Capability

| Feature | Go |
|---------|-----|
| Parameters (mode, dates, citations, max results) | `search.Parameters` + `Proto()` |
| Web / News / X / RSS sources | `WebSource`, `NewsSource`, `XSource`, `RSSSource` |
| safe_search | `WithSafeSearch` (default true) |
| X favorite/view thresholds | `WithPostFavoriteCount`, `WithPostViewCount` |
| Domain mutex / max 5 sites | validated by default on Web/News |

### Intentional differences

| Python | Go | Rationale |
|--------|-----|-----------|
| Source factories return objects | Primary returns `(*Source, error)` | Validation at construction |
| Optional strictness | Default validate; `WithoutDomainValidation` / `Unchecked*` | Explicit escape hatch |

---

## 7. Image

### Capability

| Feature | Go |
|---------|-----|
| Generate | `Sample` / `Samples` |
| n images | `WithN` |
| URL / base64 format | `WithFormatURL` / `WithFormatBase64` (default URL) |
| Reference URL / file_id (single & multi) | options + mutual exclusion rules |
| Aspect / resolution | `WithAspectRatio` (unknown → error), `WithResolution` |
| Storage options | `WithStorage` |
| Response URL/Base64/moderation/storage/public URL | accessors |
| Download bytes | `Download` (https/http, size cap) |
| Batch prepare | `Prepare` |

### Preferred call shapes

```go
// One image
img, err := client.Image.Sample(ctx, "a cat", model)

// Multiple images (no silent drop)
imgs, err := client.Image.Samples(ctx, "a cat", model, image.WithN(2))
```

### Intentional differences

| Python | Go | Rationale |
|--------|-----|-----------|
| `sample` + `sample_batch` | `Sample` + `Samples` | Go plural convention |
| Multi may return sequence from batch API | `Samples` always returns full slice | Clear multi-result type |
| `Sample` with n>1 might be ambiguous | **errors** → use `Samples` | Safety |
| Response `prompt` property from server | `Prompt()` is client-cached request prompt | Public ImageResponse proto has no prompt field |

---

## 8. Video

### Capability

| Feature | Go |
|---------|-----|
| Generate + poll | `Generate` |
| Low-level start/get | `Start`, `Get` |
| Extend + poll | `Extend` / `ExtendWith` |
| Extend start only | `ExtendStart` / `ExtendStartWith` |
| T2V / I2V / R2V / file_id inputs | options + validation |
| Duration, aspect, resolution, storage | options |
| Poll timeout/interval | `WithPollTimeout` / `WithPollInterval` |
| FAILED → domain error | `GenerationError` |
| DONE empty body | descriptive error |
| Response URL / moderation / storage | accessors |

### Intentional differences

| Python | Go | Rationale |
|--------|-----|-----------|
| `extend_start` naming | `ExtendStart` | Go exported names |
| Poll helpers mixed with high-level generate | Explicit Start/Get vs Generate/Extend | Clear lifecycle for async jobs |

---

## 9. Files

### Capability

| Feature | Go |
|---------|-----|
| Chunked upload (3 MiB) | `Upload` / `UploadPath` |
| Progress | `WithProgress(func(delta, total int64))` |
| Expires | `WithExpiresAfter` |
| List filter/sort/order/limit/page | list options |
| Get / Delete | yes |
| Content full buffer | `Content` (max `MaxContentBytes`) |
| Content stream | `ContentWriter(ctx, id, io.Writer)` |
| Public URL create/revoke | yes |
| Batch upload paths | `BatchUpload` / `BatchUploadWithOptions` |
| Batch path + reader items | `BatchUploadItems` + `BatchItem` |

### Intentional differences

| Python | Go | Rationale |
|--------|-----|-----------|
| tqdm / ProgressBarLike | simple progress callback | No Python UI dependency |
| `str \| BinaryIO` batch | `BatchItem{Path}` or `{Reader,Name,Size}` | Explicit union via struct |
| Full buffer content | both buffer and `ContentWriter` | Large-file friendly Go I/O |

---

## 10. Batch

### Capability

| Feature | Go |
|---------|-----|
| Create / Add / Get / Cancel / List | yes |
| List requests / results + pagination | yes |
| Get single request result | `GetBatchRequestResult` |
| Typed success/failure | `batch.Result` (`Succeeded`/`Failed`/`Error`) |
| Unpack chat/image/video | `ChatResponse` / `ImageResponse` / `VideoResponse` |
| Filters | `FilterSucceeded` / `FilterFailed` / `WrapResults` |
| Build batch envelopes | `ChatBatchRequest` / `ImageBatchRequest` / `VideoBatchRequest` |
| From chat session | `PrepareBatchRequest` + `BatchRequestID()` |

### Intentional differences

| Python | Go | Rationale |
|--------|-----|-----------|
| Properties on list response | helpers on `[]*Result` | Composition over inheritance |
| Chat create batch id | client-side id + prepare helper | Matches proto placement of batch_request_id |

---

## 11. Collections

### Capability

| Feature | Go |
|---------|-----|
| CRUD collection | yes (management key required) |
| Chunk config / metric / fields | create options + `ValidateChunkConfiguration` |
| Field definition helpers | `FieldDefinition`, `AddFieldDefinition`, `DeleteFieldDefinition` |
| Documents add/list/get/remove/batch get/reindex | yes |
| Upload path → add → optional wait | `UploadDocument` |
| Delete orphan file if add fails | `WithDeleteOnAddFailure` (opt-in) |
| Update document | `UpdateDocument` + options |
| Search (Documents service on API channel) | `Search` / `SearchSimple` |
| Wait for indexing | `WaitForIndexing` (FAILED → error) |
| Generate description | yes |

### Intentional differences

| Python | Go | Rationale |
|--------|-----|-----------|
| Management vs API keys often implicit | Explicit dual channel; clear error if management missing | Fail loud |
| Orphan file policy | Default leave file; optional delete | Explicit cleanup policy |
| Documents search client | Under `Collections` (same product area) | Fewer root clients |

---

## 12. Models, Auth, Tokenize

| Domain | Go | Notes |
|--------|-----|--------|
| Language / embedding / image models list+get | `client.Models` | Embedding **models** list only — not embed RPC client |
| Auth key info | `client.Auth.GetAPIKeyInfo` → `*ApiKey` proto | No property wrapper (raw proto is fine in Go) |
| Tokenize | `client.Tokenize.TokenizeText` | yes |

### Out of scope

- Embeddings **generation** client (product non-goal; Python also does not promote a full embed product client the same way)

---

## 13. Telemetry

### Capability

| Feature | Go |
|---------|-----|
| Default | noop tracer; no global provider install |
| Enable | `telemetry.Setup` / `SetTracer` |
| Env disable | `XAI_SDK_DISABLE_TRACING` |
| Sensitive attrs | opt-in include env; disable env still wins |
| Chat spans | sample / stream / parse; model, message count, conversation id, request attrs |
| Domain light spans | image sample, video generate/extend, files upload |

### Intentional differences

| Python | Go | Rationale |
|--------|-----|-----------|
| Often ambient / auto | Explicit Setup | Libraries should not steal global OTEL |
| Rich gen_ai matrix | Core attrs + extras; prompts not by default | Privacy + less surprise |
| Built-in OTLP helpers | Documented recipes (`Setup` + caller-supplied provider) | Avoid hard OTLP dependency |

---

## 14. Types / constants

| Go package `types` | Purpose |
|--------------------|---------|
| Model name constants | e.g. `ModelGrok3`, image/video model ids |
| Reasoning / service tier / tool mode / search mode / aspect | stable string constants |

Python often uses Literals / enums in typing; Go uses string constants + internal proto mapping.

---

## 15. Proto residual (kind C)

Public [xai-org/xai-proto](https://github.com/xai-org/xai-proto) can lag the Python SDK / live API. This repo keeps **local residual** patches (field numbers aligned with Python descriptors where applicable):

| Area | Residual examples |
|------|-------------------|
| Collections | management protos not fully public |
| Files | Public URL RPCs; list `filter` |
| Image / Video | `file_id`, `storage_options`, response storage |

**Implication:** functional behavior can match Python while **public proto repo** is incomplete. Re-sync when upstream publishes identical definitions; field-number drift is a release risk. Details: [`PROTO.md`](PROTO.md).

**Why Python can disagree with public xai-proto:** product SDKs track the **service wire** and ship descriptors on their release train; the public proto repo is a slower, documented subset.

---

## 16. Out of scope (kind D)

| Item | Reason |
|------|--------|
| Embed product client | Not a primary Python product surface here |
| OpenAI-compatible REST layer | Separate product |
| OAuth / non-API-key auth | Not the gRPC API Key contract |
| sync + aio dual stacks | Go uses context |
| Auto `.env` loading | App concern |
| Mirroring ProtoDecorator / Pydantic / ABC | Wrong Go model |
| Deprecated Python `*_batch` **names** | Capability via `Samples`/`WithN` etc. |

---

## 17. Examples matrix

Under `examples/` (env credentials only; never commit keys):

| Example | Demonstrates |
|---------|----------------|
| `chat`, `stream`, `function_calling`, `structured` | Core chat UX |
| `deferred`, `stored`, `compaction`, `reasoning` | Long-path chat |
| `server_side_tools`, `collections_tool` | Tools |
| `image`, `image_multi_reference`, `image_understanding` | Image |
| `video`, `video_extension` | Video (`ExtendStart`) |
| `files`, `files_chat`, `inline_files_chat` | Files + chat attachments |
| `batch` | List + `PrepareBatchRequest` / batch envelope |
| `collections` | Management key |
| `models`, `auth`, `tokenizer` | Utility clients |
| `telemetry`, `smoke` | Observability / multi-domain smoke |

---

## 18. Engineering notes (repo process)

| Topic | Status |
|-------|--------|
| Module path | `github.com/fun7257/xai-sdk-go` |
| Go version | 1.26+ |
| Tests | offline mock/bufconn preferred; live smoke optional |
| Version constant | `xai.Version` (e.g. `0.2.0`) |
| Git remote / tags | process-dependent (workspace may ship without `.git`) |

---

## 19. Migration cheat-sheet (from older Go dual names)

| Older / Python-shaped | Prefer now |
|-----------------------|------------|
| `SampleN(ctx, n)` | `Samples(ctx, WithN(n))` |
| `StreamReaderN(ctx, n)` | `StreamReader(ctx, WithN(n))` |
| `StreamN(ctx, n)` | `Stream(ctx, WithN(n))` |
| `DeferN(ctx, n, …)` | `Defers(ctx, WithDeferN(n), …)` |
| `SampleAll(...)` | `Samples(...)` |
| `WebSearchChecked` | `WebSearch` (same validation) |
| `XSearchChecked` | `XSearch` |
| Unvalidated tools/sources as default | Use primary constructors; `Unchecked*` only if intentional |

---

## 20. Summary

| Question | Answer |
|----------|--------|
| Do we match Python **product capability**? | **Yes** for the supported gRPC domains listed above (with residual proto risk noted). |
| Do we match Python **API shapes**? | **No — by design.** Go uses options, singular/plural multi APIs, validate-on-primary, explicit wrappers. |
| What is still “not aligned”? | Kind **C** (public proto residual), kind **D** (non-goals), and intentional kind **A** differences in this document — not unfinished P0 product gaps. |

For a short design summary, see [`PARITY.md`](PARITY.md). For wire residual detail, see [`PROTO.md`](PROTO.md).
