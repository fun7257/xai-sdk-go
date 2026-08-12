# Changelog

## Unreleased

### Fixed (low-severity audit batch)

- **chat.Parse**: session `ResponseFormat` restore now uses `defer`, so a panicking Sample cannot leave the session mutated; **chat.Compact** surfaces `Append` errors instead of discarding them.
- **chat.ProcessChunk**: response-level fields (id, model, usage, system_fingerprint, service_tier) are only overwritten by chunks that carry them — a trailing empty chunk no longer erases accumulated values.
- **files.Content**: returns no partial data alongside an error.
- **image.Response**: negative index degrades gracefully instead of panicking.
- **collections.Update**: no longer mutates the caller's request (collection id set on an internal clone).
- **telemetry**: `SetTracer` now toggles `Enabled()` in line with its docs; truthy env parsing accepts `True`; package docs cover `XAI_SDK_INCLUDE_SENSITIVE_TELEMETRY_ATTRIBUTES`; `SetupConsoleRecipeDoc` godoc fixed.
- **conn / xai.WithMetadata**: an odd number of metadata elements now fails `NewClient` fast instead of silently dropping the orphan key; the single-value-per-key limitation is documented.
- **tools**: `Function` validates pre-encoded (string/[]byte/RawMessage) parameters as JSON; `WithMCPHeaders` copies the map; `CollectionsSearch` never returns nil and copies the limit value.
- **search.Parameters.Proto**: copies the Sources slice and MaxSearchResults value so later caller mutations cannot leak into built requests.

### Docs (low-severity audit batch)

- Install pins bumped to `v0.1.2` (README, README.zh-CN, API reference).
- `docs/SECURITY.md`: best-effort response expectations table (acknowledgement ≤7d, triage ≤14d).
- `examples/video` and `examples/deferred` bound their polling with context timeouts.
- `files.StorageOptions.Proto` documents that a non-nil zero value still requests storage with server defaults.

### Fixed

- **files.Upload**: a server-side stream abort (quota/size rejection) now surfaces the real gRPC status instead of a bare `EOF` (`Send` == `io.EOF` → status via `CloseAndRecv`).
- **chat streaming hardening**: `Response.ProcessChunk` and multi-response wrapping drop outputs with negative or absurdly large indexes instead of panicking / allocating unbounded memory on malformed stream data.
- **files.ContentWriter**: cancels the server stream promptly on early returns (e.g. writer errors) instead of holding it until the caller's context ends.
- **files.BatchUploadItems**: bounded worker pool (no longer one goroutine per item); `Path`+`Reader` both set is now an error for that item; panics from an item upload or the `onComplete` callback are recovered into that item's error instead of crashing the process.
- **collections.UploadDocument**: `WithDeleteOnAddFailure` cleanup now runs on a context detached from the caller's (30s bound), so orphan files are still deleted when the add failed due to cancellation/deadline.
- **batch.Result.Error**: preserves structured status details (`status.FromProto`) instead of rebuilding from code+message only.

### Changed

- **Strict enum inputs**: `image.WithResolution`, `video.WithResolution`, and `collections.WithMetric` now cause the call to return an error for unknown values (previously silent fallback), matching aspect-ratio handling. `tools.Mode` / `files.WithListSortBy` document their fallback.
- **search.XSource** now validates included/excluded handle mutual exclusion and returns `(*xaiv1.Source, error)`, matching `WebSource`/`NewsSource`; use `UncheckedXSource` for the previous no-error shape. Added `ValidateXHandles`.
- **tools bool options**: `WithImageUnderstanding(false)` / `WithImageSearch(false)` / `WithXMediaUnderstanding(false, false)` now send an explicit `false` on the wire (previously a no-op); unset options remain absent.
- **files TTL validation**: `Upload` rejects `WithExpiresAfter` values outside `[1h, 30d]` (`MinExpiresAfter` / `MaxExpiresAfter`) before streaming; `CreatePublicURL` rejects sub-second TTLs.
- Docs: cleartext warnings on `xai.WithInsecure` and `tools.WithMCPAuth` / `tools.MCP`.

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
