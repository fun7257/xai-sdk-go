# Contributing

## Prerequisites

- Go **1.26+**
- Optional: `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc` (proto regeneration only)
- Optional: `golangci-lint` v2.x (or run `make lint` which uses `go run` when needed)

## Development

```bash
make test      # go test ./...
make vet
make examples  # go build ./examples/...
make race      # race on concurrent packages
make lint      # golangci-lint
make vuln      # govulncheck
gofmt -w $(find . -name '*.go' -not -path './xai/api/v1/*')
```

These targets mirror CI (`.github/workflows/ci.yaml`).

## Proto generation

See [`docs/PROTO.md`](docs/PROTO.md).

```bash
make proto
make test
```

Do not hand-edit `xai/api/v1/*.pb.go`. Residual synthesized fields are documented in PROTO.md.

## Pull requests

- Clear description of *why* and *what* (no legacy “PR plan ID” required)
- `make test` / `make vet` / `make examples` green
- No secrets or API keys in the diff
- Update `CHANGELOG.md` for user-visible changes
- Update [`docs/DIFF.md`](docs/DIFF.md) / [`docs/PARITY.md`](docs/PARITY.md) if public API or Go UX changes
- Update [`docs/COMPATIBILITY.md`](docs/COMPATIBILITY.md) if stability policy changes

## Style

- Idiomatic Go: context-first, functional options, errors as values
- Prefer the **Go-first** public entry points (`Sample`/`Samples`, validating tool constructors) — see DIFF.md
- Domain packages: thin public API; generated pb via `Proto()` escape hatch
- Avoid panics on recoverable input paths in new code

## Security

See [`docs/SECURITY.md`](docs/SECURITY.md). Never commit `.env`, keys, or tokens.
Examples must read `XAI_API_KEY` / `XAI_MANAGEMENT_KEY` from the environment.

## License

By contributing, you agree that your contributions are licensed under the Apache License 2.0 (see `LICENSE`).
