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
	Error error
}

type CmdInvocation struct {
	Room,
	From,
	Args string
	Error error
}

type Session struct {
	conn         *grpc.ClientConn
	pluginClient plugin.PluginClient

	LastError error
}

// NewSession connects to the Devzat server and creates a session. The address should be in the form of "host:port".
func NewSession(address string, token string) (*Session, error) {
	conn, err := grpc.Dial(address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStreamInterceptor(
			func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
				return streamer(ctx, desc, cc, method, opts...)
			},
		),
		grpc.WithUnaryInterceptor(
			func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
				ctx = metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
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

// check s.LastError when sending a response and message.Error when reading messages.
func (s *Session) RegisterListener(middleware, once bool, regex string) (messageChan chan Message, middlewareResponseChan chan string, err error) {
	client, err := s.pluginClient.RegisterListener(context.Background())
	if err != nil {
		return
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
		return
	}

	messageChan = make(chan Message)
	var e *plugin.Event
	go func() {
		for {
			e, err = client.Recv()
			if err != nil {
				messageChan <- Message{Error: err}
				continue
			}
			messageChan <- Message{Room: e.Room, From: e.From, Data: e.Msg}
		}
	}()

	if !middleware {
		return
	}

	middlewareResponseChan = make(chan string)
	go func() {
		for {
			response := <-middlewareResponseChan
			err := client.Send(&plugin.ListenerClientData{Data: &plugin.ListenerClientData_Response{Response: &plugin.MiddlewareResponse{Msg: &response}}})
			if err != nil { // TODO: users can miss an error if they check too fast. use an error channel?
				s.LastError = err
				return
			}
		}
	}()

	return
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

// read CmdInvocation.Error each time
func (s *Session) RegisterCmd(name, argsInfo, info string) (chan CmdInvocation, error) {
	client, err := s.pluginClient.RegisterCmd(context.Background(), &plugin.CmdDef{
		Name:     name,
		ArgsInfo: argsInfo,
		Info:     info,
	})
	if err != nil {
		return nil, err
	}
	cmdInvocChan := make(chan CmdInvocation)
	go func() {
		for {
			invoc, err := client.Recv()
			if err != nil {
				cmdInvocChan <- CmdInvocation{Error: err}
				continue
			}
			cmdInvocChan <- CmdInvocation{Room: invoc.Room, From: invoc.From, Args: invoc.Args}
		}
	}()
	return cmdInvocChan, nil
}
