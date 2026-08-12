package telemetry_test

import (
	"testing"

	"go.opentelemetry.io/otel/trace/noop"

	"github.com/fun7257/xai-sdk-go/telemetry"
)

// SetTracer must toggle Enabled in line with its documentation.
func TestSetTracerTogglesEnabled(t *testing.T) {
	t.Cleanup(telemetry.Reset)

	telemetry.Reset()
	if telemetry.Enabled() {
		t.Fatal("enabled after Reset")
	}
	telemetry.SetTracer(noop.NewTracerProvider().Tracer("custom"))
	if !telemetry.Enabled() {
		t.Fatal("SetTracer(non-nil) must enable tracing")
	}
	telemetry.SetTracer(nil)
	if telemetry.Enabled() {
		t.Fatal("SetTracer(nil) must disable tracing")
	}
}

// Truthy env parsing must accept common capitalizations.
func TestTracingDisabledAcceptsMixedCase(t *testing.T) {
	for _, v := range []string{"1", "true", "TRUE", "True"} {
		t.Setenv("XAI_SDK_DISABLE_TRACING", v)
		if !telemetry.TracingDisabled() {
			t.Fatalf("value %q must disable tracing", v)
		}
	}
	t.Setenv("XAI_SDK_DISABLE_TRACING", "0")
	if telemetry.TracingDisabled() {
		t.Fatal("0 must not disable tracing")
	}
}
