package conn

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

// SDKVersion is set by the root package init to xai.Version.
// Default matches version.go so metadata stays correct if init order differs.
var SDKVersion = "0.2.0"

// ResolveAPIKey returns an explicit key or XAI_API_KEY.
// Missing key paths wrap [ErrNoAPIKey] for errors.Is.
func ResolveAPIKey(explicit string, skipEnv bool) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if skipEnv {
		return "", fmt.Errorf("API key not provided: %w", ErrNoAPIKey)
	}
	if v := os.Getenv("XAI_API_KEY"); v != "" {
		return v, nil
	}
	return "", fmt.Errorf("trying to read the xAI API key from the XAI_API_KEY environment variable but it doesn't exist: %w", ErrNoAPIKey)
}

// ResolveManagementKey returns optional management key.
func ResolveManagementKey(explicit string, skipEnv bool) string {
	if explicit != "" {
		return explicit
	}
	if skipEnv {
		return ""
	}
	return os.Getenv("XAI_MANAGEMENT_KEY")
}

// DefaultServiceConfig is the gRPC service config used for UNAVAILABLE retries,
// aligned with the Python SDK defaults (maxAttempts=5, 0.1s→1s, multiplier 2).
const DefaultServiceConfig = `{
  "methodConfig": [{
    "name": [{}],
    "retryPolicy": {
      "maxAttempts": 5,
      "initialBackoff": "0.1s",
      "maxBackoff": "1s",
      "backoffMultiplier": 2,
      "retryableStatusCodes": ["UNAVAILABLE"]
    }
  }]
}`

// Dial creates a gRPC connection with Bearer auth and SDK metadata.
func Dial(ctx context.Context, target, apiKey string, insecureDial bool, timeout time.Duration, extraMD []string, extra []grpc.DialOption) (*grpc.ClientConn, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("empty xAI API key provided: %w", ErrEmptyAPIKey)
	}
	const mib = 1 << 20
	opts := []grpc.DialOption{
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(20*mib),
			grpc.MaxCallSendMsgSize(20*mib),
		),
		grpc.WithDefaultServiceConfig(DefaultServiceConfig),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
		grpc.WithUnaryInterceptor(timeoutUnary(timeout)),
		grpc.WithPerRPCCredentials(BearerCredentials{
			APIKey:   apiKey,
			Metadata: BuildMetadata(extraMD),
			Insecure: insecureDial,
		}),
	}
	if insecureDial {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewTLS(nil)))
	}
	opts = append(opts, extra...)
	// grpc.NewClient is the supported API; DialContext is deprecated.
	// Dial is lazy; connection errors surface on the first RPC.
	_ = ctx
	return grpc.NewClient(target, opts...)
}

// BuildMetadata constructs SDK + user metadata.
func BuildMetadata(extra []string) metadata.MD {
	md := metadata.MD{}
	md.Set("xai-sdk-version", "go/"+SDKVersion)
	md.Set("xai-sdk-language", "go/"+strings.TrimPrefix(runtime.Version(), "go"))
	for i := 0; i+1 < len(extra); i += 2 {
		md.Append(extra[i], extra[i+1])
	}
	return md
}

// BearerCredentials implements grpc.PerRPCCredentials.
type BearerCredentials struct {
	APIKey   string
	Metadata metadata.MD
	Insecure bool
}

// GetRequestMetadata attaches authorization and SDK metadata.
//
// User metadata from WithMetadata is applied first; the SDK always forces
// authorization to Bearer <APIKey> last so callers cannot clobber the token
// via a custom "authorization" metadata key. Prefer not setting reserved keys
// (authorization, xai-sdk-version, xai-sdk-language) in user metadata.
func (b BearerCredentials) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	out := map[string]string{}
	for k, vals := range b.Metadata {
		if len(vals) == 0 {
			continue
		}
		lk := strings.ToLower(k)
		// Skip reserved keys that the SDK owns; still force Bearer below.
		if lk == "authorization" {
			continue
		}
		out[k] = vals[0]
	}
	out["authorization"] = "Bearer " + b.APIKey
	return out, nil
}

// RequireTransportSecurity reports whether TLS is required.
func (b BearerCredentials) RequireTransportSecurity() bool { return !b.Insecure }

func timeoutUnary(d time.Duration) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if d > 0 {
			if _, ok := ctx.Deadline(); !ok {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, d)
				defer cancel()
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
