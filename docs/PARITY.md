# Product capability & Go API design

This SDK exposes the **xAI gRPC product surface** with **idiomatic Go call UX**.
It is **not** a Python API clone: we keep capability parity with the service, and
choose Go-native shapes when they are clearer or safer.

**Full difference guide (domain-by-domain):** [`DIFF.md`](DIFF.md)

## Design principles

1. **Call-site first** — common paths should be short and hard to misuse.
2. **No silent data loss** — multi-result RPCs never drop extra indices; singular APIs reject n>1.
3. **Validate on the primary path** — illegal option combos fail where users construct tools/sources.
4. **One primary name family** — not Python dual stacks or deprecated `*_batch` names.
5. **Opt-in global side effects** — OTEL is off until `telemetry.Setup`.

## Preferred entry points (summary)

| Domain | Preferred Go API |
|--------|------------------|
| Chat one-shot | `Sample(ctx)` → `*Response` |
| Chat multi | `Samples(ctx, chat.WithN(k))` → `[]*Response` |
| Chat stream | `StreamReader(ctx, chat.WithN(k)?)` + `Recv` / `Close` |
| Chat deferred | `Defer(ctx)` or `Defers(ctx, chat.WithDeferN(k))` |
| Chat structured | `Parse(ctx, schemaJSON, &dest)` |
| Image one-shot | `Sample(...)` → `*Response` |
| Image multi | `Samples(..., image.WithN(k))` → `[]*Response` |
| Tools | `tools.WebSearch` / `XSearch` → `(*Tool, error)` (validated) |
| Tools escape | `UncheckedWebSearch` / `UncheckedXSearch` |
| Search sources | `search.WebSource` / `NewsSource` → `(*Source, error)` |
| Files large download | `ContentWriter(ctx, id, w)` |
| Video extend async | `ExtendStart` then `Get` |

## Out of scope

- Embed product client, OpenAI REST, OAuth, aio dual clients, auto `.env`
- Cloning Python names / Pydantic / ProtoDecorator

## Proto residual

See [`PROTO.md`](PROTO.md). Functional wire parity may use local residual fields until public `xai-org/xai-proto` catches up.

## Verify

```bash
go test ./...
go vet ./...
go build ./examples/...
```
