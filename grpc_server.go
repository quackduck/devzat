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

func (s *pluginServer) RegisterListener(listener *pb.Listener, stream pb.Plugin_RegisterListenerServer) error {
	return nil
}

func (s *pluginServer) RegisterCmd(def *pb.CmdDef, stream pb.Plugin_RegisterCmdServer) error {
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
	if msg.From == nil {
		f := ""
		msg.From = &f
	}
	r.broadcast(*msg.From, msg.Msg)
	return &pb.MessageRes{}, nil
}

func (s *pluginServer) SendImage(ctx context.Context, image *pb.Image) (*pb.ImageRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *pluginServer) MiddlewareEditMessage(ctx context.Context, msg *pb.MiddlewareMessage) (*pb.MiddlewareEditMessageRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (s *pluginServer) MiddlewareDone(ctx context.Context, msg *pb.MiddlewareDoneMessage) (*pb.MiddlewareDoneRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
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
