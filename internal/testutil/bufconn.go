// Package testutil provides in-memory gRPC helpers for unit tests.
package testutil

import (
	"context"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1 << 20

// Server is an in-memory gRPC server with a dialed client conn.
type Server struct {
	Lis  *bufconn.Listener
	GRPC *grpc.Server
	Conn *grpc.ClientConn
}

// Start starts a bufconn server and returns a client connection.
func Start(register func(*grpc.Server)) (*Server, error) {
	lis := bufconn.Listen(bufSize)
	s := grpc.NewServer()
	register(s)
	go func() { _ = s.Serve(lis) }()
	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) {
			return lis.DialContext(ctx)
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		s.Stop()
		return nil, err
	}
	return &Server{Lis: lis, GRPC: s, Conn: conn}, nil
}

// Close stops the server and connection.
func (s *Server) Close() {
	if s.Conn != nil {
		_ = s.Conn.Close()
	}
	if s.GRPC != nil {
		s.GRPC.Stop()
	}
	if s.Lis != nil {
		_ = s.Lis.Close()
	}
}

// AuthBearerFromContext extracts the Bearer token from incoming metadata.
func AuthBearerFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	vals := md.Get("authorization")
	if len(vals) == 0 {
		return ""
	}
	const p = "Bearer "
	if len(vals[0]) > len(p) && vals[0][:len(p)] == p {
		return vals[0][len(p):]
	}
	return vals[0]
}
