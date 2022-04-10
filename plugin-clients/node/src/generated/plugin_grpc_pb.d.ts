// package: plugin
// file: plugin.proto

/* tslint:disable */
/* eslint-disable */

import * as grpc from "grpc";
import * as plugin_pb from "./plugin_pb";

interface IPluginService extends grpc.ServiceDefinition<grpc.UntypedServiceImplementation> {
    registerListener: IPluginService_IRegisterListener;
    registerCmd: IPluginService_IRegisterCmd;
    sendMessage: IPluginService_ISendMessage;
}

interface IPluginService_IRegisterListener extends grpc.MethodDefinition<plugin_pb.ListenerClientData, plugin_pb.Event> {
    path: "/plugin.Plugin/RegisterListener";
    requestStream: true;
    responseStream: true;
    requestSerialize: grpc.serialize<plugin_pb.ListenerClientData>;
    requestDeserialize: grpc.deserialize<plugin_pb.ListenerClientData>;
    responseSerialize: grpc.serialize<plugin_pb.Event>;
    responseDeserialize: grpc.deserialize<plugin_pb.Event>;
}
interface IPluginService_IRegisterCmd extends grpc.MethodDefinition<plugin_pb.CmdDef, plugin_pb.CmdInvocation> {
    path: "/plugin.Plugin/RegisterCmd";
    requestStream: false;
    responseStream: true;
    requestSerialize: grpc.serialize<plugin_pb.CmdDef>;
    requestDeserialize: grpc.deserialize<plugin_pb.CmdDef>;
    responseSerialize: grpc.serialize<plugin_pb.CmdInvocation>;
    responseDeserialize: grpc.deserialize<plugin_pb.CmdInvocation>;
}
interface IPluginService_ISendMessage extends grpc.MethodDefinition<plugin_pb.Message, plugin_pb.MessageRes> {
    path: "/plugin.Plugin/SendMessage";
    requestStream: false;
    responseStream: false;
    requestSerialize: grpc.serialize<plugin_pb.Message>;
    requestDeserialize: grpc.deserialize<plugin_pb.Message>;
    responseSerialize: grpc.serialize<plugin_pb.MessageRes>;
    responseDeserialize: grpc.deserialize<plugin_pb.MessageRes>;
}

export const PluginService: IPluginService;

export interface IPluginServer {
    registerListener: grpc.handleBidiStreamingCall<plugin_pb.ListenerClientData, plugin_pb.Event>;
    registerCmd: grpc.handleServerStreamingCall<plugin_pb.CmdDef, plugin_pb.CmdInvocation>;
    sendMessage: grpc.handleUnaryCall<plugin_pb.Message, plugin_pb.MessageRes>;
}

export interface IPluginClient {
    registerListener(): grpc.ClientDuplexStream<plugin_pb.ListenerClientData, plugin_pb.Event>;
    registerListener(options: Partial<grpc.CallOptions>): grpc.ClientDuplexStream<plugin_pb.ListenerClientData, plugin_pb.Event>;
    registerListener(metadata: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientDuplexStream<plugin_pb.ListenerClientData, plugin_pb.Event>;
    registerCmd(request: plugin_pb.CmdDef, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<plugin_pb.CmdInvocation>;
    registerCmd(request: plugin_pb.CmdDef, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<plugin_pb.CmdInvocation>;
    sendMessage(request: plugin_pb.Message, callback: (error: grpc.ServiceError | null, response: plugin_pb.MessageRes) => void): grpc.ClientUnaryCall;
    sendMessage(request: plugin_pb.Message, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: plugin_pb.MessageRes) => void): grpc.ClientUnaryCall;
    sendMessage(request: plugin_pb.Message, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: plugin_pb.MessageRes) => void): grpc.ClientUnaryCall;
}

export class PluginClient extends grpc.Client implements IPluginClient {
    constructor(address: string, credentials: grpc.ChannelCredentials, options?: object);
    public registerListener(options?: Partial<grpc.CallOptions>): grpc.ClientDuplexStream<plugin_pb.ListenerClientData, plugin_pb.Event>;
    public registerListener(metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientDuplexStream<plugin_pb.ListenerClientData, plugin_pb.Event>;
    public registerCmd(request: plugin_pb.CmdDef, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<plugin_pb.CmdInvocation>;
    public registerCmd(request: plugin_pb.CmdDef, metadata?: grpc.Metadata, options?: Partial<grpc.CallOptions>): grpc.ClientReadableStream<plugin_pb.CmdInvocation>;
    public sendMessage(request: plugin_pb.Message, callback: (error: grpc.ServiceError | null, response: plugin_pb.MessageRes) => void): grpc.ClientUnaryCall;
    public sendMessage(request: plugin_pb.Message, metadata: grpc.Metadata, callback: (error: grpc.ServiceError | null, response: plugin_pb.MessageRes) => void): grpc.ClientUnaryCall;
    public sendMessage(request: plugin_pb.Message, metadata: grpc.Metadata, options: Partial<grpc.CallOptions>, callback: (error: grpc.ServiceError | null, response: plugin_pb.MessageRes) => void): grpc.ClientUnaryCall;
}
