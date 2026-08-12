package conn_test

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"

	"github.com/fun7257/xai-sdk-go/internal/conn"
	xaiv1 "github.com/fun7257/xai-sdk-go/xai/api/v1"
)

func TestResolveManagementKey(t *testing.T) {
	if got := conn.ResolveManagementKey("explicit", false); got != "explicit" {
		t.Fatalf("explicit: %q", got)
	}
	t.Setenv("XAI_MANAGEMENT_KEY", "from-env")
	if got := conn.ResolveManagementKey("", false); got != "from-env" {
		t.Fatalf("env: %q", got)
	}
	if got := conn.ResolveManagementKey("", true); got != "" {
		t.Fatalf("skipEnv must ignore env: %q", got)
	}
	if got := conn.ResolveManagementKey("explicit", true); got != "explicit" {
		t.Fatalf("explicit wins over skipEnv: %q", got)
	}
}

// deadlineCapture records the deadline and metadata each RPC arrives with.
type deadlineCapture struct {
	xaiv1.UnimplementedTokenizeServer
	deadline    time.Time
	hasDeadline bool
	md          metadata.MD
}

func (s *deadlineCapture) TokenizeText(ctx context.Context, req *xaiv1.TokenizeTextRequest) (*xaiv1.TokenizeTextResponse, error) {
	s.deadline, s.hasDeadline = ctx.Deadline()
	s.md, _ = metadata.FromIncomingContext(ctx)
	return &xaiv1.TokenizeTextResponse{}, nil
}

// dialBufconn dials through conn.Dial (real interceptor/credentials wiring)
// into an in-memory server.
func dialBufconn(t *testing.T, capture *deadlineCapture, timeout time.Duration, extraMD []string) *grpc.ClientConn {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer()
	xaiv1.RegisterTokenizeServer(srv, capture)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop(); _ = lis.Close() })

	cc, err := conn.Dial(context.Background(), "passthrough:///bufnet", "test-key", true, timeout, extraMD,
		[]grpc.DialOption{grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		})})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cc.Close() })
	return cc
}

// The unary interceptor must inject the default timeout when the caller has
// no deadline, and leave caller deadlines untouched.
func TestDialTimeoutInterceptor(t *testing.T) {
	capture := &deadlineCapture{}
	cc := dialBufconn(t, capture, 5*time.Second, nil)
	cli := xaiv1.NewTokenizeClient(cc)

	// No caller deadline: interceptor applies the 5s default.
	start := time.Now()
	if _, err := cli.TokenizeText(context.Background(), &xaiv1.TokenizeTextRequest{Text: "x", Model: "m"}); err != nil {
		t.Fatal(err)
	}
	if !capture.hasDeadline {
		t.Fatal("expected injected deadline")
	}
	if remain := capture.deadline.Sub(start); remain <= 0 || remain > 5*time.Second+time.Second {
		t.Fatalf("injected deadline out of range: %s", remain)
	}

	// Caller deadline shorter than the default is preserved.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if _, err := cli.TokenizeText(ctx, &xaiv1.TokenizeTextRequest{Text: "x", Model: "m"}); err != nil {
		t.Fatal(err)
	}
	if remain := time.Until(capture.deadline); remain > time.Second+500*time.Millisecond {
		t.Fatalf("caller deadline overridden: %s remaining", remain)
	}
}

// Timeout 0 disables deadline injection.
func TestDialZeroTimeoutMeansNoDeadline(t *testing.T) {
	capture := &deadlineCapture{}
	cc := dialBufconn(t, capture, 0, nil)
	cli := xaiv1.NewTokenizeClient(cc)
	if _, err := cli.TokenizeText(context.Background(), &xaiv1.TokenizeTextRequest{Text: "x", Model: "m"}); err != nil {
		t.Fatal(err)
	}
	if capture.hasDeadline {
		t.Fatalf("expected no deadline, got %s", time.Until(capture.deadline))
	}
}

// Bearer auth and SDK/user metadata must arrive on the wire through Dial's
// PerRPC credentials.
func TestDialAttachesAuthAndMetadata(t *testing.T) {
	capture := &deadlineCapture{}
	cc := dialBufconn(t, capture, time.Minute, []string{"x-team", "growth"})
	cli := xaiv1.NewTokenizeClient(cc)
	if _, err := cli.TokenizeText(context.Background(), &xaiv1.TokenizeTextRequest{Text: "x", Model: "m"}); err != nil {
		t.Fatal(err)
	}
	if got := capture.md.Get("authorization"); len(got) != 1 || got[0] != "Bearer test-key" {
		t.Fatalf("authorization=%v", got)
	}
	if got := capture.md.Get("xai-sdk-version"); len(got) != 1 || got[0] == "" {
		t.Fatalf("xai-sdk-version=%v", got)
	}
	if got := capture.md.Get("x-team"); len(got) != 1 || got[0] != "growth" {
		t.Fatalf("x-team=%v", got)
	}
}
