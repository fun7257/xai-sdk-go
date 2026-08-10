# Release checklist

This module follows **standard Go module releases**: the published version is the
**git tag**. There is no hand-maintained `version.go` constant to bump.

## Version source of truth

| Source | Role |
|--------|------|
| **Git tag** `vX.Y.Z` | Module version consumers get via `go get …@vX.Y.Z` / proxy |
| **`xai.Version()`** | Runtime report of that module version (from `runtime/debug` build info) for logs and gRPC metadata `xai-sdk-version: go/<ver>` |
| Local untagged tree | `Version()` returns `devel` |

Do **not** reintroduce a duplicate in-repo version number.  
Semver policy: [`COMPATIBILITY.md`](COMPATIBILITY.md).

## Before tagging

1. Fold **`CHANGELOG.md`** Unreleased notes into `## [X.Y.Z] — YYYY-MM-DD`.
2. Offline gates ([`CONTRIBUTING.md`](../CONTRIBUTING.md)):

   ```bash
   make check && make race && make lint && make vuln
   ```

3. Optional live smoke (API cost):

   ```bash
   export XAI_API_KEY=xai-...
   make integration   # or: go run ./examples/smoke
   ```

4. Commit on `main` (clean tree at the commit you will tag).

## Tag and publish

Tags must be annotated or lightweight **semver with a leading `v`** (Go module convention):

```bash
git tag vX.Y.Z
git push origin main
git push origin vX.Y.Z
```

Consumers:

```bash
go get github.com/fun7257/xai-sdk-go@vX.Y.Z
```

After the tag is public, `xai.Version()` in dependent binaries reports `X.Y.Z`
(leading `v` stripped for display/wire).

## Post-release

- Confirm default CI green on the tagged commit  
- Dependabot: `.github/dependabot.yml`  
- Security: [`SECURITY.md`](SECURITY.md)

## Not required

- Hand-edited version constants or dual version files  
- Goreleaser multi-platform binaries (importable library module)  
- Mandatory live e2e on every PR  
