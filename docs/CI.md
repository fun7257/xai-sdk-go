# Continuous integration

How GitHub Actions run for this **library module** (no binary release artifacts).

## Workflows

| Workflow | File | Purpose |
|----------|------|---------|
| **ci** | [`.github/workflows/ci.yaml`](../.github/workflows/ci.yaml) | Offline gates: vet, test + coverage profile (uploaded artifact), race (scoped), examples, lint, govulncheck |
| **integration** | [`.github/workflows/integration.yml`](../.github/workflows/integration.yml) | Optional live API smoke (`//go:build integration`) |
| **release** | [`.github/workflows/release.yml`](../.github/workflows/release.yml) | On `v*` tags: publish GitHub Release notes from the matching `CHANGELOG.md` section; manual dispatch backfills existing tags |

Local equivalents: [`CONTRIBUTING.md`](../CONTRIBUTING.md) (`make check`, `race`, `lint`, `vuln`, `integration`).

---

## `ci` — when it runs

| Trigger | Ref | Notes |
|---------|-----|--------|
| **pull_request** | PR head | Every PR; blocks merge when required in branch protection |
| **push** `main` / `master` | Branch tip (**latest** code) | Same suite as PR |
| **push** tags `v*` | Release tag (e.g. `v0.1.0`) | Suite runs on the tagged commit |
| **schedule** | Default branch | **Daily 02:00 UTC** — same offline test suite (regression) |
| **workflow_dispatch** | Chosen ref | Manual re-run in Actions UI |

### Jobs

1. **offline gates** (`test`) — always for the triggers above.  
2. **update dev tag** (`publish-dev`) — **only** after a **green** `push` to `main` or `master`.

`publish-dev` does **not** run on PRs, tags, or the schedule (schedule uses the default branch checkout but is not a “new tip publish” event in our condition; only branch push updates `dev`).

---

## Development version tag `dev` (overwrite)

After every successful CI on **main/master**, Actions **force-moves** a floating git tag:

```text
dev  →  current main/master SHA   (replaced every green push)
```

| Channel | How consumers get it | Stability |
|---------|----------------------|-----------|
| **Release** | `go get github.com/fun7257/xai-sdk-go@vX.Y.Z` | Immutable semver tag |
| **Development** | `go get github.com/fun7257/xai-sdk-go@dev` | **Mutable** — points at latest green main |

Notes:

- `dev` is **not** a semver release; do not use it in production pins.  
- Module proxy may cache floating tags; for the absolute tip use  
  `GOPROXY=direct go get …@dev` or pin a commit SHA.  
- Requires the workflow `contents: write` permission on that job (granted only there).

---

## `integration` — live smoke (optional)

| Trigger | When |
|---------|------|
| **workflow_dispatch** | Manual |
| **schedule** | Weekly Monday 06:00 UTC |

- Needs repository secret **`XAI_API_KEY`**.  
- Missing secret → job **succeeds** with skip (does not fail the workflow).  
- **Not** run on pull_request (offline CI stays free of secrets).

---

## What CI does *not* do

- Build or attach goreleaser / binary “packages”  
- Publish to a package registry (Go modules use the **git host + tags**)  
- Require live xAI access for PR merge  

Release process: [`RELEASE.md`](RELEASE.md).
