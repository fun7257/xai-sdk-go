# Changelog

## Unreleased

### Changed

- **Versioning:** releases are **git tags only** (`vX.Y.Z`). `xai.Version()` reports the module version from the Go module graph (`runtime/debug`), or `devel` for untagged local trees.
- **CI:** offline suite on PR, main, **`v*` tags**, and a **daily schedule**; floating tag **`dev`** force-updated after green main builds. See `docs/CI.md`.

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

### Proto

- Baseline from [xai-org/xai-proto](https://github.com/xai-org/xai-proto); residual fields in `docs/PROTO.md`.

### Dependencies

- `google.golang.org/grpc` v1.82.1, OpenTelemetry v1.43.0 (govulncheck-clean called paths).
