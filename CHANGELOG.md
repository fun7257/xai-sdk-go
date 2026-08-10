# Changelog

## Unreleased

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
