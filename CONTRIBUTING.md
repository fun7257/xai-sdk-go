# Contributing

## Prerequisites

- Go **1.26+**
- Optional: `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc` (proto regen only)
- Optional: `golangci-lint` v2.x (`make lint` can `go run` a pin)

## Offline quality gates (primary)

These match default CI (`.github/workflows/ci.yaml`):

```bash
make check   # vet + test + examples build
make race    # scoped race: files, collections, chat, conn, batch
make lint    # golangci-lint
make vuln    # govulncheck
gofmt -w $(find . -name '*.go' -not -path './xai/api/v1/*')
```

## Optional live integration

**Not** part of `make test` or PR CI. Build tag `integration`; skips without key:

```bash
export XAI_API_KEY=xai-...
make integration
# or: go test -tags=integration ./integration/ -count=1 -v
```

Secrets-gated schedule/dispatch: `.github/workflows/integration.yml`.  
Maintainer context: [`docs/RELEASE.md`](docs/RELEASE.md).

## Proto generation

Primary doc: [`docs/PROTO.md`](docs/PROTO.md).

```bash
make proto
make test
```

Do not hand-edit `xai/api/v1/*.pb.go`.

## Pull requests

- Describe *why* and *what*
- Offline gates green (`make check` minimum)
- No secrets or API keys in the diff
- Update `CHANGELOG.md` for user-visible changes
- Public API / Go UX → [`docs/DIFF.md`](docs/DIFF.md) and/or [`docs/PARITY.md`](docs/PARITY.md)
- Stability policy → [`docs/COMPATIBILITY.md`](docs/COMPATIBILITY.md)
- Prefer Go-first entry points (`Sample`/`Samples`, validating tools) — see PARITY

PR template: [`.github/pull_request_template.md`](.github/pull_request_template.md).

## Style

- Idiomatic Go: `context` first, functional options, errors as values
- Domain packages: thin public API; `Proto()` escape hatch for pb
- Avoid panics on recoverable input in new code (`NewUser` vs `User` — see godoc / GUIDE)

## Security & release

- Security: [`docs/SECURITY.md`](docs/SECURITY.md) — never commit `.env` or keys  
- Release / `xai.Version`: [`docs/RELEASE.md`](docs/RELEASE.md)  
- Usage reference: [`docs/GUIDE.zh-en.md`](docs/GUIDE.zh-en.md)

## License

Contributions are under Apache License 2.0 (`LICENSE`).
