# Examples

Credentials from the environment only. **Never commit API keys.**

| Start here | Description |
|------------|-------------|
| [`complete/`](complete/) | Annotated end-to-end (Sample + Stream + tools), EN/中文 comments |
| [`../docs/GUIDE.zh-en.md`](../docs/GUIDE.zh-en.md) | Full parameter tables (bilingual) |
| [`../README.md`](../README.md) | Install & quick start |

```bash
export XAI_API_KEY=xai-...
go run ./examples/complete
go build ./examples/...
```

## Catalog

| Example | Capability | Env |
|---------|------------|-----|
| [`complete/`](complete/) | Full walkthrough | `XAI_API_KEY` |
| [`smoke/`](smoke/) | Multi-domain live smoke | `XAI_API_KEY` |
| [`chat/`](chat/) | Chat Sample | `XAI_API_KEY` |
| [`stream/`](stream/) | StreamReader | `XAI_API_KEY` |
| [`function_calling/`](function_calling/) | Client function tools | `XAI_API_KEY` |
| [`structured/`](structured/) | Chat.Parse | `XAI_API_KEY` |
| [`image/`](image/) | Image Sample | `XAI_API_KEY` |
| [`image_multi_reference/`](image_multi_reference/) | Samples + multi refs | `XAI_API_KEY` |
| [`image_understanding/`](image_understanding/) | chat.Image parts | `XAI_API_KEY` |
| [`video/`](video/) | Video Generate | `XAI_API_KEY` |
| [`video_extension/`](video_extension/) | ExtendStart | `XAI_API_KEY`, `XAI_VIDEO_URL` |
| [`files/`](files/) | Files List | `XAI_API_KEY` |
| [`files_chat/`](files_chat/) | FileByID | `XAI_API_KEY`, `XAI_FILE_ID` |
| [`inline_files_chat/`](inline_files_chat/) | Inline file data | `XAI_API_KEY` |
| [`batch/`](batch/) | Batch list | `XAI_API_KEY` |
| [`collections/`](collections/) | Collections list | `XAI_API_KEY`, `XAI_MANAGEMENT_KEY` |
| [`collections_tool/`](collections_tool/) | CollectionsSearch tool | `XAI_API_KEY` |
| [`telemetry/`](telemetry/) | OTEL Setup | `XAI_API_KEY` |
| [`deferred/`](deferred/) | Chat.Defer | `XAI_API_KEY` |
| [`stored/`](stored/) | store + GetStoredCompletion | `XAI_API_KEY` |
| [`compaction/`](compaction/) | Chat.Compact | `XAI_API_KEY` |
| [`server_side_tools/`](server_side_tools/) | WebSearch / CodeExecution | `XAI_API_KEY` |
| [`reasoning/`](reasoning/) | reasoning_effort | `XAI_API_KEY` |
| [`models/`](models/) | ListLanguageModels | `XAI_API_KEY` |
| [`auth/`](auth/) | GetAPIKeyInfo | `XAI_API_KEY` |
| [`tokenizer/`](tokenizer/) | TokenizeText | `XAI_API_KEY` |

Optional repo integration tests (not under `examples/`): `make integration` — see [`CONTRIBUTING.md`](../CONTRIBUTING.md).
