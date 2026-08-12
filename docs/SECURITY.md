# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| `v0.x` (`github.com/fun7257/xai-sdk-go`) | Yes — best-effort on default branch |
| Untagged / pre-release trees | Development only; pin commit or tag in production |

This library is **v0**; breaking changes may occur before `v1.0.0`. Stability detail: [`COMPATIBILITY.md`](COMPATIBILITY.md).

## Reporting a vulnerability

**Do not** open a public GitHub issue for security-sensitive reports.

1. **GitHub Security Advisories** (private) on the repository, if enabled.  
2. Contact the maintainer on the GitHub profile for module path `fun7257`, subject `[SECURITY] xai-sdk-go`.

Include: module version or commit SHA · impact · minimal repro **without** real API keys · whether already public.

## Response expectations

Best-effort for a single-maintainer v0 project:

| Stage | Target |
|-------|--------|
| Acknowledgement | Within **7 days** of a private report |
| Triage / severity assessment | Within **14 days** |
| Fix or mitigation for confirmed issues | Next release; critical issues prioritized |
| Coordinated disclosure | Agreed with the reporter case-by-case; default after a fix ships |

No bug bounty is offered.

## Notes for consumers

| Topic | Guidance |
|-------|----------|
| API keys | Env (`XAI_API_KEY`, `XAI_MANAGEMENT_KEY`) or options; never commit; SDK does not log credentials |
| Metadata | Do not put reserved auth headers in `WithMetadata` (SDK forces `Authorization` last) |
| Downloads | `image.Response.Download`: http(s) only, size-capped; use API-returned URLs |
| Telemetry | Opt-in `telemetry.Setup`; sensitive prompt attributes only if explicitly enabled |
| Dependencies | `make vuln` / CI govulncheck — see [`CONTRIBUTING.md`](../CONTRIBUTING.md) |
