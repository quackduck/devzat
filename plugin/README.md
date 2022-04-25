# Plugin API docs

__This documentation is for the raw plugin API. If you want to write a plugin in Node.js, use [the Node client package](https://yarn.pm/devzat)__

The Devzat plugin API allows you to build bots and other tools that integrate with Devzat, similar to a Discord bot or a Slack app. The API uses gRPC for communication; if you're already familiar with gRPC, you can jump right in with [the `.proto` file](./plugin.proto), but if you're not, this document will explain how to get started.

Click [here](#3-use-the-api) to skip setup and jump to the API docs.

## 0. Enable the plugin API

In order to use the gRPC plugin API, you need to enable the integration in Devzat's config file. [See here for instructions on how to do that.](../Admin's%20Manual.md#using-the-plugin-api-integration)

## 1. Set up a gRPC client

First, you'll need to set up a gRPC client by following [the instructions in the gRPC docs for your language](https://grpc.io/docs/languages/). You can copy the [`plugin.proto` file](./plugin.proto) directly into your project, or if your language's package manager supports it, add the Devzat GitHub repository as a dependency to your project and link directly to the `plugin.proto` file (for example, for Node.js, run `yarn add https://github.com/quackduck/devzat` to obtain `node_modules/devzat/plugin/pluin.proto`).

## 2. Connect to Devzat's gRPC server

Configure your gRPC client to connect to Devzat's gRPC server, running on the port specified in the integration config file. Use `insecure` credentials, and configure the client to send an `Authorization` header with the contents `Bearer <token>`, where token is the token you configured in the integration config. The mechanism to achieve this varies by language; see [here](https://grpc.io/docs/guides/auth/#extending-grpc-to-support-other-authentication-mechanisms) for an example of how to do this in C++.

## 3. Use the API

Here's a summary of all the methods the gRPC API provides. All methods are under the `Plugin` service.

### `RegisterListener`

The `RegisterListener` method is used to register an event listener or middleware. It accepts a stream of messages of type `ListenerClientData` and returns a stream of `Event`s. A `ListenerClientData` can be either a `Listener` or a `MiddlewareResponse` message. 

When you first establish the connection, send a `Listener` to set it up. In that message, you can set the event type (currently only supports send events), whether the listener is middleware (allowing you to intercept and edit messages before they are sent), and whether the listener should only fire `once`. Devzat will send an `Event` containing details of the event when it occurs; if you registered a middleware listener, you have to send back a `MiddlewareResponse` to allow Devzat to continue processing the event.

This is by far the most complicated part of the gRPC service; you can reference the [Node.js implementation](https://github.com/Merlin04/devzat-node/blob/main/src/index.ts#L80) as an example.

Signature:
```protobuf
rpc RegisterListener(stream ListenerClientData) returns (stream Event);
```

Relevant message types:
```protobuf
message ListenerClientData {
  oneof data {
    Listener listener = 1;
    MiddlewareResponse response = 2;
  }
}

message Listener {
  EventType event = 1;
  optional bool middleware = 2;
  optional bool once = 3;
}

message MiddlewareResponse {
  oneof res {
    MiddlewareSendResponse send = 1;
  }
}

message MiddlewareSendResponse {
  optional string msg = 1;
}

message Event {
  oneof event {
    SendEvent send = 1;
  }
}

enum EventType {
  SEND = 0;
}

message SendEvent {
  string room = 1;
  string from = 2;
  string msg = 3;
}
```

### `RegisterCmd`

The `RegisterCmd` method is used to register a command with Devzat. This will allow it to show up when `cmds` is run by a user. The server will send a `CmdInvocation` whenever your command is invoked, allowing you to perform some action such as responding to the user.

Signature:
```protobuf
rpc RegisterCmd(CmdDef) returns (stream CmdInvocation);
```

Relevant types:
```protobuf
message CmdDef {
  string name = 1;
  string argsInfo = 2;
  string info = 3;
}

message CmdInvocation {
  string room = 1;
  string from = 2;
  string args = 3;
}
```

### `SendMessage`

The `SendMessage` method is used to send a message. You can specify the room to send it to, the message to send, and optionally the name to send it from and a user to send the message _ephemerally_ to (allowing only them to see it).

Signature:
```protobuf
rpc SendMessage(Message) returns (MessageRes);
```

Relevant types:
```protobuf
message Message {
  string room = 1;
  optional string from = 2;
  string msg = 3;
  optional string ephemeral_to = 4;
}

message MessageRes {}
```
