# Changelog

## Unreleased

### Fixed

- CI: install Go **1.26.6+** (`check-latest: true`) so daily `govulncheck` uses the stdlib security patch (GO-2026-6218 / 6090 / 5972 / 5026). `go.mod` now has `toolchain go1.26.6`.

### Docs

- README (en/zh): guidance on choosing between this gRPC SDK and xAI's OpenAI-compatible REST API (`base_url = https://api.x.ai/v1` with any OpenAI client). An in-SDK OpenAI compatibility layer was evaluated and deliberately not built — the official REST path already covers that need.

## [0.3.0] — 2026-08-12

Wire-alignment release: the vendored collections/files protos are now field-for-field identical to the official xAI API schema (#14), plus the remaining API-completeness gaps (#13) and automated release notes (#12).

### Fixed — wire alignment with the official API (breaking)

Vendored protos for collections/files had drifted from the official schema (verified against the official Python SDK `proto/v6` descriptors and their git history): reserved field numbers had been compacted, so several requests were misread or dropped server-side and several responses were misread client-side. Aligned:

- **CreateCollectionRequest**: `index_configuration`/`chunk_configuration`/`metric_space`/`field_definitions`/`collection_description` renumbered to the official `#4/#5/#7/#9/#10` (previously `#3–#7`; every optioned Create was scrambled on the wire). `metric_space`/`collection_description`/`team_id` are now proto3 optional.
- **CollectionMetadata**: `field_definitions`/`collection_description`/`total_file_size` now read from official `#8/#9/#10` (previously misread `#7–#9`).
- **AddDocumentToCollectionRequest**: `fields` moved to official `#5` (previously `#4`, silently dropped server-side).
- **ChunkConfiguration**: chars/tokens/bytes strategies are now a proto **oneof** (`chars=#1, tokens=#2, bytes=#11`) with `strip_whitespace=#3` / `inject_name_into_chunks=#4`, matching upstream (previously `bytes=#3, strip=#4, inject=#5` — all misaligned). `ValidateChunkConfiguration` now only rejects the none-set zero value (the oneof enforces at-most-one). `GenerateCollectionDescriptionResponse.description` renamed to `collection_description` (wire-compatible).
- **File**: gains `public_url` (#7) and `public_url_expires_at` (#8). **CreatePublicUrlResponse**: `#2` is `expires_at` (Timestamp) — the phantom `file_id` field never existed upstream. **RevokePublicUrlResponse**: gains `revoked` (#2) and `public_url` (#3).
- New `xai/api/v1/wire_pin_test.go` pins the exact field numbers of every previously-drifted message against the official layout; `docs/PROTO.md` documents the parity baseline and the no-compaction rule for reserved numbers.
- **collections**: new `WithChunkOverlap` (applies to the active chars/tokens/bytes strategy) and `WithTokensEncodingName` chunk options, exposing the official overlap/encoding parameters.

### Added

- **chat.DeferStart / chat.DeferGet**: split deferred-completion shape (submit → request id → poll on your own schedule, resumable across processes), matching the video Start/Get pattern. `Defer`/`Defers` remain the synchronous-polling path.
- **tools.AttachmentSearch**: constructor for the attachment search server-side tool (the last `Tool` oneof branch without one); pairs with `chat.WithInclude("attachment_search_call_output")`.
- **chat.Named**: sets the participant name on a message (`Message.name` was previously unreachable via builders).
- **collections chunk constructors**: `ChunkByChars` / `ChunkByTokens` / `ChunkByBytes` with `WithStripWhitespace` / `WithInjectNameIntoChunks`; results always pass `ValidateChunkConfiguration`.
- **release workflow** (`.github/workflows/release.yml`): pushing a `v*` tag now publishes a GitHub Release whose notes are the matching `CHANGELOG.md` section (auto-generated notes as fallback); manual dispatch backfills or refreshes notes for existing tags.

### Changed

- **video.Extend / PrepareExtension**: options without `ExtendVideoRequest` fields (`WithAspectRatio`, `WithResolution`, first-frame image, reference images, Generate-only video-source options) now return an error instead of being silently ignored.

## [0.2.0] — 2026-08-12

Audit-driven release: high/medium findings fixed (#9), low-severity follow-ups (#10), API completeness (#11), plus test-coverage and CI improvements.

### Breaking

- **search.XSource** now validates included/excluded handle mutual exclusion and returns `(*xaiv1.Source, error)`, matching `WebSource`/`NewsSource`; use `UncheckedXSource` for the previous no-error shape. Added `ValidateXHandles`.
- **Strict enum inputs**: `image.WithResolution`, `video.WithResolution`, and `collections.WithMetric` cause the call to return an error for unknown values (previously silent fallback), matching aspect-ratio handling. `tools.Mode` / `files.WithListSortBy` document their fallback.
- **files upload validation**: `Upload` rejects `WithExpiresAfter` values outside `[1h, 30d]` (`MinExpiresAfter` / `MaxExpiresAfter`) and filenames violating the API contract (≤ `MaxFilenameChars` (255) Unicode chars; no control characters, line terminators, `"`, `;`, `\`) before streaming; `CreatePublicURL` rejects sub-second TTLs. `xai.NewClient` fails fast on an odd number of `WithMetadata` elements.

### Added

- **Auto-pagination helpers**: `files.ListAll`, `collections.ListAll`, `collections.ListAllDocuments`, and `batch.ListAll` follow pagination tokens until the final page (with a defensive stop on non-advancing tokens).
- **collections.WithDocumentsPaginationToken**: resume `ListDocuments` from a prior page token (the proto field was previously unreachable via options).
- **files.WithContentFormat**: request `original` (raw bytes, default) or `text` (server-extracted text) on `Content` / `ContentWriter`; unknown values error. Both methods now accept variadic `ContentOption`s (source-compatible).
- **types**: completed aspect-ratio constants (`Aspect2_3`, `Aspect3_2`, video `1:1`/`4:3`/`3:4`/`3:2`/`2:3`), image resolutions (`ImageRes1K`/`ImageRes2K`), and content formats (`ContentFormatOriginal`/`ContentFormatText`); a test locks constants to their validators.

### Fixed

- **files.Upload**: a server-side stream abort (quota/size rejection) now surfaces the real gRPC status instead of a bare `EOF` (`Send` == `io.EOF` → status via `CloseAndRecv`).
- **chat streaming hardening**: `Response.ProcessChunk` and multi-response wrapping drop outputs with negative or absurdly large indexes instead of panicking / allocating unbounded memory on malformed stream data; response-level fields (id, model, usage, system_fingerprint, service_tier) are only overwritten by chunks that carry them.
- **chat.Parse**: session `ResponseFormat` restore now uses `defer`, so a panicking Sample cannot leave the session mutated; **chat.Compact** surfaces `Append` errors.
- **files.ContentWriter**: cancels the server stream promptly on early returns; **files.Content** returns no partial data alongside an error.
- **files.BatchUploadItems**: bounded worker pool (no longer one goroutine per item); `Path`+`Reader` both set is an error for that item; panics from an item upload or the `onComplete` callback are recovered into that item's error instead of crashing the process.
- **collections.UploadDocument**: `WithDeleteOnAddFailure` cleanup runs on a context detached from the caller's (30s bound), so orphan files are deleted even when the add failed due to cancellation/deadline; **collections.Update** no longer mutates the caller's request.
- **batch.Result.Error**: preserves structured status details (`status.FromProto`).
- **image.Response**: negative index degrades gracefully instead of panicking.
- **telemetry**: `SetTracer` toggles `Enabled()` in line with its docs; truthy env parsing accepts any casing of `true`; package docs cover `XAI_SDK_INCLUDE_SENSITIVE_TELEMETRY_ATTRIBUTES`.
- **tools**: `Function` validates pre-encoded parameters as JSON; `WithMCPHeaders` copies the map; `CollectionsSearch` never returns nil and copies the limit value; **search.Parameters.Proto** copies the Sources slice and MaxSearchResults value.

### Changed

- **tools bool options**: `WithImageUnderstanding(false)` / `WithImageSearch(false)` / `WithXMediaUnderstanding(false, false)` send an explicit `false` on the wire (previously a no-op); unset options remain absent.
- Docs: cleartext warnings on `xai.WithInsecure` and `tools.WithMCPAuth` / `tools.MCP`; `docs/SECURITY.md` response-expectations table; `examples/video` / `examples/deferred` bound polling with context timeouts; install pins updated.

### Tests / CI

- New coverage for previously untested surfaces: `image.Download` trust boundary (scheme allowlist, redirects, 100 MiB cap), collections management lifecycle + `ErrNoManagementKey` paths, `internal/conn` dial interceptor and metadata wiring. Package coverage: collections 57→84%, image 53→73%, conn 54→96%.
- CI collects a coverage profile (generated `xai/api/v1` excluded from the total) and uploads it as an artifact; `make cover` mirrors it locally.

### Dependencies

- Bumped `go.opentelemetry.io/otel{,/sdk,/trace,/metric}` to v1.45.0, `golang.org/x/net` to v0.57.0, `golang.org/x/text` to v0.41.0, `golang.org/x/sys` to v0.47.0 (clears all govulncheck findings in imported packages).

## [0.1.2] — 2026-08-11

### Changed

- Test suite leaner (~84→81 `Test*`, ~3.7k→3.5k hand-written test LOC): root wiring smoke replaces multi-domain RPC mega-test; merge tools/chat/image/video overlaps; domain packages keep sole RPC coverage (files get/delete/public URL, collections CRUD/search, etc.).
- Style pass aligned with project Go constraints: `goimports -local` import groups (stdlib → third-party → module), godoc on exported chat options/response accessors, `Client.Close` joins multi-conn errors via `errors.Join`, safer `proto.Clone` type assert, package comments for `internal/*`.
- `make fmt` prefers `goimports -local github.com/fun7257/xai-sdk-go` (falls back to `gofmt`).

### Docs

- 完备公开 API 参考：`docs/API.md`（各包函数/类型/选项说明与引用）。
- README 英文主文档 + 中文 [`README.zh-CN.md`](README.zh-CN.md)；页内语言切换。
- CONTRIBUTING style section documents import groups, godoc, and error-handling expectations.
- `docs/PROTO.md`: inventory of kept protos, omit reasons for embed/Sample RPC, regen notes for stale `*_grpc.pb.go`.

### Removed

- **Embed proto/generated client surface** (`third_party/.../embed.proto`, `embed*.pb.go`) — product non-goal; unused by library packages.
- **Sample RPC surface** — slimmed `sample.proto` to `FinishReason` only (still required by chat); removed `sample_grpc.pb.go` and unused SampleText messages.

## [0.1.1] — 2026-08-10

### Fixed

- CI: bump golangci-lint to **v2.12.2** (built with Go 1.26) so lint works with `go 1.26` in `go.mod`. Older v2.1.x failed: "Go language version used to build golangci-lint is lower than the targeted Go version".

## [0.1.0] — 2026-08-10

First public release of the Go gRPC client for the xAI API.

### Added

- Domain clients: Auth, Chat, Image, Video, Files, Batch, Collections, Models, Tokenize.
- Go-first chat UX: `Sample` / `Samples`+`WithN`, `StreamReader`, `Defer` / `Defers`, `Parse`, stored/compact helpers.
- Tools/search validate on primary path (`Unchecked*` escape hatches).
- Files: list filter/sort, batch upload options, public URL, `ContentWriter`.
- Image/Video: file_id inputs, storage options, multi-sample shapes, download helpers.
- Collections: management-key CRUD, `UploadDocument`, search options, chunk validation.
- Batch: typed `Result` (Chat/Image/Video accessors).
- Telemetry: opt-in OTEL `Setup`.
- Sentinel errors: `ErrNoAPIKey`, `ErrEmptyAPIKey`, `ErrNoManagementKey`.
- Offline `Example*` tests; bufconn coverage for auth/models/tokenize and core packages.
- Examples under `examples/` (including bilingual `examples/complete`).
- Governance: SECURITY, COMPATIBILITY, CONTRIBUTING, Dependabot, CODEOWNERS, issue/PR templates.
- CI workflows: offline gates + optional integration smoke; floating `dev` channel.
- Docs: `GUIDE.zh-en.md`, `PARITY.md`, `DIFF.md`, `PROTO.md`, `RELEASE.md`, `CI.md`.

### Changed

- **Versioning:** releases are **git tags only** (`vX.Y.Z`). `xai.Version()` reports the module version from the Go module graph (`runtime/debug`), or `devel` for untagged local trees.
- **CI:** offline suite on PR, main, **`v*` tags**, and a **daily schedule**; floating tag **`dev`** force-updated after green main builds. See `docs/CI.md`.

### Proto

- Baseline from [xai-org/xai-proto](https://github.com/xai-org/xai-proto); residual fields in `docs/PROTO.md`.

### Dependencies

- `google.golang.org/grpc` v1.82.1, OpenTelemetry v1.43.0 (govulncheck-clean called paths).
