# Changelog

## Unreleased

### Changed

- **Versioning:** releases are **git tags only** (`vX.Y.Z`). Removed hand-maintained version constant; `xai.Version()` now reports the module version from the Go module graph (`runtime/debug`), or `devel` for untagged local trees. Wire metadata still uses that value.
- **CI:** offline suite runs on PR, push to main/master, **`v*` tags**, and a **daily schedule**; after green main builds, floating tag **`dev`** is force-updated (development channel). Documented in `docs/CI.md`.

### Docs

- **Doc consolidation:** single navigation map in root `README.md`; each topic has one primary home (GUIDE usage, PARITY design, DIFF Python diffs, CONTRIBUTING gates, RELEASE version, COMPATIBILITY stability). Slimmed overlapping tables; `ROADMAP_OSS` collapsed to completed status card.
- Link integrity: `docs/doclinks_test.go` checks in-repo Markdown links and README nav roles.
- Bilingual complete usage guide: `docs/GUIDE.zh-en.md`; annotated example: `examples/complete/`.

### Operations (P2)

- Dependabot: weekly `gomod` + `github-actions` (`.github/dependabot.yml`).
- Collaboration: `CODEOWNERS`, issue templates (bug/feature), PR template.
- Optional live integration smoke: `//go:build integration` under `integration/`; `make integration`; secrets-gated workflow `.github/workflows/integration.yml` (never required on offline PR CI).
- Release checklist: `docs/RELEASE.md`; `xai.Version` documented as single public version source of truth (wire metadata via `init`).

### Expression & correctness (P1)

- Exported sentinel errors: `ErrNoAPIKey`, `ErrEmptyAPIKey`, `ErrNoManagementKey` (usable with `errors.Is`; construction paths wrap with `%w`).
- Runnable package `Example*` for root, `chat`, `tools`, and `image` (offline; no live network).
- Raised test coverage on `batch` (~90%), `tools` (~97%), `chat` (≥55%) via real public-method tests.
- Documented panic contract on `chat.User` / `System` / `Assistant` / `Developer`; dual-path tests with `New*` error constructors.

### Open-source hygiene (senior SDK floor)

- Added `docs/SECURITY.md` (vulnerability reporting + consumer notes).
- Added `docs/COMPATIBILITY.md` (v0 semver policy, public surface, protobuf contract).
- Added `docs/ROADMAP_OSS.md` (P0–P2 remediation plan).
- Expanded CI (`.github/workflows/ci.yaml`): vet, test, scoped race, examples build, golangci-lint, govulncheck.
- Expanded `Makefile`: `test`, `vet`, `examples`, `race`, `lint`, `vuln`, `check`; portable proto includes.
- Real bufconn tests for previously zero-coverage packages: `auth`, `models`, `tokenize`.
- Library quality: `grpc.NewClient` dial path; Close errcheck on library I/O; package docs on domain clients.
- Dependencies: `google.golang.org/grpc` → v1.82.1; OpenTelemetry → v1.43.0 (govulncheck clean on called symbols).
- golangci-lint config for v2; CI action pin aligned.
- CONTRIBUTING/README point at security, compatibility, and quality gates (no deleted plan docs as process).

### API / design docs

- Comprehensive Go vs Python difference guide: `docs/DIFF.md`.
- Go-first public API: `Samples`/`WithN`, `Defers`/`WithDeferN`, StreamReader/Stream + `WithN`; tools/search validate on primary path (`Unchecked*` escape hatches).
- Code polish: request clone for chat `makeRequest`, `ContentWriter`, `WithDeleteOnAddFailure`, `Stream + WithN`, index-scoped `ToolOutputs`.

## [0.2.0] — 2026-08-10

Product-depth release for the Go gRPC client (chat, image, video, files, batch,
collections, models, tokenize, auth, opt-in telemetry).

### Added

- Files: `WithListFilter`, `WithListSortBy`, `BatchUploadWithOptions` + per-file callback.
- Image: `WithImageFileID` / `WithImageFileIDs`, storage options, FileOutput/PublicURL accessors, `Download`.
- Video: file_id inputs, storage wire-up, extend + poll tests.
- Collections: `UploadDocument`, list/search options, retrieval_mode, chunk validation, field helpers.
- Batch: typed `Result` (`Succeeded`/`Failed`, Chat/Image/Video accessors), `GetBatchRequestResult`.
- Chat: `StreamReader.Recv`, response format options, deferred multi-completion shapes.
- Tools: WebSearch/XSearch validation on primary path; CollectionsSearch retrieval_mode.
- Telemetry: opt-in OTEL `Setup`; env disable flags.
- Examples under `examples/`; design docs `docs/PARITY.md`, `docs/DIFF.md`, `docs/PROTO.md`.

### Proto

- Baseline from [xai-org/xai-proto](https://github.com/xai-org/xai-proto); residual fields documented in `docs/PROTO.md`.

## [0.1.0] — preview

- Initial gRPC SDK surface: Client, Chat, Image, Video, Files, Batch, Collections, Models, Tokenize, Auth.
