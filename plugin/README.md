# Plugin API docs

__This documentation is for the raw plugin API. If you want to write a plugin in Node.js, use [the Node client package](https://yarn.pm/devzat)__

The Devzat plugin API allows you to build bots and other tools that integrate with Devzat, similar to a Discord bot or a Slack app. The API uses gRPC for communication. If you're already familiar with gRPC, you can jump right in with [the `.proto` file](./plugin.proto), but if you're not, this document will explain how to get started.

Click [here](#3-using-the-api) to skip setup.

## 0. Setting up the plugin API

In order to use the gRPC plugin API, you need to [enable the integration](../Admin's%20Manual.md#using-the-plugin-api-integration) in Devzat's config file. Get an authentication token for your client by having an admin run `grant [username] [description]` in your Devzat instance. If the token needs to be revoked in the future, use `lstokens` to get the hash of the token, then have an admin run `revoke <hash>`.

## 1. Setting up a gRPC client

First, you'll need to set up a gRPC client by following [the gRPC docs for your language](https://grpc.io/docs/languages/). You can copy the [`plugin.proto` file](./plugin.proto) directly into your project, or if your language's package manager supports it, add the Devzat GitHub repository as a dependency to your project and link directly to the `plugin.proto` file (for example, for Node.js, run `yarn add https://github.com/quackduck/devzat` to get `node_modules/devzat/plugin/plugin.proto`).

## 2. Connecting to Devzat's gRPC server

Configure your gRPC client to connect to Devzat's gRPC server on the port specified in the integration config file. Use `insecure` credentials, and configure the client to send an `Authorization` header with the contents `Bearer <auth token>`. How this is done depends on the language: see [here](https://grpc.io/docs/guides/auth/#extending-grpc-to-support-other-authentication-mechanisms) for an example of how to do this in C++.

## 3. Using the API

Here's a summary of all the methods the gRPC API provides. All methods are under the `Plugin` service.

### `RegisterListener`

The `RegisterListener` method is used to register an event listener or middleware. It accepts a stream of messages of type `ListenerClientData` and returns a stream of `Event`s. A `ListenerClientData` can be either a `Listener` or a `MiddlewareResponse` message. 

When you first establish the connection, send a `Listener` to set it up. In that message, you can set whether the listener is middleware (allowing you to intercept and edit messages before they are sent), whether the listener should only fire `once`, and optionally provide a regex, allowing you to control when the event fires (useful for reducing latency when building middleware). Devzat will send an `Event` containing details of the event when it occurs. If you registered a middleware listener, you should send back a `MiddlewareResponse` to allow Devzat to continue processing the event.

This is the most complicated part of the gRPC service, so you can reference the [Node.js implementation](https://github.com/Merlin04/devzat-node/blob/main/src/index.ts#L99) as an example of its usage.

Signature:
```protobuf
rpc RegisterListener(stream ListenerClientData) returns (stream Event);
```

Relevant message types:
```protobuf
message Listener {
  optional bool middleware = 1;
  optional bool once = 2;
  // Regex to match against to determine if this listener should be called
  // Does not include slashes or flags
  optional string regex = 3;
}

message MiddlewareResponse {
  optional string msg = 1;
}

message Event {
  string room = 1;
  string from = 2;
  string msg = 3;
}
```

### `RegisterCmd`

The `RegisterCmd` method is used to register a command with Devzat, which will then show up when a user runs `plugins`. The server will send a `CmdInvocation` whenever your command is invoked, allowing you to perform some action such as responding to the user.

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

The `SendMessage` method is used to send a message. You must specify the room to send it to and the message to send. You can optionally include the name to send it as and a user to send the message _ephemerally_ to (allowing only that user to see it: aka DMs).

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
