# Remediation roadmap: senior-grade open-source SDK

Target: a public Go gRPC client that meets **L2 (collaborative OSS)** now and
hardens toward **L3 (senior expression)** without another Python-alignment pass.

## Written 整改计划 (P0 → P2)

### P0 — Open-source floor (this effort; must ship)

| # | Task | Acceptance | Audit line |
|---|------|------------|------------|
| P0.1 | Governance docs | `docs/SECURITY.md`, `docs/COMPATIBILITY.md` non-empty; README + CONTRIBUTING link them; no MISSING/PR_PLAN as primary process | Ops / Security |
| P0.2 | CI + Makefile gates | Workflow + `Makefile`: `vet`, `test`, `examples`, scoped `race`, `lint` (golangci), `vuln` (govulncheck) | Hygiene |
| P0.3 | Zero-pkg tests | `auth`, `models`, `tokenize` have bufconn/mock tests hitting real public methods | Correctness |
| P0.4 | Lint/library fixes | `grpc.NewClient`; library Close errcheck; primary lint path clean | Expression |
| P0.5 | Local VCS readiness | `git init` + initial commit when `.git` absent; remote/tag left to maintainer | Ops |
| P0.6 | Verification captures | Plan verification steps run; green suite; evidence under implementer scratch | Ship integrity |

**Owner:** implementer (this pass). **Exit:** all acceptance criteria in the goal plan met.

### P1 — Expression & correctness (follow-up wave)

| # | Task | Acceptance |
|---|------|------------|
| P1.1 | Sentinel errors | `ErrNoAPIKey`, management-key, and key domain errors with `%w` / `errors.Is` |
| P1.2 | Package godoc + `Example*` | Root + `chat` (and preferably tools/image) have runnable examples |
| P1.3 | Coverage floor | Cover report excluding examples/generated; raise floors on `chat`, `batch`, `tools` |
| P1.4 | Panic factories | Document `User`/`System` panic contract; keep `NewUser` as error path |
| P1.5 | CHANGELOG hygiene | Historical entries do not present deleted plan docs as current process |

### P2 — Operations (optional hardening)

| # | Task | Acceptance |
|---|------|------------|
| P2.1 | Dependabot / Renovate | Module + Actions updates |
| P2.2 | CODEOWNERS + issue/PR templates | GitHub collaboration surface |
| P2.3 | Integration smoke | `//go:build integration` optional job with secrets |
| P2.4 | Release automation | Tag checklist or goreleaser; `xai.Version` single source of truth |

## Wave status

| Wave | Status |
|------|--------|
| P0 / Wave 1 | **Done** — governance docs, CI/Makefile gates, zero-pkg tests, lint+vuln clean, local git |
| P1 / Wave 2 | Planned (sentinel errors, Example godoc, coverage floors) |
| P2 / Wave 3 | Planned (Dependabot, templates, integration smoke, release automation) |

## Non-goals

- Dropping residual proto product fields
- Full public API without `xaiv1` in one pass
- Embed / OpenAI REST / OAuth clients
- Live network e2e as a required CI gate
- Full 90%+ coverage on every package

## Local gates (match CI)

```bash
make check   # vet + test + examples
make race
make lint
make vuln
```
