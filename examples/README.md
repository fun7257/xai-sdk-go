# Examples

All examples read credentials from the environment. **Never commit API keys.**

| Example | Capability | Env |
|---------|------------|-----|
| [`smoke/`](smoke/) | Live multi-domain smoke | `XAI_API_KEY` |
| [`chat/`](chat/) | Chat Sample | `XAI_API_KEY` |
| [`stream/`](stream/) | StreamReader.Recv + Close | `XAI_API_KEY` |
| [`function_calling/`](function_calling/) | Client function tools | `XAI_API_KEY` |
| [`structured/`](structured/) | Chat.Parse JSON schema | `XAI_API_KEY` |
| [`image/`](image/) | Image Sample | `XAI_API_KEY` |
| [`video/`](video/) | Video Generate | `XAI_API_KEY` |
| [`files/`](files/) | Files List | `XAI_API_KEY` |
| [`batch/`](batch/) | Batch list | `XAI_API_KEY` |
| [`collections/`](collections/) | Collections list | `XAI_API_KEY`, `XAI_MANAGEMENT_KEY` |
| [`telemetry/`](telemetry/) | Opt-in OTEL Setup | `XAI_API_KEY` |
| [`deferred/`](deferred/) | Chat.Defer | `XAI_API_KEY` |
| [`stored/`](stored/) | store_messages + GetStoredCompletion | `XAI_API_KEY` |
| [`compaction/`](compaction/) | Chat.Compact | `XAI_API_KEY` |
| [`server_side_tools/`](server_side_tools/) | WebSearch / CodeExecution | `XAI_API_KEY` |
| [`reasoning/`](reasoning/) | reasoning_effort + ReasoningContent | `XAI_API_KEY` |
| [`image_understanding/`](image_understanding/) | chat.Image in messages | `XAI_API_KEY` |
| [`files_chat/`](files_chat/) | FileByID | `XAI_API_KEY`, `XAI_FILE_ID` |
| [`inline_files_chat/`](inline_files_chat/) | File data mode | `XAI_API_KEY` |
| [`models/`](models/) | ListLanguageModels | `XAI_API_KEY` |
| [`auth/`](auth/) | GetAPIKeyInfo | `XAI_API_KEY` |
| [`tokenizer/`](tokenizer/) | TokenizeText | `XAI_API_KEY` |
| [`video_extension/`](video_extension/) | ExtendStart | `XAI_API_KEY`, `XAI_VIDEO_URL` |
| [`image_multi_reference/`](image_multi_reference/) | Samples multi refs (WithN / multi image inputs) | `XAI_API_KEY` |
| [`collections_tool/`](collections_tool/) | CollectionsSearch tool | `XAI_API_KEY` |

```bash
export XAI_API_KEY=xai-...
go run ./examples/chat
go build ./examples/...
```
