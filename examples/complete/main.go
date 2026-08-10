// Complete bilingual call example / 中英双语完整调用示例
//
// Demonstrates client construction, chat Sample, streaming, and optional tools,
// with parameter notes in English and 中文.
//
// 演示客户端构造、Chat Sample、流式输出、可选工具；参数说明中英对照。
//
//	export XAI_API_KEY=xai-...
//	go run ./examples/complete
//
// Optional env / 可选环境变量:
//
//	XAI_MODEL              chat model (default grok-3) / 对话模型
//	COMPLETE_SKIP_STREAM=1 skip streaming section / 跳过流式
//	COMPLETE_SKIP_TOOLS=1  skip server-side WebSearch / 跳过网页搜索工具
//	COMPLETE_MAX_TOKENS    max tokens (default 64) / 最大生成 token
package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/chat"
	"github.com/fun7257/xai-sdk-go/tools"
	"github.com/fun7257/xai-sdk-go/types"
)

func main() {
	// -------------------------------------------------------------------------
	// 1) Credentials / 凭证
	//    EN: Prefer XAI_API_KEY. WithAPIKey overrides env when set.
	//    中文: 优先环境变量 XAI_API_KEY；WithAPIKey 可覆盖环境变量。
	// -------------------------------------------------------------------------
	if os.Getenv("XAI_API_KEY") == "" {
		log.Fatal("XAI_API_KEY is required / 需要设置环境变量 XAI_API_KEY")
	}

	// -------------------------------------------------------------------------
	// 2) NewClient options / 客户端选项
	//
	//    | Option                    | EN                                      | 中文                    |
	//    |---------------------------|-----------------------------------------|-------------------------|
	//    | WithAPIKey                | Business key (optional if env set)      | 业务密钥（有 env 可省略） |
	//    | WithManagementAPIKey      | Collections CRUD only                   | 仅集合管理               |
	//    | WithTimeout               | Default unary RPC timeout if no deadline| 无截止时间时的默认超时   |
	//    | WithAPIHost               | Default api.x.ai:443                    | 业务 API 地址            |
	//    | WithInsecure              | No TLS (local only; Bearer still sent)  | 明文拨号（仅本地）       |
	//    | WithoutEnv                | Do not read keys from environment       | 不读环境变量密钥         |
	//
	//    Errors: ErrNoAPIKey, ErrEmptyAPIKey (errors.Is)
	//    错误: 可用 errors.Is 判断 ErrNoAPIKey / ErrEmptyAPIKey
	// -------------------------------------------------------------------------
	client, err := xai.NewClient(
		// xai.WithAPIKey(os.Getenv("XAI_API_KEY")), // optional explicit / 可显式传入
		xai.WithTimeout(3*time.Minute),
	)
	if err != nil {
		log.Fatalf("NewClient: %v", err)
	}
	defer func() {
		// EN: Always Close owned connections.
		// 中文: 务必关闭客户端持有的连接。
		if err := client.Close(); err != nil {
			log.Printf("Close: %v", err)
		}
	}()

	fmt.Printf("SDK version / SDK 版本: %s\n", xai.Version())
	fmt.Println(strings.Repeat("-", 60))

	model := envOr("XAI_MODEL", types.ModelGrok3)
	maxTokens := int32(envInt("COMPLETE_MAX_TOKENS", 64))
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// -------------------------------------------------------------------------
	// 3) Chat.Create — session options / 会话级参数
	//
	//    Create(model, opts...) builds a multi-turn session (no RPC yet).
	//    Create 只构建会话，此时不发起 RPC。
	//
	//    | Option              | Param        | EN                         | 中文           |
	//    |---------------------|--------------|----------------------------|----------------|
	//    | WithMessages        | *Message...  | Initial history            | 初始消息       |
	//    | WithMaxTokens       | int32        | Cap completion length      | 限制生成长度   |
	//    | WithTemperature     | float32      | Higher → more random       | 越高越随机     |
	//    | WithTopP            | float32      | Nucleus sampling           | 核采样         |
	//    | WithSeed            | int32        | Reproducibility (if any)   | 可复现种子     |
	//    | WithStop            | string...    | Stop sequences             | 停止序列       |
	//    | WithReasoningEffort | string       | none|low|medium|high       | 推理强度       |
	//    | WithTools           | *Tool...     | Server/client tools        | 工具列表       |
	//    | WithStoreMessages   | bool         | Persist for later Get      | 是否存储消息   |
	//
	//    Messages / 消息:
	//      chat.System / User / Assistant — panic on illegal part types
	//      chat.NewUser / NewSystem — return error instead
	//      非法 part: 前者 panic，后者返回 error
	// -------------------------------------------------------------------------
	session := client.Chat.Create(model,
		chat.WithMessages(
			chat.System("You are a concise assistant. Reply in one short sentence. / 你是简洁助手，用一句话回答。"),
		),
		chat.WithMaxTokens(maxTokens),
		chat.WithTemperature(0.4),
		// chat.WithTopP(0.95),
		// chat.WithSeed(42),
		// chat.WithReasoningEffort(types.ReasoningLow),
	)

	// Append user turn / 追加用户轮次
	if err := session.Append(chat.User("What is 2+2? Answer with the number only.")); err != nil {
		log.Fatalf("Append: %v", err)
	}

	// -------------------------------------------------------------------------
	// 4) Sample — single completion / 单次补全
	//    EN: Use Sample for n=1. For multiple results: Samples(ctx, chat.WithN(k)).
	//    中文: 单次用 Sample；多次用 Samples(ctx, chat.WithN(k))，禁止静默丢弃。
	// -------------------------------------------------------------------------
	fmt.Println("## Sample / 单次补全")
	resp, err := session.Sample(ctx)
	if err != nil {
		log.Fatalf("Sample: %v", err)
	}
	fmt.Printf("content / 内容: %s\n", resp.Content())
	if u := resp.Usage(); u != nil {
		// EN: PromptTokens / CompletionTokens / TotalTokens
		// 中文: 提示 / 补全 / 总计 token
		fmt.Printf("usage / 用量: prompt=%d completion=%d total=%d\n",
			u.GetPromptTokens(), u.GetCompletionTokens(), u.GetTotalTokens())
	}
	if cost, ok := resp.CostUSD(); ok {
		fmt.Printf("cost_usd / 费用美元: %g\n", cost)
	}
	fmt.Println(strings.Repeat("-", 60))

	// -------------------------------------------------------------------------
	// 5) StreamReader — primary streaming API / 主流式 API
	//
	//    | Step        | EN                              | 中文                 |
	//    |-------------|---------------------------------|----------------------|
	//    | StreamReader| Start stream (optional WithN)   | 开启流（可带 WithN） |
	//    | Recv        | Next event; io.EOF when done    | 下一事件；结束为 EOF |
	//    | Close       | Always; ends span / cancels RPC | 务必调用，结束 RPC   |
	//    | ev.Chunk    | Incremental delta               | 增量片段             |
	//    | ev.Response | Final aggregated (if present)   | 最终聚合（若有）     |
	// -------------------------------------------------------------------------
	if os.Getenv("COMPLETE_SKIP_STREAM") != "1" {
		fmt.Println("## Stream / 流式")
		streamChat := client.Chat.Create(model,
			chat.WithMaxTokens(maxTokens),
			chat.WithMessages(chat.User("Count from 1 to 3, one number per line.")),
		)
		sr, err := streamChat.StreamReader(ctx)
		if err != nil {
			log.Fatalf("StreamReader: %v", err)
		}
		// EN: Prefer checking Close error in library code; examples may ignore.
		// 中文: 库代码应检查 Close 错误；示例中可简化。
		defer func() { _ = sr.Close() }()

		var built strings.Builder
		for {
			ev, err := sr.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Recv: %v", err)
			}
			if ev.Chunk != nil {
				piece := ev.Chunk.Content()
				fmt.Print(piece)
				built.WriteString(piece)
			}
		}
		fmt.Printf("\n(stream complete / 流式结束, bytes=%d)\n", built.Len())
		fmt.Println(strings.Repeat("-", 60))
	}

	// -------------------------------------------------------------------------
	// 6) Server-side tools (optional) / 服务端工具（可选）
	//
	//    tools.WebSearch(opts...) returns (*Tool, error) — validates domains.
	//    allowed + excluded domains are mutually exclusive.
	//    WebSearch 校验域名；allowed 与 excluded 互斥。
	//
	//    | Option                 | EN                    | 中文         |
	//    |------------------------|-----------------------|--------------|
	//    | WithAllowedDomains     | Allow-list            | 域名白名单   |
	//    | WithExcludedDomains    | Deny-list (xor allow) | 黑名单（互斥）|
	//    | WithUserLocation       | Geo preference        | 地理位置偏好 |
	//    | UncheckedWebSearch     | Skip validation       | 跳过校验     |
	// -------------------------------------------------------------------------
	if os.Getenv("COMPLETE_SKIP_TOOLS") != "1" {
		fmt.Println("## Tools (WebSearch) / 工具（网页搜索）")
		web, err := tools.WebSearch(
		// tools.WithAllowedDomains("example.com"),
		// tools.WithUserLocation("US", "San Francisco", "CA", "America/Los_Angeles"),
		)
		if err != nil {
			log.Fatalf("WebSearch: %v", err)
		}
		toolChat := client.Chat.Create(model,
			chat.WithTools(web),
			chat.WithMaxTokens(128),
			chat.WithMessages(chat.User("In one sentence: what is xAI?")),
		)
		toolResp, err := toolChat.Sample(ctx)
		if err != nil {
			log.Fatalf("Sample(tools): %v", err)
		}
		fmt.Printf("content / 内容: %s\n", toolResp.Content())
		if usage := toolResp.ServerSideToolUsage(); len(usage) > 0 {
			fmt.Printf("server_side_tool_usage / 服务端工具用量: %v\n", usage)
		}
		fmt.Println(strings.Repeat("-", 60))
	}

	// -------------------------------------------------------------------------
	// 7) Multi-completion shape (documented; not always executed)
	//    多补全形态（说明；本示例默认不执行，避免额外费用）
	//
	//    resps, err := session.Samples(ctx, chat.WithN(2))
	//    // WithN(n): n completions; Sample rejects n>1
	//    // WithN(n): 请求 n 个补全；Sample 在 n>1 时拒绝
	// -------------------------------------------------------------------------
	fmt.Println("## Done / 完成")
	fmt.Println("See docs/GUIDE.zh-en.md for full parameter tables.")
	fmt.Println("完整参数表见 docs/GUIDE.zh-en.md")
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}
