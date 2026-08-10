# Security Policy

## Supported versions

| Version | Supported |
|---------|-----------|
| `v0.x` (module path `github.com/fun7257/xai-sdk-go`) | Yes — best-effort fixes on the default branch |
| Pre-release / untagged trees | Yes — treat as development; pin a commit or tag in production |

This library is currently in **v0** (semantic versioning). Breaking changes may occur before `v1.0.0`.

## Reporting a vulnerability

Please **do not** open a public GitHub issue for security-sensitive reports.

Prefer one of:

1. **GitHub Security Advisories** (private report) on the repository, if enabled.
2. Contact the maintainer via the GitHub profile associated with the module path (`fun7257`), with subject line `[SECURITY] xai-sdk-go`.

Include:

- Affected module version or commit SHA
- Description of the issue and impact
- Minimal reproduction (without real API keys)
- Whether the issue is already public

You should receive an acknowledgement when the report is seen. Coordinated disclosure timelines are handled case by case for a v0 project.

## Security notes for consumers

- **API keys**: Never commit keys. Use environment variables (`XAI_API_KEY`, `XAI_MANAGEMENT_KEY`) or explicit options. The SDK does not log credentials.
- **Metadata**: Do not put untrusted header names into `WithMetadata` that are reserved for auth (the SDK forces `Authorization` last).
- **Downloads**: `image.Response.Download` only allows `http`/`https`, caps body size, and should only be used on API-returned URLs (see godoc).
- **Telemetry**: Tracing is opt-in (`telemetry.Setup`). Prompt content attributes are not attached unless you explicitly enable sensitive attributes.
- **Dependencies**: Run `make vuln` / `govulncheck ./...` regularly; CI also runs vulnerability checks when configured.
