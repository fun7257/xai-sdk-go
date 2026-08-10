package telemetry_test

import (
	"context"
	"testing"

	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"

	"github.com/fun7257/xai-sdk-go/telemetry"
)

func TestFlags(t *testing.T) {
	t.Setenv("XAI_SDK_DISABLE_TRACING", "1")
	if !telemetry.TracingDisabled() {
		t.Fatal("expected disabled")
	}
	t.Setenv("XAI_SDK_DISABLE_SENSITIVE_TELEMETRY_ATTRIBUTES", "true")
	if !telemetry.SensitiveAttributesDisabled() {
		t.Fatal("expected sensitive disabled")
	}
}

func TestSetupAndSpans(t *testing.T) {
	t.Setenv("XAI_SDK_DISABLE_TRACING", "")
	telemetry.Reset()
	defer telemetry.Reset()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	telemetry.Setup(telemetry.SetupOptions{TracerProvider: tp})
	if !telemetry.Enabled() {
		t.Fatal("expected enabled")
	}

	ctx, span := telemetry.StartSpan(context.Background(), telemetry.SpanChatSample,
		telemetry.ChatSampleAttrs("grok-3", 2)...,
	)
	_ = ctx
	telemetry.EndSpan(span, nil)

	spans := exporter.GetSpans()
	if len(spans) != 1 {
		t.Fatalf("spans=%d", len(spans))
	}
	if spans[0].Name != telemetry.SpanChatSample {
		t.Fatalf("name=%s", spans[0].Name)
	}
	found := false
	for _, a := range spans[0].Attributes {
		if a.Key == attribute.Key("gen_ai.request.model") && a.Value.AsString() == "grok-3" {
			found = true
		}
	}
	if !found {
		t.Fatalf("attrs=%v", spans[0].Attributes)
	}
}

func TestDisableEnvNoSpan(t *testing.T) {
	t.Setenv("XAI_SDK_DISABLE_TRACING", "1")
	telemetry.Reset()
	defer telemetry.Reset()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	telemetry.Setup(telemetry.SetupOptions{TracerProvider: tp})
	if telemetry.Enabled() {
		t.Fatal("should not be enabled when env disables")
	}
	_, span := telemetry.StartSpan(context.Background(), "x")
	telemetry.EndSpan(span, nil)
	if len(exporter.GetSpans()) != 0 {
		t.Fatalf("unexpected spans: %v", exporter.GetSpans())
	}
}

func TestSensitivePromptAttr(t *testing.T) {
	// Off by default without explicit include.
	t.Setenv("XAI_SDK_INCLUDE_SENSITIVE_TELEMETRY_ATTRIBUTES", "")
	t.Setenv("XAI_SDK_DISABLE_SENSITIVE_TELEMETRY_ATTRIBUTES", "")
	if len(telemetry.AttrPrompt("secret")) != 0 {
		t.Fatal("expected no prompt attr by default")
	}
	t.Setenv("XAI_SDK_INCLUDE_SENSITIVE_TELEMETRY_ATTRIBUTES", "1")
	if len(telemetry.AttrPrompt("secret")) != 1 {
		t.Fatal("expected prompt attr when include enabled")
	}
	t.Setenv("XAI_SDK_DISABLE_SENSITIVE_TELEMETRY_ATTRIBUTES", "1")
	if len(telemetry.AttrPrompt("secret")) != 0 {
		t.Fatal("expected no prompt attr when disable wins")
	}
}

func TestDisableDoesNotReturnParentSpan(t *testing.T) {
	t.Setenv("XAI_SDK_DISABLE_TRACING", "")
	telemetry.Reset()
	defer telemetry.Reset()

	exporter := tracetest.NewInMemoryExporter()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSyncer(exporter))
	telemetry.Setup(telemetry.SetupOptions{TracerProvider: tp})

	// Parent span on context from app tracer.
	parentCtx, parent := tp.Tracer("app").Start(context.Background(), "parent")
	t.Setenv("XAI_SDK_DISABLE_TRACING", "1")
	_, span := telemetry.StartSpan(parentCtx, "sdk.span")
	telemetry.EndSpan(span, nil)
	// Parent must still be recording (not ended by SDK).
	if !parent.IsRecording() {
		t.Fatal("parent span was ended by SDK EndSpan while tracing disabled")
	}
	parent.End()
}

func TestChatRequestAttrsConversationID(t *testing.T) {
	attrs := telemetry.ChatRequestAttrs("grok-3", 2, "conv-1", attribute.Float64("gen_ai.request.temperature", 0.5))
	foundConv, foundTemp := false, false
	for _, a := range attrs {
		if a.Key == "gen_ai.conversation.id" && a.Value.AsString() == "conv-1" {
			foundConv = true
		}
		if a.Key == "gen_ai.request.temperature" {
			foundTemp = true
		}
	}
	if !foundConv || !foundTemp {
		t.Fatalf("attrs=%v", attrs)
	}
}
