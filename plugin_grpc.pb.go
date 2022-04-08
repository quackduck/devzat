// Code generated by protoc-gen-go-grpc. DO NOT EDIT.
// versions:
// - protoc-gen-go-grpc v1.2.0
// - protoc             v3.15.8
// source: plugin.proto

package main

import (
	context "context"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
// Requires gRPC-Go v1.32.0 or later.
const _ = grpc.SupportPackageIsVersion7

// PluginClient is the client API for Plugin service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://pkg.go.dev/google.golang.org/grpc/?tab=doc#ClientConn.NewStream.
type PluginClient interface {
	// Events are implemented through a stream that is held open
	RegisterListener(ctx context.Context, in *Listener, opts ...grpc.CallOption) (Plugin_RegisterListenerClient, error)
	RegisterCmd(ctx context.Context, in *CmdDef, opts ...grpc.CallOption) (Plugin_RegisterCmdClient, error)
	// Commands a plugin can call
	SendMessage(ctx context.Context, in *Message, opts ...grpc.CallOption) (*MessageRes, error)
	SendImage(ctx context.Context, in *Image, opts ...grpc.CallOption) (*ImageRes, error)
	MiddlewareEditMessage(ctx context.Context, in *MiddlewareMessage, opts ...grpc.CallOption) (*MiddlewareEditMessageRes, error)
	// Have to name it that to avoid colliding with the RPC method name
	MiddlewareDone(ctx context.Context, in *MiddlewareDoneMessage, opts ...grpc.CallOption) (*MiddlewareDoneRes, error)
}

type pluginClient struct {
	cc grpc.ClientConnInterface
}

func NewPluginClient(cc grpc.ClientConnInterface) PluginClient {
	return &pluginClient{cc}
}

func (c *pluginClient) RegisterListener(ctx context.Context, in *Listener, opts ...grpc.CallOption) (Plugin_RegisterListenerClient, error) {
	stream, err := c.cc.NewStream(ctx, &Plugin_ServiceDesc.Streams[0], "/plugin.Plugin/RegisterListener", opts...)
	if err != nil {
		return nil, err
	}
	x := &pluginRegisterListenerClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Plugin_RegisterListenerClient interface {
	Recv() (*Event, error)
	grpc.ClientStream
}

type pluginRegisterListenerClient struct {
	grpc.ClientStream
}

func (x *pluginRegisterListenerClient) Recv() (*Event, error) {
	m := new(Event)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *pluginClient) RegisterCmd(ctx context.Context, in *CmdDef, opts ...grpc.CallOption) (Plugin_RegisterCmdClient, error) {
	stream, err := c.cc.NewStream(ctx, &Plugin_ServiceDesc.Streams[1], "/plugin.Plugin/RegisterCmd", opts...)
	if err != nil {
		return nil, err
	}
	x := &pluginRegisterCmdClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Plugin_RegisterCmdClient interface {
	Recv() (*CmdInvokation, error)
	grpc.ClientStream
}

type pluginRegisterCmdClient struct {
	grpc.ClientStream
}

func (x *pluginRegisterCmdClient) Recv() (*CmdInvokation, error) {
	m := new(CmdInvokation)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *pluginClient) SendMessage(ctx context.Context, in *Message, opts ...grpc.CallOption) (*MessageRes, error) {
	out := new(MessageRes)
	err := c.cc.Invoke(ctx, "/plugin.Plugin/SendMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginClient) SendImage(ctx context.Context, in *Image, opts ...grpc.CallOption) (*ImageRes, error) {
	out := new(ImageRes)
	err := c.cc.Invoke(ctx, "/plugin.Plugin/SendImage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginClient) MiddlewareEditMessage(ctx context.Context, in *MiddlewareMessage, opts ...grpc.CallOption) (*MiddlewareEditMessageRes, error) {
	out := new(MiddlewareEditMessageRes)
	err := c.cc.Invoke(ctx, "/plugin.Plugin/MiddlewareEditMessage", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *pluginClient) MiddlewareDone(ctx context.Context, in *MiddlewareDoneMessage, opts ...grpc.CallOption) (*MiddlewareDoneRes, error) {
	out := new(MiddlewareDoneRes)
	err := c.cc.Invoke(ctx, "/plugin.Plugin/MiddlewareDone", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// PluginServer is the server API for Plugin service.
// All implementations must embed UnimplementedPluginServer
// for forward compatibility
type PluginServer interface {
	// Events are implemented through a stream that is held open
	RegisterListener(*Listener, Plugin_RegisterListenerServer) error
	RegisterCmd(*CmdDef, Plugin_RegisterCmdServer) error
	// Commands a plugin can call
	SendMessage(context.Context, *Message) (*MessageRes, error)
	SendImage(context.Context, *Image) (*ImageRes, error)
	MiddlewareEditMessage(context.Context, *MiddlewareMessage) (*MiddlewareEditMessageRes, error)
	// Have to name it that to avoid colliding with the RPC method name
	MiddlewareDone(context.Context, *MiddlewareDoneMessage) (*MiddlewareDoneRes, error)
	mustEmbedUnimplementedPluginServer()
}

// UnimplementedPluginServer must be embedded to have forward compatible implementations.
type UnimplementedPluginServer struct {
}

func (UnimplementedPluginServer) RegisterListener(*Listener, Plugin_RegisterListenerServer) error {
	return status.Errorf(codes.Unimplemented, "method RegisterListener not implemented")
}
func (UnimplementedPluginServer) RegisterCmd(*CmdDef, Plugin_RegisterCmdServer) error {
	return status.Errorf(codes.Unimplemented, "method RegisterCmd not implemented")
}
func (UnimplementedPluginServer) SendMessage(context.Context, *Message) (*MessageRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendMessage not implemented")
}
func (UnimplementedPluginServer) SendImage(context.Context, *Image) (*ImageRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method SendImage not implemented")
}
func (UnimplementedPluginServer) MiddlewareEditMessage(context.Context, *MiddlewareMessage) (*MiddlewareEditMessageRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MiddlewareEditMessage not implemented")
}
func (UnimplementedPluginServer) MiddlewareDone(context.Context, *MiddlewareDoneMessage) (*MiddlewareDoneRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "method MiddlewareDone not implemented")
}
func (UnimplementedPluginServer) mustEmbedUnimplementedPluginServer() {}

// UnsafePluginServer may be embedded to opt out of forward compatibility for this service.
// Use of this interface is not recommended, as added methods to PluginServer will
// result in compilation errors.
type UnsafePluginServer interface {
	mustEmbedUnimplementedPluginServer()
}

func RegisterPluginServer(s grpc.ServiceRegistrar, srv PluginServer) {
	s.RegisterService(&Plugin_ServiceDesc, srv)
}

func _Plugin_RegisterListener_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Listener)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PluginServer).RegisterListener(m, &pluginRegisterListenerServer{stream})
}

type Plugin_RegisterListenerServer interface {
	Send(*Event) error
	grpc.ServerStream
}

type pluginRegisterListenerServer struct {
	grpc.ServerStream
}

func (x *pluginRegisterListenerServer) Send(m *Event) error {
	return x.ServerStream.SendMsg(m)
}

func _Plugin_RegisterCmd_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(CmdDef)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(PluginServer).RegisterCmd(m, &pluginRegisterCmdServer{stream})
}

type Plugin_RegisterCmdServer interface {
	Send(*CmdInvokation) error
	grpc.ServerStream
}

type pluginRegisterCmdServer struct {
	grpc.ServerStream
}

func (x *pluginRegisterCmdServer) Send(m *CmdInvokation) error {
	return x.ServerStream.SendMsg(m)
}

func _Plugin_SendMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Message)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).SendMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.Plugin/SendMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).SendMessage(ctx, req.(*Message))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugin_SendImage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Image)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).SendImage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.Plugin/SendImage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).SendImage(ctx, req.(*Image))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugin_MiddlewareEditMessage_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MiddlewareMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).MiddlewareEditMessage(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.Plugin/MiddlewareEditMessage",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).MiddlewareEditMessage(ctx, req.(*MiddlewareMessage))
	}
	return interceptor(ctx, in, info, handler)
}

func _Plugin_MiddlewareDone_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(MiddlewareDoneMessage)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(PluginServer).MiddlewareDone(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/plugin.Plugin/MiddlewareDone",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(PluginServer).MiddlewareDone(ctx, req.(*MiddlewareDoneMessage))
	}
	return interceptor(ctx, in, info, handler)
}

// Plugin_ServiceDesc is the grpc.ServiceDesc for Plugin service.
// It's only intended for direct use with grpc.RegisterService,
// and not to be introspected or modified (even as a copy)
var Plugin_ServiceDesc = grpc.ServiceDesc{
	ServiceName: "plugin.Plugin",
	HandlerType: (*PluginServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "SendMessage",
			Handler:    _Plugin_SendMessage_Handler,
		},
		{
			MethodName: "SendImage",
			Handler:    _Plugin_SendImage_Handler,
		},
		{
			MethodName: "MiddlewareEditMessage",
			Handler:    _Plugin_MiddlewareEditMessage_Handler,
		},
		{
			MethodName: "MiddlewareDone",
			Handler:    _Plugin_MiddlewareDone_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "RegisterListener",
			Handler:       _Plugin_RegisterListener_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "RegisterCmd",
			Handler:       _Plugin_RegisterCmd_Handler,
			ServerStreams: true,
		},
	},
	Metadata: "plugin.proto",
}
