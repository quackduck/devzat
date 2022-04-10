// package: plugin
// file: plugin.proto

/* tslint:disable */
/* eslint-disable */

import * as jspb from "google-protobuf";

export class SendEvent extends jspb.Message { 
    getRoom(): string;
    setRoom(value: string): SendEvent;
    getFrom(): string;
    setFrom(value: string): SendEvent;
    getMsg(): string;
    setMsg(value: string): SendEvent;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): SendEvent.AsObject;
    static toObject(includeInstance: boolean, msg: SendEvent): SendEvent.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: SendEvent, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): SendEvent;
    static deserializeBinaryFromReader(message: SendEvent, reader: jspb.BinaryReader): SendEvent;
}

export namespace SendEvent {
    export type AsObject = {
        room: string,
        from: string,
        msg: string,
    }
}

export class ListenerClientData extends jspb.Message { 

    hasListener(): boolean;
    clearListener(): void;
    getListener(): Listener | undefined;
    setListener(value?: Listener): ListenerClientData;

    hasResponse(): boolean;
    clearResponse(): void;
    getResponse(): MiddlewareResponse | undefined;
    setResponse(value?: MiddlewareResponse): ListenerClientData;

    getDataCase(): ListenerClientData.DataCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): ListenerClientData.AsObject;
    static toObject(includeInstance: boolean, msg: ListenerClientData): ListenerClientData.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: ListenerClientData, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): ListenerClientData;
    static deserializeBinaryFromReader(message: ListenerClientData, reader: jspb.BinaryReader): ListenerClientData;
}

export namespace ListenerClientData {
    export type AsObject = {
        listener?: Listener.AsObject,
        response?: MiddlewareResponse.AsObject,
    }

    export enum DataCase {
        DATA_NOT_SET = 0,
        LISTENER = 1,
        RESPONSE = 2,
    }

}

export class Listener extends jspb.Message { 
    getEvent(): EventType;
    setEvent(value: EventType): Listener;

    hasMiddleware(): boolean;
    clearMiddleware(): void;
    getMiddleware(): boolean | undefined;
    setMiddleware(value: boolean): Listener;

    hasOnce(): boolean;
    clearOnce(): void;
    getOnce(): boolean | undefined;
    setOnce(value: boolean): Listener;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Listener.AsObject;
    static toObject(includeInstance: boolean, msg: Listener): Listener.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Listener, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Listener;
    static deserializeBinaryFromReader(message: Listener, reader: jspb.BinaryReader): Listener;
}

export namespace Listener {
    export type AsObject = {
        event: EventType,
        middleware?: boolean,
        once?: boolean,
    }
}

export class Event extends jspb.Message { 

    hasSend(): boolean;
    clearSend(): void;
    getSend(): SendEvent | undefined;
    setSend(value?: SendEvent): Event;

    getEventCase(): Event.EventCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Event.AsObject;
    static toObject(includeInstance: boolean, msg: Event): Event.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Event, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Event;
    static deserializeBinaryFromReader(message: Event, reader: jspb.BinaryReader): Event;
}

export namespace Event {
    export type AsObject = {
        send?: SendEvent.AsObject,
    }

    export enum EventCase {
        EVENT_NOT_SET = 0,
        SEND = 1,
    }

}

export class MiddlewareResponse extends jspb.Message { 

    hasSend(): boolean;
    clearSend(): void;
    getSend(): MiddlewareSendResponse | undefined;
    setSend(value?: MiddlewareSendResponse): MiddlewareResponse;

    getResCase(): MiddlewareResponse.ResCase;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): MiddlewareResponse.AsObject;
    static toObject(includeInstance: boolean, msg: MiddlewareResponse): MiddlewareResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: MiddlewareResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): MiddlewareResponse;
    static deserializeBinaryFromReader(message: MiddlewareResponse, reader: jspb.BinaryReader): MiddlewareResponse;
}

export namespace MiddlewareResponse {
    export type AsObject = {
        send?: MiddlewareSendResponse.AsObject,
    }

    export enum ResCase {
        RES_NOT_SET = 0,
        SEND = 1,
    }

}

export class MiddlewareSendResponse extends jspb.Message { 

    hasMsg(): boolean;
    clearMsg(): void;
    getMsg(): string | undefined;
    setMsg(value: string): MiddlewareSendResponse;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): MiddlewareSendResponse.AsObject;
    static toObject(includeInstance: boolean, msg: MiddlewareSendResponse): MiddlewareSendResponse.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: MiddlewareSendResponse, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): MiddlewareSendResponse;
    static deserializeBinaryFromReader(message: MiddlewareSendResponse, reader: jspb.BinaryReader): MiddlewareSendResponse;
}

export namespace MiddlewareSendResponse {
    export type AsObject = {
        msg?: string,
    }
}

export class CmdDef extends jspb.Message { 
    getName(): string;
    setName(value: string): CmdDef;
    getArgsinfo(): string;
    setArgsinfo(value: string): CmdDef;
    getInfo(): string;
    setInfo(value: string): CmdDef;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CmdDef.AsObject;
    static toObject(includeInstance: boolean, msg: CmdDef): CmdDef.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CmdDef, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CmdDef;
    static deserializeBinaryFromReader(message: CmdDef, reader: jspb.BinaryReader): CmdDef;
}

export namespace CmdDef {
    export type AsObject = {
        name: string,
        argsinfo: string,
        info: string,
    }
}

export class CmdInvocation extends jspb.Message { 
    getRoom(): string;
    setRoom(value: string): CmdInvocation;
    getFrom(): string;
    setFrom(value: string): CmdInvocation;
    getArgs(): string;
    setArgs(value: string): CmdInvocation;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): CmdInvocation.AsObject;
    static toObject(includeInstance: boolean, msg: CmdInvocation): CmdInvocation.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: CmdInvocation, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): CmdInvocation;
    static deserializeBinaryFromReader(message: CmdInvocation, reader: jspb.BinaryReader): CmdInvocation;
}

export namespace CmdInvocation {
    export type AsObject = {
        room: string,
        from: string,
        args: string,
    }
}

export class Message extends jspb.Message { 
    getRoom(): string;
    setRoom(value: string): Message;

    hasFrom(): boolean;
    clearFrom(): void;
    getFrom(): string | undefined;
    setFrom(value: string): Message;
    getMsg(): string;
    setMsg(value: string): Message;

    hasEphemeralTo(): boolean;
    clearEphemeralTo(): void;
    getEphemeralTo(): string | undefined;
    setEphemeralTo(value: string): Message;

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): Message.AsObject;
    static toObject(includeInstance: boolean, msg: Message): Message.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: Message, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): Message;
    static deserializeBinaryFromReader(message: Message, reader: jspb.BinaryReader): Message;
}

export namespace Message {
    export type AsObject = {
        room: string,
        from?: string,
        msg: string,
        ephemeralTo?: string,
    }
}

export class MessageRes extends jspb.Message { 

    serializeBinary(): Uint8Array;
    toObject(includeInstance?: boolean): MessageRes.AsObject;
    static toObject(includeInstance: boolean, msg: MessageRes): MessageRes.AsObject;
    static extensions: {[key: number]: jspb.ExtensionFieldInfo<jspb.Message>};
    static extensionsBinary: {[key: number]: jspb.ExtensionFieldBinaryInfo<jspb.Message>};
    static serializeBinaryToWriter(message: MessageRes, writer: jspb.BinaryWriter): void;
    static deserializeBinary(bytes: Uint8Array): MessageRes;
    static deserializeBinaryFromReader(message: MessageRes, reader: jspb.BinaryReader): MessageRes;
}

export namespace MessageRes {
    export type AsObject = {
    }
}

export enum EventType {
    SEND = 0,
}
