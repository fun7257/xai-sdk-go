# Release checklist

Maintainer steps to ship a version of `github.com/fun7257/xai-sdk-go`.

## Version single source of truth

| Symbol | Role |
|--------|------|
| **`xai.Version`** in `version.go` | **Only public version constant.** Bump this for releases. |
| `conn.SDKVersion` | Wire metadata (`xai-sdk-version: go/<ver>`). Set from `xai.Version` in root package `init()`. Default string in `internal/conn` must match `version.go` until init runs. |

Do **not** invent a second public version API. Tests assert `conn.SDKVersion == xai.Version` after import.

## Before tagging

1. Update **`version.go`** (`xai.Version = "X.Y.Z"`).
2. Keep **`internal/conn` default `SDKVersion`** string in sync with that value (same `X.Y.Z`).
3. Move **`CHANGELOG.md`** Unreleased notes into `## [X.Y.Z] — YYYY-MM-DD`.
4. Run offline gates:

   ```bash
   make check
   make race
   make lint
   make vuln
   ```

5. Optional live smoke (costs API usage):

   ```bash
   export XAI_API_KEY=xai-...
   make integration
   # or: go run ./examples/smoke
   ```

6. Commit on `main` with a clear message.

## Tag and publish

```bash
git tag vX.Y.Z
git push origin main
git push origin vX.Y.Z
```

Consumers:

```bash
go get github.com/fun7257/xai-sdk-go@vX.Y.Z
```

pkg.go.dev indexes the module after the tag is public.

## Post-release

- Confirm GitHub Actions CI is green on the tag/commit.
- Open Dependabot PRs as needed (weekly schedule in `.github/dependabot.yml`).
- For security fixes, follow `docs/SECURITY.md`.

## Not required for a library module

- Multi-platform binary goreleaser artifacts (this is an importable Go module, not a CLI ship).
- Mandatory live e2e on every PR (integration is opt-in / secrets-gated).
