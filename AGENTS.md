# AGENTS.md

## Cursor Cloud specific instructions

This repo is the **xAI Go SDK** (`github.com/fun7257/xai-sdk-go`) — a single Go module
that is an idiomatic gRPC client for the remote xAI API (Grok). It is a **library**, not a
deployable service: there are no local servers, databases, or web frontends. The "backend"
is xAI's hosted gRPC endpoints (`api.x.ai:443`, `management-api.x.ai:443`).

### Toolchain (non-obvious)
- Requires **Go 1.26+** (`go.mod` declares `go 1.26`). The base VM historically ships an
  older Go (e.g. 1.22), and in this environment `GOTOOLCHAIN=auto` **cannot auto-download**
  the 1.26 toolchain (fails with `toolchain not available`). Go 1.26 is therefore installed
  directly under `/usr/local/go` (symlinked into `/usr/local/bin/go`). Verify with
  `go version` (should report `go1.26.x`).

### Commands (offline, no API key — this is what CI enforces)
Standard commands live in the `Makefile` and `CONTRIBUTING.md`. Summary:
- `make check` — `go vet ./...` + `go test ./...` + build examples (primary gate).
- `make race` — race detector on concurrent packages (`files collections chat internal/conn batch`).
- `make lint` — `golangci-lint`. **Not installed as a binary**; the target falls back to
  `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2`, so the first run
  downloads and compiles the linter (slow, needs network).
- `make vuln` — `govulncheck`, likewise `go run` on first use.

### Live end-to-end (optional, requires a real key)
- Business calls (Chat, Image, Video, Files, Batch, Models, Tokenize, Auth, Search) need
  `XAI_API_KEY`. Collections CRUD additionally needs `XAI_MANAGEMENT_KEY`.
- Without a key: `make integration` (build tag `integration`) **skips**, and
  `go run ./examples/smoke` exits with code 2. With a valid key, run
  `XAI_API_KEY=xai-... go run ./examples/smoke` for a low-cost live smoke, or
  `XAI_API_KEY=xai-... make integration`.
- Set `XAI_SDK_DISABLE_TRACING=1` to disable OpenTelemetry tracing.
- Regenerating protobufs (`make proto`) needs `protoc` + plugins and is only relevant when
  editing `third_party/**/*.proto`; generated code in `xai/api/v1/` is committed.
