// Live smoke for xai-sdk-go — low-cost completeness check with chat + image + video.
//
// Covers: NewClient, Auth, Models, Tokenize, Chat Sample/Stream, Image Sample,
// Video Generate (short), Close.
//
// Skips: Batch / Collections (management key / long jobs).
//
// Usage:
//
//	export XAI_API_KEY=xai-...
//	go run ./examples/smoke
//
// Optional env:
//
//	SMOKE_MODEL=grok-4.5-latest          // chat (default)
//	SMOKE_IMAGE_MODEL=grok-imagine-image // image (default)
//	SMOKE_VIDEO_MODEL=grok-imagine-video // video (default)
//	SMOKE_SKIP_STREAM=1
//	SMOKE_SKIP_IMAGE=1
//	SMOKE_SKIP_VIDEO=1
//	SMOKE_VIDEO_DURATION=1               // seconds, default 1 (min cost)
package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	xai "github.com/fun7257/xai-sdk-go"
	"github.com/fun7257/xai-sdk-go/chat"
	"github.com/fun7257/xai-sdk-go/image"
	"github.com/fun7257/xai-sdk-go/types"
	"github.com/fun7257/xai-sdk-go/video"
)

type step struct {
	name    string
	timeout time.Duration
	fn      func(context.Context, *xai.Client) error
}

func main() {
	key := os.Getenv("XAI_API_KEY")
	if key == "" {
		fmt.Fprintln(os.Stderr, "XAI_API_KEY is required (do not hardcode keys in source)")
		os.Exit(2)
	}

	chatModel := envOr("SMOKE_MODEL", types.ModelGrok45Latest)
	imageModel := envOr("SMOKE_IMAGE_MODEL", types.ModelImagineImage)
	videoModel := envOr("SMOKE_VIDEO_MODEL", types.ModelImagineVideo)
	videoDur := int32(1)
	if v := os.Getenv("SMOKE_VIDEO_DURATION"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			videoDur = int32(n)
		}
	}

	client, err := xai.NewClient(
		xai.WithAPIKey(key),
		// Default client timeout is long; per-step contexts still bound each call.
		xai.WithTimeout(15*time.Minute),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "NewClient: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	steps := []step{
		{"auth.GetAPIKeyInfo", 45 * time.Second, stepAuth},
		{"models.ListLanguageModels", 45 * time.Second, stepModels},
		{"tokenize.TokenizeText", 45 * time.Second, func(ctx context.Context, c *xai.Client) error {
			return stepTokenize(ctx, c, chatModel)
		}},
		{"chat.Sample (max_tokens=16)", 60 * time.Second, func(ctx context.Context, c *xai.Client) error {
			return stepChatSample(ctx, c, chatModel)
		}},
	}
	if os.Getenv("SMOKE_SKIP_STREAM") != "1" {
		steps = append(steps, step{
			name:    "chat.Stream (max_tokens=16)",
			timeout: 60 * time.Second,
			fn: func(ctx context.Context, c *xai.Client) error {
				return stepChatStream(ctx, c, chatModel)
			},
		})
	}
	if os.Getenv("SMOKE_SKIP_IMAGE") != "1" {
		steps = append(steps, step{
			name:    "image.Sample",
			timeout: 2 * time.Minute,
			fn: func(ctx context.Context, c *xai.Client) error {
				return stepImage(ctx, c, imageModel)
			},
		})
	}
	if os.Getenv("SMOKE_SKIP_VIDEO") != "1" {
		steps = append(steps, step{
			name:    fmt.Sprintf("video.Generate (duration=%ds,480p)", videoDur),
			timeout: 12 * time.Minute,
			fn: func(ctx context.Context, c *xai.Client) error {
				return stepVideo(ctx, c, videoModel, videoDur)
			},
		})
	}

	var failed int
	fmt.Printf("xai-sdk-go smoke  version=%s\n", xai.Version)
	fmt.Printf("  chat=%s  image=%s  video=%s\n", chatModel, imageModel, videoModel)
	fmt.Println(strings.Repeat("-", 60))
	for _, s := range steps {
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		start := time.Now()
		err := s.fn(ctx, client)
		cancel()
		elapsed := time.Since(start).Round(time.Millisecond)
		if err != nil {
			failed++
			fmt.Printf("FAIL  %-40s  %s\n       %v\n", s.name, elapsed, err)
			continue
		}
		fmt.Printf("OK    %-40s  %s\n", s.name, elapsed)
	}
	fmt.Println(strings.Repeat("-", 60))
	if failed > 0 {
		fmt.Printf("result: %d/%d failed\n", failed, len(steps))
		os.Exit(1)
	}
	fmt.Printf("result: all %d steps passed\n", len(steps))
	fmt.Println("note: batch/collections not exercised (management key / long jobs)")
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func stepAuth(ctx context.Context, c *xai.Client) error {
	info, err := c.Auth.GetAPIKeyInfo(ctx)
	if err != nil {
		return err
	}
	fmt.Printf("       key=%s  team=%s  name=%q\n",
		info.GetRedactedApiKey(), info.GetTeamId(), info.GetName())
	return nil
}

func stepModels(ctx context.Context, c *xai.Client) error {
	models, err := c.Models.ListLanguageModels(ctx)
	if err != nil {
		return err
	}
	if len(models) == 0 {
		return fmt.Errorf("empty model list")
	}
	// Prefer showing grok-4.5* if present.
	var hit []string
	for _, m := range models {
		n := m.GetName()
		if strings.Contains(n, "4.5") || strings.Contains(n, "grok-4") {
			hit = append(hit, n)
			if len(hit) >= 5 {
				break
			}
		}
	}
	if len(hit) == 0 {
		for i := 0; i < len(models) && i < 3; i++ {
			hit = append(hit, models[i].GetName())
		}
	}
	fmt.Printf("       count=%d  grok4x≈%v\n", len(models), hit)
	return nil
}

func stepTokenize(ctx context.Context, c *xai.Client, model string) error {
	toks, err := c.Tokenize.TokenizeText(ctx, "hi", model)
	if err != nil {
		// Tokenize may not support every alias; fall back once.
		for _, fb := range []string{types.ModelGrok45, "grok-4", "grok-3"} {
			if fb == model {
				continue
			}
			toks, err = c.Tokenize.TokenizeText(ctx, "hi", fb)
			if err == nil {
				model = fb
				break
			}
		}
		if err != nil {
			return err
		}
	}
	if len(toks) == 0 {
		return fmt.Errorf("no tokens returned")
	}
	fmt.Printf("       model=%s  tokens=%d  first=%q\n",
		model, len(toks), toks[0].GetStringToken())
	return nil
}

func stepChatSample(ctx context.Context, c *xai.Client, model string) error {
	ch := c.Chat.Create(model,
		chat.WithMessages(chat.User("Reply with exactly: pong")),
		chat.WithMaxTokens(16),
		chat.WithTemperature(0),
	)
	resp, err := ch.Sample(ctx)
	if err != nil {
		return err
	}
	content := strings.TrimSpace(resp.Content())
	if content == "" {
		return fmt.Errorf("empty content")
	}
	usd, ok := resp.CostUSD()
	printUsage("content="+quote(truncate(content, 50)), usd, ok, resp.Usage())
	return nil
}

func stepChatStream(ctx context.Context, c *xai.Client, model string) error {
	ch := c.Chat.Create(model,
		chat.WithMessages(chat.User("Reply with exactly: ok")),
		chat.WithMaxTokens(16),
		chat.WithTemperature(0),
	)
	events, errc := ch.Stream(ctx)
	var last *chat.Response
	chunks := 0
	for ev := range events {
		last = ev.Response
		if ev.Chunk != nil && ev.Chunk.Content() != "" {
			chunks++
		}
	}
	if err := <-errc; err != nil {
		return err
	}
	if last == nil || strings.TrimSpace(last.Content()) == "" {
		return fmt.Errorf("empty stream accumulation")
	}
	fmt.Printf("       chunks_with_text≈%d  content=%q\n", chunks, truncate(last.Content(), 50))
	return nil
}

func stepImage(ctx context.Context, c *xai.Client, model string) error {
	// Single 1k URL image — cheapest image path.
	resp, err := c.Image.Sample(ctx,
		"A simple solid blue circle on white background, minimal",
		model,
		image.WithFormatURL(),
		image.WithAspectRatio("1:1"),
		image.WithResolution("1k"),
		image.WithN(1),
	)
	if err != nil {
		return err
	}
	url, err := resp.URL()
	if err != nil {
		return err
	}
	if url == "" {
		return fmt.Errorf("empty image url")
	}
	fmt.Printf("       url=%s\n", truncate(url, 72))
	usd, ok := resp.CostUSD()
	printUsage("image", usd, ok, resp.Usage())
	return nil
}

func stepVideo(ctx context.Context, c *xai.Client, model string, duration int32) error {
	// Shortest practical T2V: 1s @ 480p to limit spend on a $3 key.
	resp, err := c.Video.Generate(ctx,
		"A single blue circle gently pulsing on white, minimal motion",
		model,
		video.WithDuration(duration),
		video.WithAspectRatio("1:1"),
		video.WithResolution("480p"),
		video.WithPollInterval(2*time.Second),
		video.WithPollTimeout(10*time.Minute),
	)
	if err != nil {
		return err
	}
	url, err := resp.URL()
	if err != nil {
		return err
	}
	if url == "" {
		return fmt.Errorf("empty video url")
	}
	fmt.Printf("       url=%s  duration=%ds\n", truncate(url, 64), resp.Duration())
	usd, ok := resp.CostUSD()
	printUsage("video", usd, ok, resp.Usage())
	return nil
}

type usageGetter interface {
	GetTotalTokens() int32
}

func printUsage(label string, usd float64, ok bool, u usageGetter) {
	if u != nil && ok {
		fmt.Printf("       %s  cost_usd=%.6f  usage_total=%d\n", label, usd, u.GetTotalTokens())
		return
	}
	if ok {
		fmt.Printf("       %s  cost_usd=%.6f\n", label, usd)
		return
	}
	if u != nil {
		fmt.Printf("       %s  usage_total=%d\n", label, u.GetTotalTokens())
		return
	}
	fmt.Printf("       %s\n", label)
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", "\\n")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func quote(s string) string { return strconv.Quote(s) }
