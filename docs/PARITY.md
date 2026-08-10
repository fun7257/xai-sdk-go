# Product capability & Go API design

This SDK exposes the **xAI gRPC product surface** with **idiomatic Go call UX**.  
It is **not** a Python API clone: keep service capability, choose Go-native shapes when clearer or safer.

| Doc | Role |
|-----|------|
| **This file** | Principles + preferred entry points (short) |
| [`DIFF.md`](DIFF.md) | Domain-by-domain Go vs Python differences (truth source) |
| [`GUIDE.zh-en.md`](GUIDE.zh-en.md) | Full parameter reference (EN + 中文) |
| [`PROTO.md`](PROTO.md) | Residual proto fields |

## Design principles

1. **Call-site first** — common paths short and hard to misuse.
2. **No silent data loss** — multi-result RPCs never drop extra indices; singular APIs reject `n>1`.
3. **Validate on the primary path** — illegal option combos fail at tool/source construction.
4. **One primary name family** — not Python dual stacks or deprecated `*_batch` names.
5. **Opt-in global side effects** — OTEL off until `telemetry.Setup`.

## Preferred entry points

| Domain | Preferred Go API |
|--------|------------------|
| Chat one-shot | `Sample(ctx)` → `*Response` |
| Chat multi | `Samples(ctx, chat.WithN(k))` → `[]*Response` |
| Chat stream | `StreamReader(ctx, chat.WithN(k)?)` + `Recv` / `Close` |
| Chat deferred | `Defer(ctx)` or `Defers(ctx, chat.WithDeferN(k))` |
| Chat structured | `Parse(ctx, schemaJSON, &dest)` |
| Image one-shot / multi | `Sample(...)` / `Samples(..., image.WithN(k))` |
| Tools | `tools.WebSearch` / `XSearch` → `(*Tool, error)`; escape: `Unchecked*` |
| Search sources | `search.WebSource` / `NewsSource` → `(*Source, error)` |
| Files large download | `ContentWriter(ctx, id, w)` |
| Video extend async | `ExtendStart` then `Get` |

## Out of scope

Embed product client · OpenAI REST · OAuth · aio dual clients · auto `.env` · cloning Python names / Pydantic.

## Next

- Usage & options → [`GUIDE.zh-en.md`](GUIDE.zh-en.md) · [`examples/complete/`](../examples/complete/)
- Stability → [`COMPATIBILITY.md`](COMPATIBILITY.md)
- Contribute → [`../CONTRIBUTING.md`](../CONTRIBUTING.md)
