package conn_test

import (
	"context"
	"errors"
	"testing"

	"github.com/fun7257/xai-sdk-go/internal/conn"
)

func TestResolveAPIKey(t *testing.T) {
	k, err := conn.ResolveAPIKey("explicit", true)
	if err != nil || k != "explicit" {
		t.Fatalf("%q %v", k, err)
	}
	_, err = conn.ResolveAPIKey("", true)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, conn.ErrNoAPIKey) {
		t.Fatalf("want ErrNoAPIKey, got %v", err)
	}
	t.Setenv("XAI_API_KEY", "env-key")
	k, err = conn.ResolveAPIKey("", false)
	if err != nil || k != "env-key" {
		t.Fatalf("%q %v", k, err)
	}
	t.Setenv("XAI_API_KEY", "")
	_, err = conn.ResolveAPIKey("", false)
	if !errors.Is(err, conn.ErrNoAPIKey) {
		t.Fatalf("want ErrNoAPIKey on missing env, got %v", err)
	}
}

func TestDialEmptyKey(t *testing.T) {
	_, err := conn.Dial(context.Background(), "127.0.0.1:9", "", true, 0, nil, nil)
	if !errors.Is(err, conn.ErrEmptyAPIKey) {
		t.Fatalf("want ErrEmptyAPIKey, got %v", err)
	}
}

func TestBearerMetadata(t *testing.T) {
	b := conn.BearerCredentials{
		APIKey:   "k1",
		Metadata: conn.BuildMetadata([]string{"x-custom", "v"}),
		Insecure: true,
	}
	md, err := b.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if md["authorization"] != "Bearer k1" {
		t.Fatalf("auth=%q", md["authorization"])
	}
	if md["xai-sdk-version"] == "" || md["xai-sdk-language"] == "" {
		t.Fatalf("missing sdk metadata: %#v", md)
	}
	if md["x-custom"] != "v" {
		t.Fatalf("custom=%q", md["x-custom"])
	}
	if b.RequireTransportSecurity() {
		t.Fatal("insecure should not require TLS")
	}
}

func TestBearerMetadataIgnoresUserAuthorization(t *testing.T) {
	b := conn.BearerCredentials{
		APIKey:   "real-key",
		Metadata: conn.BuildMetadata([]string{"authorization", "Bearer evil", "Authorization", "Bearer evil2"}),
		Insecure: true,
	}
	md, err := b.GetRequestMetadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if md["authorization"] != "Bearer real-key" {
		t.Fatalf("auth clobbered: %q", md["authorization"])
	}
}

func TestDefaultServiceConfigRetryPolicy(t *testing.T) {
	cfg := conn.DefaultServiceConfig
	if cfg == "" {
		t.Fatal("empty service config")
	}
	for _, needle := range []string{
		`"maxAttempts": 5`,
		`"initialBackoff": "0.1s"`,
		`"maxBackoff": "1s"`,
		`"backoffMultiplier": 2`,
		`"UNAVAILABLE"`,
		`"retryPolicy"`,
	} {
		if !contains(cfg, needle) {
			t.Fatalf("service config missing %s\n%s", needle, cfg)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		}())
}
