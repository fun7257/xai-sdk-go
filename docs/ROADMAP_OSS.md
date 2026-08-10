# OSS maturity status

Remediation waves **P0–P2 are complete**. This file is a **status card**, not an active work tracker.

| Wave | Focus | Status |
|------|-------|--------|
| **P0** | Governance docs, CI/Makefile, zero-pkg tests, lint/vuln, local git | **Done** |
| **P1** | Sentinel errors, `Example*`, coverage floors, panic factory docs | **Done** |
| **P2** | Dependabot, CODEOWNERS/templates, integration smoke, release checklist | **Done** |

## What “done” means

- **Consumers:** install, use preferred APIs, pin `xai.Version` / tags — see root [`README.md`](../README.md) and [`GUIDE.zh-en.md`](GUIDE.zh-en.md).
- **Contributors:** offline gates in [`CONTRIBUTING.md`](../CONTRIBUTING.md).
- **Maintainers:** ship via [`RELEASE.md`](RELEASE.md).

## Explicit non-goals (still)

- Full public API without `xaiv1` in one pass  
- Embed / OpenAI REST / OAuth clients  
- Live e2e as a required PR CI gate  
- 90%+ coverage on every package  

Historical task tables from the remediation effort lived only in this roadmap and are collapsed above. Do not treat deleted `PR_PLAN` / `MISSING` docs as process.
