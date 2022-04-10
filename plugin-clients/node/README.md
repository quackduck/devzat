# Devzat plugin API client for Node.js

This NPM package allows you to build Devzat plugins/bots with JavaScript/TypeScript. See [example/index.ts](example/index.ts) for a full example of a bot made with this package.

## Getting started

Install the package:

```shell
yarn add devzat
```

Create an instance of the client:

```ts
import Devzat from "devzat";

const plugin = new Devzat({
    address: "localhost:5556", // The address to the Devzat server's plugin API (different than the SSH port)
    name: "Demo bot" // Name of your bot (can be overridden later)
});
```

From there you can use various methods to send, receive, and intercept messages.

## `Devzat.sendMessage(message: Message): Promise<{}>`

Send a message in a given room.

```ts
type Message = {
    room: string, // The room name (including the `#`) to send to
    from?: string | null, // The name of the user sending the message (defaults to the bot's name),
                          // can be set to null to not have any name attached to the message
    msg: string, // Message text (in Markdown)
    ephemeralTo?: string // TODO not implemented
}
```

## `Devzat.onMessageSend(listener: Listener, callback: (e: SendEvent) => string | void | Promise<string> | Promise<void>): () => void`

Register an event listener to fire on every message send. Returns a function to remove the listener.

```ts
type Listener = {
    middleware?: boolean, // If true, the listener can edit the user's message before it is sent
    once?: boolean // Remove the listener after the first event
}

type SendEvent = {
    room: string,
    from: string,
    msg: string
}
```

## `Devzat.command(command: CmdDef, callback: (e: CmdInvocation) => string | void | Promise<string> | Promise<void>): () => void`

Register a command to be handled by your bot. Returning a value from the callback will reply to the message.

```ts
type CmdDef = {
    name: string, // Name of the command (triggers when a message starts with it)
    argsInfo: string, // Summary of arguments (like <name>)
    info: string // Description of the command
}

type CmdInvocation = {
    room: string, // Room the command was issued in
    from: string, // User who issued the command
    args: string // Everything after the command name
}
```