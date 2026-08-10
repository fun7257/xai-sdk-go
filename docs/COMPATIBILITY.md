# API compatibility and stability

## Semantic versioning

| Channel | Policy |
|---------|--------|
| Module | `github.com/fun7257/xai-sdk-go` |
| Series | **v0.x** — release tags `vX.Y.Z` (immutable); optional floating `dev` for tip of main — [`RELEASE.md`](RELEASE.md) · [`CI.md`](CI.md) |
| Pre-v1 | **Breaking changes allowed** on minor tags when needed for correctness or Go UX; prefer deprecations when practical |
| Future v1 | Intended stability floor for preferred entry points in [`PARITY.md`](PARITY.md) |

Always pin a module version (or commit) in `go.mod`. Read [`CHANGELOG.md`](../CHANGELOG.md) when upgrading.  
Runtime report: `xai.Version()` (from the module graph / tag, not a hand-maintained constant).

## Public surface

**Supported for applications:**

- Root `xai`: `NewClient`, `Option`s, `Client` domains, `Close`, `Version()` (module/tag version)
- Sentinels: `ErrNoAPIKey`, `ErrEmptyAPIKey`, `ErrNoManagementKey` (`errors.Is`)
- Domains: `chat`, `image`, `video`, `files`, `batch`, `collections`, `tools`, `search`, `telemetry`, `types`, `auth`, `models`, `tokenize`
- Preferred call shapes — [`PARITY.md`](PARITY.md); parameters — [`GUIDE.zh-en.md`](GUIDE.zh-en.md)

**Escape hatches** (supported; track server/proto more closely):

- `Proto()` accessors
- Passing `xai/api/v1` (`xaiv1`) types in signatures already exposed

## Generated protobuf (`xai/api/v1`)

| Topic | Policy |
|-------|--------|
| Source | From `third_party` protos; **do not hand-edit** `*.pb.go` |
| Importable | Yes — part of the module surface today |
| Stability | Field sets may change with residual patches / upstream; **not** a frozen ABI in v0 |
| Residuals | [`PROTO.md`](PROTO.md) |

For a narrower façade without raw protos, wrap this SDK in your own types.

## Deprecated names

Aliases such as `WebSearchChecked` / `XSearchChecked` may linger with godoc deprecation. Prefer validating constructors (`WebSearch`, `XSearch`).

## Quality bar

Offline gates for contributors and CI: see **[`CONTRIBUTING.md`](../CONTRIBUTING.md)** (`make check`, lint, vuln, race). Not repeated here.
