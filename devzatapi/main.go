package devzatapi

import (
	"context"
	"devzat/plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Session struct {
	conn         *grpc.ClientConn
	pluginClient plugin.PluginClient
}

// NewSession connects to the Devzat server and creates a session. The address should be in the form of "host:port".
func NewSession(address string, token string, name string) (*Session, error) {
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStreamInterceptor(
			func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
				return streamer(ctx, desc, cc, method, opts...)
			},
		),
		grpc.WithUnaryInterceptor(
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
				return invoker(ctx, method, req, reply, cc, opts...)
			},
		),
	)
	if err != nil {
		return nil, err
	}
	return &Session{conn, plugin.NewPluginClient(conn)}, nil
}

// Close closes the session.
func (s *Session) Close() error {
	return s.conn.Close()
}

func (s *Session) RegisterListener() error {
	client, err := s.pluginClient.RegisterListener(context.Background())
	if err != nil {
		return err
	}
	for {
		_, err = client.Recv()
		if err != nil {
			return err
		}
	}
}
