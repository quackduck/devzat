package main

import (
	"context"
	pb "devchat/plugin"
	"fmt"
	"io"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

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
		return status.Errorf(codes.InvalidArgument, "First message must be a listener")
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
			return status.Errorf(codes.Unavailable, "Plugin disconnected")
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
					return status.Errorf(codes.InvalidArgument, "Middleware returned a listener instead of a response")
				case *pb.ListenerClientData_Response:
					*c <- data
				}
			}
		default:
			return status.Errorf(codes.Unimplemented, "unimplemented")
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

var pluginCmds = map[string]cmdInst{}

func (s *pluginServer) RegisterCmd(def *pb.CmdDef, stream pb.Plugin_RegisterCmdServer) error {
	Log.Printf("[gRPC] Registering command with name %s", def.Name)
	pluginCmds[def.Name] = cmdInst{
		argsInfo: def.ArgsInfo,
		info:     def.Info,
		c:        make(chan *pb.CmdInvocation),
	}

	defer delete(pluginCmds, def.Name)

	for {
		invocation := <-pluginCmds[def.Name].c
		err := stream.Send(invocation)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *pluginServer) SendMessage(ctx context.Context, msg *pb.Message) (*pb.MessageRes, error) {
	if msg.GetEphemeralTo() != "" {
		u, success := findUserByName(Rooms[msg.Room], *msg.EphemeralTo)
		if !success {
			return nil, status.Errorf(codes.NotFound, "Could not find user %s", *msg.EphemeralTo)
		}
		u.writeln(msg.GetFrom(), msg.Msg)
	} else {
		r := Rooms[msg.Room]
		if r == nil {
			return nil, status.Errorf(codes.InvalidArgument, "Room does not exist")
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
		return status.Errorf(codes.Unauthenticated, "Missing metadata")
	}

	values := md["authorization"]
	if len(values) == 0 {
		return status.Errorf(codes.Unauthenticated, "Missing authorization header")
	}

	token := values[0]
	if token != "Bearer "+Config.PluginToken {
		return status.Errorf(codes.Unauthenticated, "Invalid authorization header")
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

func startPluginServer(port int) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		Log.Fatalf("[gRPC] Failed to listen for plugin server: %v", err)
	}
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(unaryInterceptor),
		grpc.StreamInterceptor(streamInterceptor),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			Time: time.Second * 10,
		}),
	}
	// TODO: add TLS if configured
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterPluginServer(grpcServer, newPluginServer())
	Log.Printf("[gRPC] Plugin server started on port %d\n", port)
	grpcServer.Serve(lis)
}
