# Proto sources

## Source of truth

Official protos from [xai-org/xai-proto](https://github.com/xai-org/xai-proto) live under `third_party/xai/api/v1/`.  
Generated Go: `xai/api/v1/` with:

```text
option go_package = "github.com/fun7257/xai-sdk-go/xai/api/v1;xaiv1";
```

**Last synced upstream commit:** `af1be87b733dc177c0857fbd624f1ff12128fbd2` (see CHANGELOG).

Import stability of `xaiv1` types: [`COMPATIBILITY.md`](COMPATIBILITY.md).

## Tree inventory (kept inputs)

| Proto | Why kept |
|-------|----------|
| `auth.proto` | Auth client |
| `batch.proto` | Batch client (+ `google/rpc/status.proto`) |
| `chat.proto` | Chat client |
| `collections.proto` | Collections management |
| `deferred.proto` | Shared deferred status (chat/video) |
| `documents.proto` | Documents.Search (collections RAG) |
| `files.proto` | Files client |
| `image.proto` | Image client |
| `models.proto` | Models client (incl. embedding *model list*, not Embed RPC) |
| `sample.proto` | **Slim residual:** only `FinishReason` (imported by chat). No Sample RPC / no `sample_grpc.pb.go` |
| `shared_extra.proto` | Imported by collections |
| `tokenize.proto` | Tokenize client |
| `types.proto` | Chunk config types for collections |
| `usage.proto` | Shared usage messages |
| `video.proto` | Video client |
| `google/rpc/status.proto` | Batch result error status |

### Intentionally omitted

| Item | Reason |
|------|--------|
| **`embed.proto` / Embedder service** | Product **non-goal** (no Embed client). Removed from `third_party` and generated tree. |
| **Sample RPC** (`SampleText` / streaming) | Unused; product completions use Chat. Types stripped except `FinishReason`. |

## Residual fields (local synthesis)

Public xai-proto may lag the live API / Python descriptors. This repo keeps **local residual** patches for wire parity:

| Area | Residual |
|------|----------|
| Collections | `collections.proto` / related management messages |
| Files | Public URL RPCs; `ListFilesRequest.filter` |
| Image / Video | `file_id` inputs, `storage_options`, response storage fields |
| `sample.proto` | Slimmed vs upstream (enum-only) for chat shared types |

When upstream publishes identical definitions, re-sync and drop residuals.

## Wire-parity baseline & guard

Residual synthesis once drifted from the live API (compacted `reserved` field
numbers in collections/files — see CHANGELOG). The protos are now aligned to
the **official Python SDK descriptors** (`xai-sdk-python` `proto/v6`, the
freshest public artifact of the server schema; the public xai-proto repo lags
it). Guard rails:

- `xai/api/v1/wire_pin_test.go` pins the exact field numbers of every message
  that previously drifted; a failing pin means a wire-breaking edit.
- **Never compact or reuse `reserved` field numbers** when hand-editing
  residuals; deleted upstream fields keep their numbers reserved forever.
- To re-verify the whole surface, dump both sides' descriptors and diff
  (message full-name package prefixes may differ — `rag.`/`shared.` upstream
  vs `xai_api.` here — which is wire-irrelevant for messages; gRPC service
  paths must match exactly).

## Generate

```bash
make proto   # needs protoc + protoc-gen-go + protoc-gen-go-grpc
# If a service was removed from a .proto, delete stale *_grpc.pb.go before or after regen.
go test ./...
```

Do **not** hand-edit `xai/api/v1/*.pb.go` except removing orphaned `*_grpc.pb.go` after dropping a service.  
Contributor context: [`CONTRIBUTING.md`](../CONTRIBUTING.md).
