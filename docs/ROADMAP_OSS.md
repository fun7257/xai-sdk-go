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

### P1 — Expression & correctness (**done**)

| # | Task | Acceptance |
|---|------|------------|
| P1.1 | Sentinel errors | `ErrNoAPIKey`, `ErrEmptyAPIKey`, `ErrNoManagementKey` with `%w` / `errors.Is` |
| P1.2 | Package godoc + `Example*` | Root + `chat` + `tools` + `image` offline `Example*` |
| P1.3 | Coverage floor | `chat` / `batch` / `tools` each ≥55% statement coverage |
| P1.4 | Panic factories | Documented panic contract; `New*` error path + panic tests |
| P1.5 | CHANGELOG hygiene | Unreleased records P1; no deleted plan docs as active process |

### P2 — Operations (**done**)

| # | Task | Acceptance |
|---|------|------------|
| P2.1 | Dependabot | Weekly `gomod` + `github-actions` in `.github/dependabot.yml` |
| P2.2 | CODEOWNERS + issue/PR templates | `.github/CODEOWNERS`, ISSUE_TEMPLATE, pull_request_template |
| P2.3 | Integration smoke | `//go:build integration`, `make integration`, secrets-gated workflow |
| P2.4 | Release path | `docs/RELEASE.md`; `xai.Version` single public version source |

## Wave status

| Wave | Status |
|------|--------|
| P0 / Wave 1 | **Done** — governance docs, CI/Makefile gates, zero-pkg tests, lint+vuln clean, local git |
| P1 / Wave 2 | **Done** — sentinels, Example*, coverage floors, panic factory docs |
| P2 / Wave 3 | **Done** — Dependabot, collab templates, integration smoke, release checklist |

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
