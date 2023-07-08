package devzatapi

import (
	"context"
	"devzat/plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

type Message struct {
	Room,
	From,
	Data string
}

type Session struct {
	conn         *grpc.ClientConn
	pluginClient plugin.PluginClient

	LastError error
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
	return &Session{conn: conn, pluginClient: plugin.NewPluginClient(conn)}, nil
}

// Close closes the session.
func (s *Session) Close() error {
	return s.conn.Close()
}

func (s *Session) RegisterListener(middleware, once bool, regex string) (messageChan chan Message, middlewareResponseChan chan string, err error) {
	client, err := s.pluginClient.RegisterListener(context.Background())
	if err != nil {
		return nil, nil, err
	}
	pointerRegex := &regex
	if regex == "" {
		pointerRegex = nil
	}
	err = client.Send(&plugin.ListenerClientData{Data: &plugin.ListenerClientData_Listener{Listener: &plugin.Listener{
		Middleware: &middleware,
		Once:       &once,
		Regex:      pointerRegex,
	}}})
	if err != nil {
		return nil, nil, err
	}
	messageChan = make(chan Message)
	go func() {
		for {
			e, err := client.Recv()
			if err != nil {
				s.LastError = err
				continue
			}
			messageChan <- Message{Room: e.Room, From: e.From, Data: e.Msg}
		}
	}()
	if !middleware {
		return messageChan, nil, nil
	}
	middlewareResponseChan = make(chan string)
	go func() {
		for {
			response := <-middlewareResponseChan
			err := client.Send(&plugin.ListenerClientData{Data: &plugin.ListenerClientData_Response{Response: &plugin.MiddlewareResponse{Msg: &response}}})
			if err != nil {
				s.LastError = err
				return
			}
		}
	}()
	return messageChan, middlewareResponseChan, nil
}

func (s *Session) SendMessage(room, from, msg, dmTo string) error {
	fromPtr := &from
	if from == "" {
		fromPtr = nil
	}
	dmToPtr := &dmTo
	if dmTo == "" {
		dmToPtr = nil
	}
	_, err := s.pluginClient.SendMessage(context.Background(), &plugin.Message{
		Room:        room,
		From:        fromPtr,
		Msg:         msg,
		EphemeralTo: dmToPtr,
	})
	return err
}
