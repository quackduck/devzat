import * as grpc from "@grpc/grpc-js";
import protoLoader from "@grpc/proto-loader";
//@ts-ignore - need for dts-bundle-generator
import fdB64String from "./generated/plugin-desc.pb";
import {ServiceClient} from "@grpc/grpc-js/build/src/make-client";
import type { IPluginServer } from "./generated/plugin_grpc_pb";
import type * as plugin_pb from "./generated/plugin_pb";

const packageDef = protoLoader.loadFileDescriptorSetFromBuffer(Buffer.from(fdB64String, "base64"), {
  keepCase: false,
  longs: String,
  enums: String,
  defaults: true,
  oneofs: true
});
const protoDescriptor = grpc.loadPackageDefinition(packageDef);
const plugin = protoDescriptor.plugin as grpc.GrpcObject;

type Stub = ServiceClient & {
  [K in keyof IPluginServer]: Function
}

// GRPC object types
// These are exposed to the user so they should have nice names

export type Message = plugin_pb.Message.AsObject;
export type MessageRes = plugin_pb.MessageRes.AsObject;

export type Listener = Omit<plugin_pb.Listener.AsObject, "event">;
export type SendEvent = plugin_pb.SendEvent.AsObject;

export type CmdDef = Omit<plugin_pb.CmdDef.AsObject, "argsinfo"> & { argsInfo: plugin_pb.CmdDef.AsObject["argsinfo"] };
export type CmdInvocation = plugin_pb.CmdInvocation.AsObject;

export default class Plugin {
  stub: Stub
  name?: string

  constructor({ address, creds = grpc.credentials.createInsecure(), name }: {
    address: string,
    creds?: grpc.ChannelCredentials,
    name?: string
  }) {
    this.stub = new (plugin.Plugin as grpc.ServiceClientConstructor)(address, creds) as Stub;
    this.name = name;
  }

  sendMessage(message: Message & {
    from?: string | null | undefined
  }): Promise<MessageRes> {
    return new Promise((resolve, reject) => {
      this.stub.sendMessage({
        ...message,
        from: message.from === undefined ? this.name : message.from === null ? undefined : message.from
      }, (err: grpc.ServiceError | null, res: MessageRes) => {
        if (err) {
          reject(err);
        } else {
          resolve(res);
        }
      });
    });
  }

  onMessageSend(listener: Listener, callback: (event: SendEvent) => string | void | Promise<string> | Promise<void>): () => void {
    let callCancel: () => void;

    let attemptingReconnect = false;
    const attemptReconnect = () => {
      if(attemptingReconnect) return;
      attemptingReconnect = true;
      console.log("Command invocation listener stream ended, reconnecting");
      setTimeout(() => {
        connect();
        attemptingReconnect = false;
      }, 1000);
    };

    const connect = () => {
      const call: grpc.ClientDuplexStream<plugin_pb.ListenerClientData.AsObject, plugin_pb.Event.AsObject> = this.stub.registerListener();

      call.on("data", async (e: plugin_pb.Event.AsObject) => {
        if(!e.send) throw new Error("Event listener gRPC stream received incorrect event type");
        const res = await callback(e.send);
        if(!listener.middleware) {
          if(res === undefined) return; else throw new Error("Event callback returned a value but the listener wasn't marked as middleware");
        }
        call.write({
          response: {
            send: {
              msg: res as string | undefined
            }
          }
        });
      });

      call.on("error", e => {
        if((e as any).code === grpc.status.UNAVAILABLE || (e as any).code === grpc.status.CANCELLED) {
          // The client got disconnected from the server
          attemptReconnect();
        } else throw e;
      });

      call.on("end", () => {
        if(!listener.once) {
          // the stream was supposed to still be open, so reconnect
          attemptReconnect()
        }
      })

      call.write({
        listener: {
          ...listener,
          event: 0 /* EventType.SEND */
        }
      });
    };

    connect();

    return () => callCancel();
  }

  command(command: CmdDef, callback: (event: CmdInvocation) => string | undefined | Message | Promise<string> | Promise<undefined> | Promise<Message>): () => void {
    let callCancel: () => void;

    let attemptingReconnect = false;
    const attemptReconnect = () => {
      if(attemptingReconnect) return;
      attemptingReconnect = true;
      console.log("Command invocation listener stream ended, reconnecting");
      setTimeout(() => {
        connect();
        attemptingReconnect = false;
      }, 1000);
    };

    const connect = () => {
      const call: grpc.ClientReadableStream<plugin_pb.CmdInvocation.AsObject> = this.stub.registerCmd(command);
      callCancel = call.cancel;

      call.on("data", async (e: plugin_pb.CmdInvocation.AsObject) => {
        const res = await callback(e);

        if(res !== undefined) {
          // For convenience, reply to the command invocation with the callback's return value
          await this.sendMessage(typeof res === "string" ? {
            msg: res,
            room: e.room
          } : res);
        }
      });

      call.on("error", e => {
        if((e as any).code === grpc.status.UNAVAILABLE || (e as any).code === grpc.status.CANCELLED) {
          // The client got disconnected from the server
          attemptReconnect();
        } else throw e;
      });

      call.on("end", () => {
        // the stream was supposed to still be open, so reconnect
        attemptReconnect();
      });
    };

    connect();

    return () => callCancel();
  }
}