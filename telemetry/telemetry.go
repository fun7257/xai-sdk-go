// Package telemetry provides opt-in OpenTelemetry instrumentation for the SDK.
//
// By default there is no global TracerProvider side effect: spans are only
// created when Setup is called (or a Tracer is injected via SetTracer) and
// XAI_SDK_DISABLE_TRACING is not set.
//
// Environment variables:
//
//	XAI_SDK_DISABLE_TRACING                        - disable all spans
//	XAI_SDK_DISABLE_SENSITIVE_TELEMETRY_ATTRIBUTES - omit prompts/content attributes
//	XAI_SDK_INCLUDE_SENSITIVE_TELEMETRY_ATTRIBUTES - required (truthy) before
//	    AttrPrompt attaches prompt content; sensitive attributes stay off otherwise
package telemetry

import (
	"context"
	"os"
	"strings"
	"sync"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
)

const (
	tracerName = "github.com/fun7257/xai-sdk-go"
	// Span names (gen_ai style).
	SpanChatSample    = "chat.sample"
	SpanChatStream    = "chat.stream"
	SpanImageSample   = "image.sample"
	SpanVideoGenerate = "video.generate"
	SpanVideoExtend   = "video.extend"
	SpanFilesUpload   = "files.upload"
)

// noopTracer is used when tracing is disabled so StartSpan never returns a parent span.
var noopTracer = noop.NewTracerProvider().Tracer(tracerName)

func envTruthy(name string) bool {
	v := os.Getenv(name)
	return v == "1" || strings.EqualFold(v, "true")
}

// TracingDisabled is true when XAI_SDK_DISABLE_TRACING is set.
func TracingDisabled() bool { return envTruthy("XAI_SDK_DISABLE_TRACING") }

// SensitiveAttributesDisabled is true when sensitive gen_ai attributes should be omitted.
func SensitiveAttributesDisabled() bool {
	return envTruthy("XAI_SDK_DISABLE_SENSITIVE_TELEMETRY_ATTRIBUTES")
}

var (
	mu     sync.RWMutex
	tracer trace.Tracer = noop.NewTracerProvider().Tracer(tracerName)
	// setupCalled tracks whether Setup configured a non-noop provider.
	setupCalled bool
)

// SetTracer injects a custom Tracer (tests / advanced). Does not install a
// global provider. A non-nil tracer marks tracing as enabled (Enabled returns
// true); passing nil restores the noop tracer and disables it.
func SetTracer(t trace.Tracer) {
	mu.Lock()
	defer mu.Unlock()
	if t == nil {
		tracer = noop.NewTracerProvider().Tracer(tracerName)
		setupCalled = false
		return
	}
	tracer = t
	setupCalled = true
}

// Tracer returns the current tracer (noop unless Setup/SetTracer).
func Tracer() trace.Tracer {
	mu.RLock()
	defer mu.RUnlock()
	return tracer
}

// SetupOptions configures optional OTEL setup.
type SetupOptions struct {
	// TracerProvider if non-nil is used as the source of the SDK tracer.
	// When nil, otel's global TracerProvider is used (caller must configure it).
	TracerProvider trace.TracerProvider
}

// Setup enables tracing using the given options. Safe to call multiple times;
// the last call wins. Does not call otel.SetTracerProvider unless you pass a
// provider and want that yourself — by default it only reads otel.GetTracerProvider().
//
// If XAI_SDK_DISABLE_TRACING is set, Setup becomes a no-op (noop tracer).
func Setup(opts SetupOptions) {
	mu.Lock()
	defer mu.Unlock()
	if TracingDisabled() {
		tracer = noop.NewTracerProvider().Tracer(tracerName)
		setupCalled = false
		return
	}
	tp := opts.TracerProvider
	if tp == nil {
		tp = otel.GetTracerProvider()
	}
	tracer = tp.Tracer(tracerName)
	setupCalled = true
}

// Reset restores the noop tracer (for tests).
func Reset() {
	mu.Lock()
	defer mu.Unlock()
	tracer = noop.NewTracerProvider().Tracer(tracerName)
	setupCalled = false
}

// Enabled reports whether non-noop tracing is active and not env-disabled.
func Enabled() bool {
	if TracingDisabled() {
		return false
	}
	mu.RLock()
	defer mu.RUnlock()
	return setupCalled
}

// StartSpan starts a span if tracing is enabled and not disabled via env.
// Always returns a non-nil span that is safe to End.
//
// When XAI_SDK_DISABLE_TRACING is set, returns a dedicated non-recording span
// from an internal noop tracer and leaves ctx unchanged — never SpanFromContext,
// so callers can always EndSpan without truncating an application parent span.
// When not disabled, uses the configured Tracer (Setup/SetTracer; default noop).
func StartSpan(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	if TracingDisabled() {
		// Non-recording span; End is a no-op. Do not return SpanFromContext.
		_, span := noopTracer.Start(ctx, name)
		return ctx, span
	}
	t := Tracer()
	return t.Start(ctx, name, trace.WithAttributes(attrs...))
}

// EndSpan ends the span and records err if non-nil.
func EndSpan(span trace.Span, err error) {
	if span == nil {
		return
	}
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

// ChatSampleAttrs returns base attributes for a chat sample span.
func ChatSampleAttrs(model string, messageCount int) []attribute.KeyValue {
	attrs := []attribute.KeyValue{
		attribute.String("gen_ai.system", "xai"),
		attribute.String("gen_ai.request.model", model),
		attribute.String("gen_ai.operation.name", "chat"),
		attribute.Int("gen_ai.request.message_count", messageCount),
	}
	return attrs
}

// ChatRequestAttrs builds richer gen_ai-style request attributes (CHAT-07/08, TEL-02).
// Sensitive prompt/content is never attached here; use AttrPrompt explicitly.
func ChatRequestAttrs(model string, messageCount int, conversationID string, extras ...attribute.KeyValue) []attribute.KeyValue {
	attrs := ChatSampleAttrs(model, messageCount)
	if conversationID != "" {
		attrs = append(attrs, attribute.String("gen_ai.conversation.id", conversationID))
	}
	attrs = append(attrs, extras...)
	return attrs
}

// SetupConsoleRecipeDoc returns a pointer to the exporter wiring recipes.
// The SDK does not bundle exporters; callers install their own and pass a
// TracerProvider to Setup.
//
// Recipe for OTLP (TEL-01) — callers install the exporter themselves:
//
//	exp, err := otlptracegrpc.New(ctx)
//	tp := sdktrace.NewTracerProvider(sdktrace.WithBatcher(exp))
//	telemetry.Setup(telemetry.SetupOptions{TracerProvider: tp})
//
// Recipe for console/dev:
//
//	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(stdouttrace.New(stdouttrace.WithPrettyPrint())))
//	telemetry.Setup(telemetry.SetupOptions{TracerProvider: tp})
func SetupConsoleRecipeDoc() string {
	return "See Setup + SetupOptions.TracerProvider; wire otlptracegrpc or stdouttrace exporter."
}

// AttrPrompt optionally adds a prompt attribute.
//
// Sensitive content is off by default: returns nil unless
// XAI_SDK_INCLUDE_SENSITIVE_TELEMETRY_ATTRIBUTES is truthy, and still returns nil
// when XAI_SDK_DISABLE_SENSITIVE_TELEMETRY_ATTRIBUTES is set. Chat Sample/Stream
// do not attach prompt content unless callers add these attrs themselves.
func AttrPrompt(prompt string) []attribute.KeyValue {
	if prompt == "" || SensitiveAttributesDisabled() {
		return nil
	}
	if !envTruthy("XAI_SDK_INCLUDE_SENSITIVE_TELEMETRY_ATTRIBUTES") {
		return nil
	}
	return []attribute.KeyValue{attribute.String("gen_ai.prompt", prompt)}
}
