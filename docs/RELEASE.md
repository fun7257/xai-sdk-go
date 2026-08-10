# Release checklist

This module follows **standard Go module releases**: published versions are
**immutable git tags** `vX.Y.Z`. There is no hand-maintained version constant.

CI behaviour (PR / tag / schedule / floating `dev`): see **[`CI.md`](CI.md)**.

## Version channels

| Channel | Git ref | Consumer install | Mutability |
|---------|---------|------------------|------------|
| **Release** | tag `vX.Y.Z` | `go get …@vX.Y.Z` | Immutable |
| **Development** | floating tag `dev` | `go get …@dev` | **Overwritten** on every green `main` CI |
| Local tree | no tag | `replace` / work tree | `xai.Version()` often `devel` |

Runtime: `xai.Version()` reports the module version from the Go module graph
(tag when resolved; `devel` for untagged local builds).  
Semver policy: [`COMPATIBILITY.md`](COMPATIBILITY.md).

## Before a release tag

1. Fold **`CHANGELOG.md`** Unreleased into `## [X.Y.Z] — YYYY-MM-DD`.
2. Offline gates ([`CONTRIBUTING.md`](../CONTRIBUTING.md)):

   ```bash
   make check && make race && make lint && make vuln
   ```

3. Optional live smoke:

   ```bash
   export XAI_API_KEY=xai-...
   make integration
   ```

4. Commit on `main` (clean tree at the commit you will tag).

## Tag and publish (release)

```bash
git tag vX.Y.Z
git push origin main
git push origin vX.Y.Z
```

Pushing `v*` runs the **ci** workflow on that tag (offline suite).  
Consumers:

```bash
go get github.com/fun7257/xai-sdk-go@vX.Y.Z
```

## Development tip (`dev`)

No manual step: after CI is green on `main`/`master`, Actions force-updates tag **`dev`**.

```bash
go get github.com/fun7257/xai-sdk-go@dev
# if proxy is stale:
GOPROXY=direct go get github.com/fun7257/xai-sdk-go@dev
```

Do **not** treat `dev` as a production pin.

## Post-release

- Confirm **ci** is green for the `v*` tag and for `main`  
- Dependabot: `.github/dependabot.yml`  
- Security: [`SECURITY.md`](SECURITY.md)

## Not required

- Hand-edited version constants  
- Binary / container release artifacts  
- Live e2e on every PR  
