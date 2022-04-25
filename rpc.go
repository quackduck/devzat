package main

import (
	"context"
	"fmt"
	"github.com/acarl005/stripansi"
	"io"
	"net"
	"time"

	pb "devzat/plugin"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var PluginCMDs = map[string]cmdInst{}

type pluginServer struct {
	pb.UnimplementedPluginServer
}

// Since pb.MiddlewareChannelMessage includes pb.IsEvent_Event, we'll just use that
type listenerCollection struct {
	nonMiddleware []chan pb.MiddlewareChannelMessage
	middleware    []chan pb.MiddlewareChannelMessage
}

var listeners = map[pb.EventType]*listenerCollection{
	pb.EventType_SEND: {
		nonMiddleware: make([]chan pb.MiddlewareChannelMessage, 0, 4),
		middleware:    make([]chan pb.MiddlewareChannelMessage, 0, 4),
	},
}

func (s *pluginServer) RegisterListener(stream pb.Plugin_RegisterListenerServer) error {
	Log.Println("[gRPC] Registering event listener")
	initialData, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	listener := initialData.GetListener()
	if listener == nil {
		return status.Error(codes.InvalidArgument, "First message must be a listener")
	}

	var c *chan pb.MiddlewareChannelMessage
	isMiddleware := listener.Middleware != nil && *listener.Middleware
	isOnce := listener.Once != nil && *listener.Once
	if entry, ok := listeners[listener.Event]; ok {
		var channelCollection *[]chan pb.MiddlewareChannelMessage

		if isMiddleware {
			channelCollection = &entry.middleware
		} else {
			channelCollection = &entry.nonMiddleware
		}

		*channelCollection = append(*channelCollection, make(chan pb.MiddlewareChannelMessage))
		c = &(*channelCollection)[len(*channelCollection)-1]
		defer func() {
			// Remove the channel from the channelCollection where the channel is equal to c
			for i, channel := range *channelCollection {
				if channel == *c {
					*channelCollection = append((*channelCollection)[:i], (*channelCollection)[i+1:]...)
					break
				}
			}
		}()
	}

	for {
		message := <-*c

		// If the message is somehow a *pb.ListenerClientData_Response, it means somehow the last message we sent
		// wasn't consumed, which means the plugin was probably disconnected (at least I think)
		switch message.(type) {
		case *pb.ListenerClientData_Response:
			return status.Error(codes.Unavailable, "Plugin disconnected")
		}

		switch listener.Event {
		case pb.EventType_SEND:
			err := stream.Send(&pb.Event{
				Event: message.(*pb.Event_Send),
			})

			// If something goes wrong, make sure the goroutine sending the message doesn't block on waiting for a response
			sendNilResponse := func() {
				*c <- &pb.ListenerClientData_Response{
					Response: &pb.MiddlewareResponse{
						Res: &pb.MiddlewareResponse_Send{
							Send: &pb.MiddlewareSendResponse{
								Msg: nil,
							},
						},
					},
				}
			}

			if err != nil {
				if isMiddleware {
					sendNilResponse()
				}
				return err
			}
			if isMiddleware {
				mwRes, err := stream.Recv()

				if err != nil {
					sendNilResponse()
					return err
				}

				switch data := mwRes.Data.(type) {
				case *pb.ListenerClientData_Listener:
					sendNilResponse()
					return status.Error(codes.InvalidArgument, "Middleware returned a listener instead of a response")
				case *pb.ListenerClientData_Response:
					*c <- data
				}
			}
		default:
			return status.Error(codes.Unimplemented, "unimplemented")
		}

		if isOnce {
			break
		}
	}

	return nil
}

type cmdInst struct {
	argsInfo string
	info     string
	c        chan *pb.CmdInvocation
}

func (s *pluginServer) RegisterCmd(def *pb.CmdDef, stream pb.Plugin_RegisterCmdServer) error {
	Log.Print("[gRPC] Registering command with name " + def.Name)
	PluginCMDs[def.Name] = cmdInst{
		argsInfo: def.ArgsInfo,
		info:     def.Info,
		c:        make(chan *pb.CmdInvocation),
	}

	defer delete(PluginCMDs, def.Name)

	for {
		invocation := <-PluginCMDs[def.Name].c
		err := stream.Send(invocation)
		if err != nil {
			return err
		}
	}
}

func (s *pluginServer) SendMessage(ctx context.Context, msg *pb.Message) (*pb.MessageRes, error) {
	if msg.GetEphemeralTo() != "" {
		u, success := findUserByName(Rooms[msg.Room], *msg.EphemeralTo)
		if !success {
			return nil, status.Error(codes.NotFound, "Could not find user "+*msg.EphemeralTo)
		}
		u.writeln(msg.GetFrom(), msg.Msg)
	} else {
		r := Rooms[msg.Room]
		if r == nil {
			return nil, status.Error(codes.InvalidArgument, "Room does not exist")
		}
		r.broadcast(msg.GetFrom(), msg.Msg)
	}
	return &pb.MessageRes{}, nil
}

func newPluginServer() *pluginServer {
	s := &pluginServer{}
	return s
}

func authorize(ctx context.Context) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.Unauthenticated, "Missing metadata")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return status.Error(codes.Unauthenticated, "Missing authorization header")
	}

	token := values[0]
	if token != "Bearer "+Integrations.RPC.Key {
		return status.Error(codes.Unauthenticated, "Invalid authorization header")
	}
	return nil
}

func unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	err := authorize(ctx)
	if err != nil {
		return nil, err
	}

	return handler(ctx, req)
}

func streamInterceptor(
	srv interface{},
	stream grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	err := authorize(stream.Context())
	if err != nil {
		return err
	}

	return handler(srv, stream)
}

func rpcInit() {
	if Integrations.RPC == nil {
		return
	}
	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", Integrations.RPC.Port))
		if err != nil {
			fmt.Println("[gRPC] Failed to listen for plugin server:", err)
			return
		}
		// TODO: add TLS if configured
		grpcServer := grpc.NewServer(
			grpc.UnaryInterceptor(unaryInterceptor),
			grpc.StreamInterceptor(streamInterceptor),
			grpc.KeepaliveParams(keepalive.ServerParameters{Time: time.Second * 10}),
		)
		pb.RegisterPluginServer(grpcServer, newPluginServer())
		fmt.Printf("[gRPC] Plugin server started on port %d\n", Integrations.RPC.Port)
		err = grpcServer.Serve(lis)
		if err != nil {
			fmt.Println("[gRPC] Failed to serve:", err)
		}
	}()
}

func runPluginCMDs(u *User, currCmd string, args string) (found bool) {
	if pluginCmd, ok := PluginCMDs[currCmd]; ok {
		pluginCmd.c <- &pb.CmdInvocation{
			Room: u.room.name,
			From: stripansi.Strip(u.Name),
			Args: args,
		}
		return true
	}
	return false
}

func sendMessageToPlugins(line string, u *User) {
	if len(listeners[pb.EventType_SEND].nonMiddleware) > 0 {
		for _, l := range listeners[pb.EventType_SEND].nonMiddleware {
			l <- &pb.Event_Send{
				Send: &pb.SendEvent{
					Room: u.room.name,
					From: stripansi.Strip(u.Name),
					Msg:  line,
				},
			}
		}
	}
}

func getMiddlewareResult(u *User, line string) string {
	if Integrations.RPC == nil {
		return line
	}
	// Middleware hook
	if len(listeners[pb.EventType_SEND].middleware) > 0 {
		for _, m := range listeners[pb.EventType_SEND].middleware {
			m <- &pb.Event_Send{
				Send: &pb.SendEvent{
					Room: u.room.name,
					From: stripansi.Strip(u.Name),
					Msg:  line,
				},
			}
			res := (<-m).(*pb.ListenerClientData_Response).Response.Res.(*pb.MiddlewareResponse_Send).Send
			if res.Msg != nil {
				line = *res.Msg
			}
		}
	}
	return line
}
