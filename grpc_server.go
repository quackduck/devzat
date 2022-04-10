package main

import (
	"context"
	pb "devchat/plugin"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/status"
	"io"
	"net"
	"time"
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
	fmt.Println("Registering listener")
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
		thisIndex := len(*channelCollection) - 1
		defer func() {
			fmt.Println("Cleaning up closed listener")
			*channelCollection = append((*channelCollection)[:thisIndex], (*channelCollection)[thisIndex+1:]...)
		}()
	}

	for {
		message := <-*c

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
				fmt.Println("Error sending message event:", err)
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
	fmt.Println("Registering command")
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
	opts = append(opts, grpc.KeepaliveParams(keepalive.ServerParameters{
		Time: time.Second * 10,
	}))
	// TODO: add TLS if configured
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterPluginServer(grpcServer, newPluginServer())
	l.Printf("Plugin server started on port %d\n", port)
	grpcServer.Serve(lis)
}
