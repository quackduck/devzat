package main

import (
	"context"
	pb "devchat/plugin"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
)

type pluginServer struct {
	pb.UnimplementedPluginServer
}

type middlewareCollection struct {
	channels []chan pb.IsEvent_Event
	res      chan pb.MiddlewareRes
}

type listenerCollection struct {
	nonMiddleware []chan pb.IsEvent_Event
	middleware    middlewareCollection
}

var listeners = map[pb.EventType]listenerCollection{
	pb.EventType_SEND: {
		nonMiddleware: make([]chan pb.IsEvent_Event, 0, 4),
		middleware: middlewareCollection{
			channels: make([]chan pb.IsEvent_Event, 0, 4),
			res:      make(chan pb.MiddlewareRes),
		},
	},
}

func (s *pluginServer) RegisterListener(listener *pb.Listener, stream pb.Plugin_RegisterListenerServer) error {
	var thisCollection []chan pb.IsEvent_Event
	var thisIndex int
	if entry, ok := listeners[listener.Event]; ok {
		if listener.Middleware != nil && *listener.Middleware {
			thisCollection = entry.middleware.channels
			entry.middleware.channels = append(thisCollection, make(chan pb.IsEvent_Event))
			defer func() {
				entry.middleware.channels = append(entry.middleware.channels[:thisIndex], entry.middleware.channels[thisIndex+1:]...)
			}()
		} else {
			thisCollection = entry.nonMiddleware
			entry.nonMiddleware = append(thisCollection, make(chan pb.IsEvent_Event))
			defer func() {
				entry.nonMiddleware = append(entry.nonMiddleware[:thisIndex], entry.nonMiddleware[thisIndex+1:]...)
			}()
		}
	}
	thisIndex = len(thisCollection) - 1

	for {
		message := <-thisCollection[thisIndex]

		switch listener.Event {
		case pb.EventType_SEND:
			err := stream.Send(&pb.Event{
				Event: message,
			})
			if err != nil {
				return err
			}
		default:
			return status.Errorf(codes.Unimplemented, "unimplemented")
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
	if msg.EphemeralTo != nil {
		// TODO send an ephemeral message
		return nil, status.Errorf(codes.Unimplemented, "unimplemented")
	}
	r := rooms[msg.Room]
	if r == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Room does not exist")
	}
	r.broadcast(msg.GetFrom(), msg.Msg)
	return &pb.MessageRes{}, nil
}

func (s *pluginServer) MiddlewareEditMessage(ctx context.Context, msg *pb.MiddlewareMessage) (*pb.MiddlewareEditMessageRes, error) {
	// TODO: this should somehow be a response (two-way stream?) from the plugin, not a separate method (because this lets another plugin edit the message)

	listeners[pb.EventType_SEND].middleware.res <- msg

	return &pb.MiddlewareEditMessageRes{}, nil
}

func (s *pluginServer) MiddlewareDone(ctx context.Context, msg *pb.MiddlewareDoneMessage) (*pb.MiddlewareDoneRes, error) {
	listeners[pb.EventType_SEND].middleware.res <- msg

	return &pb.MiddlewareDoneRes{}, nil
}

func newPluginServer() *pluginServer {
	s := &pluginServer{}
	return s
}

func startPluginServer(port uint32) {
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		l.Fatalf("Failed to listen for plugin server: %v", err)
	}
	var opts []grpc.ServerOption
	// TODO: add TLS if configured
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterPluginServer(grpcServer, newPluginServer())
	l.Printf("Plugin server started on port %d\n", port)
	grpcServer.Serve(lis)
}
