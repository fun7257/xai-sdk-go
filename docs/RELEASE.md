# Release checklist

Maintainer steps to ship `github.com/fun7257/xai-sdk-go`.

## Version single source of truth

| Symbol | Role |
|--------|------|
| **`xai.Version`** in `version.go` | **Only public version constant** — bump for releases |
| `conn.SDKVersion` | Wire metadata `xai-sdk-version: go/<ver>`; set from `xai.Version` in root `init()`. Keep the default string in `internal/conn` equal to `version.go` until init runs |

Do not invent a second public version API. `TestVersionSingleSourceOfTruth` asserts equality after import.  
Semver / public surface policy: [`COMPATIBILITY.md`](COMPATIBILITY.md).

## Before tagging

1. Set **`version.go`** (`xai.Version = "X.Y.Z"`).
2. Sync **`internal/conn` default `SDKVersion`** to the same string.
3. Fold **`CHANGELOG.md`** Unreleased into `## [X.Y.Z] — YYYY-MM-DD`.
4. Offline gates (full list: [`CONTRIBUTING.md`](../CONTRIBUTING.md)):

   ```bash
   make check && make race && make lint && make vuln
   ```

5. Optional live smoke (API cost):

   ```bash
   export XAI_API_KEY=xai-...
   make integration   # or: go run ./examples/smoke
   ```

6. Commit on `main`.

## Tag and publish

```bash
git tag vX.Y.Z
git push origin main
git push origin vX.Y.Z
```

```bash
go get github.com/fun7257/xai-sdk-go@vX.Y.Z
```

## Post-release

- Confirm default CI green on the tag  
- Dependabot: `.github/dependabot.yml` (weekly)  
- Security process: [`SECURITY.md`](SECURITY.md)

## Not required

- Goreleaser multi-platform binaries (this is an importable module)  
- Mandatory live e2e on every PR  
