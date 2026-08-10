# API compatibility and stability

## Semantic versioning

| Channel | Stability |
|---------|-----------|
| Module path | `github.com/fun7257/xai-sdk-go` |
| Current series | **v0.x** (`xai.Version` in `version.go`, currently `0.2.0`) |
| Pre-v1 policy | **Breaking changes are allowed** on minor bumps within v0 when needed for correctness or Go UX. Prefer deprecations when practical. |
| v1.0.0 (future) | Intended to mark a stability floor for the preferred Go entry points documented in [`PARITY.md`](PARITY.md) and [`DIFF.md`](DIFF.md). |

Always pin a module version (or commit) in applications. Read `CHANGELOG.md` when upgrading.

## What we treat as the public surface

**Supported for application use:**

- Root package `xai`: `NewClient`, `Option`s, `Client` domain fields, `Close`, `Version`
- Sentinel errors: `ErrNoAPIKey`, `ErrEmptyAPIKey`, `ErrNoManagementKey` (`errors.Is`)
- Domain packages: `chat`, `image`, `video`, `files`, `batch`, `collections`, `tools`, `search`, `telemetry`, `types`, `auth`, `models`, `tokenize`
- Preferred call shapes (e.g. `Sample` / `Samples` + `WithN`, validating `tools.WebSearch`) described in docs

**Escape hatches (supported but may track server/proto evolution more closely):**

- `Proto()` accessors returning generated messages
- Constructing or passing `xai/api/v1` (`xaiv1`) types in public function signatures where already exposed

## Generated protobuf package (`xai/api/v1`)

| Topic | Policy |
|-------|--------|
| Source | Generated from `third_party` protos; do not hand-edit `*.pb.go` |
| Public contract | **Part of the module’s importable surface** — callers may use `xaiv1` types today |
| Stability | Field sets can change when residual patches or upstream protos update; **not** a frozen ABI within v0 |
| Residual fields | Documented in [`PROTO.md`](PROTO.md); required for product wire parity with the live API / Python SDK descriptors |

If you need a narrower stable façade without raw protos, wrap this SDK in your own types; a full erase-pb redesign is out of scope for v0 hygiene work.

## Deprecated names

Aliases such as `WebSearchChecked` / `XSearchChecked` may remain for a short period and are marked deprecated in godoc. Prefer the primary validating constructors.

## CI and local quality bar

Contributors and CI should keep green:

```bash
make test
make vet
make examples
make lint    # golangci-lint when installed
make vuln    # govulncheck
make race    # scoped race on concurrent packages
```

See `Makefile` and `.github/workflows/ci.yaml`.

## Version and releases

- Public version constant: **`xai.Version`** only (`version.go`).
- Wire metadata uses the same value via package `init` → `conn.SDKVersion`.
- Maintainer release steps: [`RELEASE.md`](RELEASE.md).
