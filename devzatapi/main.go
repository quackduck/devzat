package devzatapi

import (
	"context"
	"github.com/quackduck/devzat/plugin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"time"
)

type Message struct {
	Room,
	From,
	Data,
	DMTo string
}

type CmdCall struct {
	Room,
	From,
	Args string
}

type Session struct {
	conn         *grpc.ClientConn
	pluginClient plugin.PluginClient

	ErrorChan chan error
}

// NewSession connects to the Devzat server and creates a session. The address should be in the form of "host:port".
func NewSession(address string, token string) (*Session, error) {
	backoffConfig := backoff.DefaultConfig
	backoffConfig.Multiplier = 1.1
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
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		grpc.WithConnectParams(grpc.ConnectParams{Backoff: backoffConfig, MinConnectTimeout: time.Second * 20}),
	)
	if err != nil {
		return nil, err
	}
	return &Session{conn: conn, pluginClient: plugin.NewPluginClient(conn), ErrorChan: make(chan error)}, nil
}

// Close closes the session.
func (s *Session) Close() error {
	return s.conn.Close()
}

// RegisterListener allows for message monitoring and intercepting/editing.
// Set middleware to true to intercept and edit messages.
// Set once to true to unregister the listener after the first message is received.
// Set regex to a valid regex string to only receive messages that match the regex.
//
// messageChan will receive messages that match the regex.
// middlewareResponseChan is used to send back the edited message. You must send a response if middleware is true
// even if you don't edit the message.
// See example/main.go for correct usage of this function.
func (s *Session) RegisterListener(middleware, once bool, regex string) (messageChan chan Message, middlewareResponseChan chan string, err error) {
	var client plugin.Plugin_RegisterListenerClient
	setup := func() error {
		client, err = s.pluginClient.RegisterListener(context.Background())
		if err != nil {
			return err
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
			return err
		}
		return nil
	}
	err = setup()
	if err != nil {
		return
	}

	messageChan = make(chan Message)
	var e *plugin.Event
	go func() {
		for {
			e, err = client.Recv()
			if err != nil {
				if isErrEOF(err) {
					// set up new stream
					err = setup()
					if err == nil {
						continue
					}
				}
				s.ErrorChan <- err
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
			if err != nil {
				if isErrEOF(err) {
					// set up new stream
					err = setup()
					if err == nil {
						continue
					}
				}
				s.ErrorChan <- err
				continue
			}
		}
	}()

	return
}

func (s *Session) SendMessage(m Message) error {
	if m.Data == "" {
		return nil
	}
	fromPtr := &m.From
	if m.From == "" {
		fromPtr = nil
	}
	dmToPtr := &m.DMTo
	if m.DMTo == "" {
		dmToPtr = nil
	}
	_, err := s.pluginClient.SendMessage(context.Background(), &plugin.Message{
		Room:        m.Room,
		From:        fromPtr,
		Msg:         m.Data,
		EphemeralTo: dmToPtr,
	})
	return err
}

// RegisterCmd registers a command.
// onCmd is called when the command is used.
// If onCmd returns true, the client stops listening for command usages.
func (s *Session) RegisterCmd(name, argsInfo, info string, onCmd func(CmdCall, error)) error {
	client, err := s.pluginClient.RegisterCmd(context.Background(), &plugin.CmdDef{
		Name:     name,
		ArgsInfo: argsInfo,
		Info:     info,
	})
	if err != nil {
		return err
	}
	go func() {
		for {
			i, err := client.Recv()
			if err != nil {
				if isErrEOF(err) {
					// set up new stream
					client, err = s.pluginClient.RegisterCmd(context.Background(), &plugin.CmdDef{Name: name, ArgsInfo: argsInfo, Info: info})
					if err == nil {
						continue
					}
				}
				i = &plugin.CmdInvocation{}
			}
			onCmd(CmdCall{Room: i.Room, From: i.From, Args: i.Args}, err)
		}
	}()
	return nil
}

func isErrEOF(err error) bool {
	//return errors.Is(err, io.EOF)
	return err.Error() == "rpc error: code = Unavailable desc = error reading from server: EOF"
}
