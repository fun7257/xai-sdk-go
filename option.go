package xai

import (
	"time"

	"google.golang.org/grpc"
)

// Option configures Client construction.
type Option func(*clientConfig)

type clientConfig struct {
	apiKey            string
	managementAPIKey  string
	apiHost           string
	managementAPIHost string
	insecure          bool
	timeout           time.Duration
	metadata          []string
	dialOpts          []grpc.DialOption
	apiConn           grpc.ClientConnInterface
	managementConn    grpc.ClientConnInterface
	skipEnv           bool
}

func defaultConfig() clientConfig {
	return clientConfig{
		apiHost:           "api.x.ai:443",
		managementAPIHost: "management-api.x.ai:443",
		timeout:           27 * time.Minute,
	}
}

// WithAPIKey sets the business API key (overrides XAI_API_KEY).
func WithAPIKey(key string) Option {
	return func(c *clientConfig) { c.apiKey = key }
}

// WithManagementAPIKey sets the management API key (overrides XAI_MANAGEMENT_KEY).
func WithManagementAPIKey(key string) Option {
	return func(c *clientConfig) { c.managementAPIKey = key }
}

// WithAPIHost sets the API dial target (default api.x.ai:443).
func WithAPIHost(host string) Option {
	return func(c *clientConfig) { c.apiHost = host }
}

// WithManagementAPIHost sets the management dial target.
func WithManagementAPIHost(host string) Option {
	return func(c *clientConfig) { c.managementAPIHost = host }
}

// WithInsecure dials without TLS; Bearer auth is still attached, so the API
// key travels in cleartext. Only use against local/test endpoints you control.
func WithInsecure() Option {
	return func(c *clientConfig) { c.insecure = true }
}

// WithTimeout sets the default per-RPC timeout for unary calls without a deadline.
func WithTimeout(d time.Duration) Option {
	return func(c *clientConfig) { c.timeout = d }
}

// WithMetadata appends custom gRPC metadata key/value pairs (flattened:
// k1,v1,k2,v2,...). NewClient returns an error when the total element count is
// odd. Only the first value of a repeated key is sent (PerRPCCredentials carry
// a flat map); avoid the reserved keys authorization, xai-sdk-version, and
// xai-sdk-language.
func WithMetadata(kv ...string) Option {
	return func(c *clientConfig) { c.metadata = append(c.metadata, kv...) }
}

// WithDialOptions appends raw grpc.DialOptions.
func WithDialOptions(opts ...grpc.DialOption) Option {
	return func(c *clientConfig) { c.dialOpts = append(c.dialOpts, opts...) }
}

// WithAPIConn injects a pre-built API connection (tests / custom dial).
func WithAPIConn(conn grpc.ClientConnInterface) Option {
	return func(c *clientConfig) { c.apiConn = conn }
}

// WithManagementConn injects a pre-built management connection.
func WithManagementConn(conn grpc.ClientConnInterface) Option {
	return func(c *clientConfig) { c.managementConn = conn }
}

// WithoutEnv disables reading keys from the environment.
func WithoutEnv() Option {
	return func(c *clientConfig) { c.skipEnv = true }
}
