# Proto sources

## Source of truth

Official protos from [xai-org/xai-proto](https://github.com/xai-org/xai-proto) live under `third_party/xai/api/v1/`.  
Generated Go: `xai/api/v1/` with:

```text
option go_package = "github.com/fun7257/xai-sdk-go/xai/api/v1;xaiv1";
```

**Last synced upstream commit:** `af1be87b733dc177c0857fbd624f1ff12128fbd2` (see CHANGELOG).

Import stability of `xaiv1` types: [`COMPATIBILITY.md`](COMPATIBILITY.md).

## Residual fields (local synthesis)

Public xai-proto may lag the live API / Python descriptors. This repo keeps **local residual** patches for wire parity:

| Area | Residual |
|------|----------|
| Collections | `collections.proto` / related management messages |
| Files | Public URL RPCs; `ListFilesRequest.filter` |
| Image / Video | `file_id` inputs, `storage_options`, response storage fields |

When upstream publishes identical definitions, re-sync and drop residuals.

## Generate

```bash
make proto   # needs protoc + protoc-gen-go + protoc-gen-go-grpc
go test ./...
```

Do **not** hand-edit `xai/api/v1/*.pb.go`. Contributor context: [`CONTRIBUTING.md`](../CONTRIBUTING.md).
