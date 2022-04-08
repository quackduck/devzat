// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.0
// 	protoc        v3.15.8
// source: plugin.proto

package main

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type EventType int32

const (
	EventType_SEND EventType = 0
)

// Enum value maps for EventType.
var (
	EventType_name = map[int32]string{
		0: "SEND",
	}
	EventType_value = map[string]int32{
		"SEND": 0,
	}
)

func (x EventType) Enum() *EventType {
	p := new(EventType)
	*p = x
	return p
}

func (x EventType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (EventType) Descriptor() protoreflect.EnumDescriptor {
	return file_plugin_proto_enumTypes[0].Descriptor()
}

func (EventType) Type() protoreflect.EnumType {
	return &file_plugin_proto_enumTypes[0]
}

func (x EventType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use EventType.Descriptor instead.
func (EventType) EnumDescriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{0}
}

type Image_ImageSrcType int32

const (
	Image_FILE Image_ImageSrcType = 0
	Image_URL  Image_ImageSrcType = 1
	Image_B64  Image_ImageSrcType = 2
)

// Enum value maps for Image_ImageSrcType.
var (
	Image_ImageSrcType_name = map[int32]string{
		0: "FILE",
		1: "URL",
		2: "B64",
	}
	Image_ImageSrcType_value = map[string]int32{
		"FILE": 0,
		"URL":  1,
		"B64":  2,
	}
)

func (x Image_ImageSrcType) Enum() *Image_ImageSrcType {
	p := new(Image_ImageSrcType)
	*p = x
	return p
}

func (x Image_ImageSrcType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Image_ImageSrcType) Descriptor() protoreflect.EnumDescriptor {
	return file_plugin_proto_enumTypes[1].Descriptor()
}

func (Image_ImageSrcType) Type() protoreflect.EnumType {
	return &file_plugin_proto_enumTypes[1]
}

func (x Image_ImageSrcType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Image_ImageSrcType.Descriptor instead.
func (Image_ImageSrcType) EnumDescriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{7, 0}
}

type SendEvent struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	From string `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	Msg  string `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
}

func (x *SendEvent) Reset() {
	*x = SendEvent{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SendEvent) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SendEvent) ProtoMessage() {}

func (x *SendEvent) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SendEvent.ProtoReflect.Descriptor instead.
func (*SendEvent) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{0}
}

func (x *SendEvent) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *SendEvent) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

type Listener struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Event      EventType `protobuf:"varint,1,opt,name=event,proto3,enum=plugin.EventType" json:"event,omitempty"`
	Middleware *bool     `protobuf:"varint,2,opt,name=middleware,proto3,oneof" json:"middleware,omitempty"`
	Once       *bool     `protobuf:"varint,3,opt,name=once,proto3,oneof" json:"once,omitempty"`
}

func (x *Listener) Reset() {
	*x = Listener{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Listener) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Listener) ProtoMessage() {}

func (x *Listener) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Listener.ProtoReflect.Descriptor instead.
func (*Listener) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{1}
}

func (x *Listener) GetEvent() EventType {
	if x != nil {
		return x.Event
	}
	return EventType_SEND
}

func (x *Listener) GetMiddleware() bool {
	if x != nil && x.Middleware != nil {
		return *x.Middleware
	}
	return false
}

func (x *Listener) GetOnce() bool {
	if x != nil && x.Once != nil {
		return *x.Once
	}
	return false
}

type Event struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// Types that are assignable to Event:
	//	*Event_SendEvent
	Event isEvent_Event `protobuf_oneof:"event"`
}

func (x *Event) Reset() {
	*x = Event{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Event) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Event) ProtoMessage() {}

func (x *Event) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Event.ProtoReflect.Descriptor instead.
func (*Event) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{2}
}

func (m *Event) GetEvent() isEvent_Event {
	if m != nil {
		return m.Event
	}
	return nil
}

func (x *Event) GetSendEvent() *SendEvent {
	if x, ok := x.GetEvent().(*Event_SendEvent); ok {
		return x.SendEvent
	}
	return nil
}

type isEvent_Event interface {
	isEvent_Event()
}

type Event_SendEvent struct {
	SendEvent *SendEvent `protobuf:"bytes,1,opt,name=send_event,json=sendEvent,proto3,oneof"`
}

func (*Event_SendEvent) isEvent_Event() {}

type CmdDef struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name  string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Usage string `protobuf:"bytes,2,opt,name=usage,proto3" json:"usage,omitempty"`
}

func (x *CmdDef) Reset() {
	*x = CmdDef{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CmdDef) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CmdDef) ProtoMessage() {}

func (x *CmdDef) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CmdDef.ProtoReflect.Descriptor instead.
func (*CmdDef) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{3}
}

func (x *CmdDef) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *CmdDef) GetUsage() string {
	if x != nil {
		return x.Usage
	}
	return ""
}

type CmdInvokation struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	From string   `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	Args []string `protobuf:"bytes,2,rep,name=args,proto3" json:"args,omitempty"`
}

func (x *CmdInvokation) Reset() {
	*x = CmdInvokation{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *CmdInvokation) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*CmdInvokation) ProtoMessage() {}

func (x *CmdInvokation) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use CmdInvokation.ProtoReflect.Descriptor instead.
func (*CmdInvokation) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{4}
}

func (x *CmdInvokation) GetFrom() string {
	if x != nil {
		return x.From
	}
	return ""
}

func (x *CmdInvokation) GetArgs() []string {
	if x != nil {
		return x.Args
	}
	return nil
}

type Message struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	From      *string `protobuf:"bytes,1,opt,name=from,proto3,oneof" json:"from,omitempty"`
	Msg       string  `protobuf:"bytes,2,opt,name=msg,proto3" json:"msg,omitempty"`
	Ephemeral *bool   `protobuf:"varint,3,opt,name=ephemeral,proto3,oneof" json:"ephemeral,omitempty"`
}

func (x *Message) Reset() {
	*x = Message{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Message) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Message) ProtoMessage() {}

func (x *Message) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Message.ProtoReflect.Descriptor instead.
func (*Message) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{5}
}

func (x *Message) GetFrom() string {
	if x != nil && x.From != nil {
		return *x.From
	}
	return ""
}

func (x *Message) GetMsg() string {
	if x != nil {
		return x.Msg
	}
	return ""
}

func (x *Message) GetEphemeral() bool {
	if x != nil && x.Ephemeral != nil {
		return *x.Ephemeral
	}
	return false
}

type MessageRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MessageRes) Reset() {
	*x = MessageRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MessageRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MessageRes) ProtoMessage() {}

func (x *MessageRes) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MessageRes.ProtoReflect.Descriptor instead.
func (*MessageRes) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{6}
}

type Image struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	SrcType Image_ImageSrcType `protobuf:"varint,1,opt,name=src_type,json=srcType,proto3,enum=plugin.Image_ImageSrcType" json:"src_type,omitempty"`
	Src     string             `protobuf:"bytes,2,opt,name=src,proto3" json:"src,omitempty"`
}

func (x *Image) Reset() {
	*x = Image{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Image) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Image) ProtoMessage() {}

func (x *Image) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Image.ProtoReflect.Descriptor instead.
func (*Image) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{7}
}

func (x *Image) GetSrcType() Image_ImageSrcType {
	if x != nil {
		return x.SrcType
	}
	return Image_FILE
}

func (x *Image) GetSrc() string {
	if x != nil {
		return x.Src
	}
	return ""
}

type ImageRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *ImageRes) Reset() {
	*x = ImageRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ImageRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ImageRes) ProtoMessage() {}

func (x *ImageRes) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ImageRes.ProtoReflect.Descriptor instead.
func (*ImageRes) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{8}
}

type MiddlewareMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Msg *string `protobuf:"bytes,1,opt,name=msg,proto3,oneof" json:"msg,omitempty"`
}

func (x *MiddlewareMessage) Reset() {
	*x = MiddlewareMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MiddlewareMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MiddlewareMessage) ProtoMessage() {}

func (x *MiddlewareMessage) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MiddlewareMessage.ProtoReflect.Descriptor instead.
func (*MiddlewareMessage) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{9}
}

func (x *MiddlewareMessage) GetMsg() string {
	if x != nil && x.Msg != nil {
		return *x.Msg
	}
	return ""
}

type MiddlewareEditMessageRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MiddlewareEditMessageRes) Reset() {
	*x = MiddlewareEditMessageRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MiddlewareEditMessageRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MiddlewareEditMessageRes) ProtoMessage() {}

func (x *MiddlewareEditMessageRes) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MiddlewareEditMessageRes.ProtoReflect.Descriptor instead.
func (*MiddlewareEditMessageRes) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{10}
}

type MiddlewareDoneMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MiddlewareDoneMessage) Reset() {
	*x = MiddlewareDoneMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[11]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MiddlewareDoneMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MiddlewareDoneMessage) ProtoMessage() {}

func (x *MiddlewareDoneMessage) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[11]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MiddlewareDoneMessage.ProtoReflect.Descriptor instead.
func (*MiddlewareDoneMessage) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{11}
}

type MiddlewareDoneRes struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *MiddlewareDoneRes) Reset() {
	*x = MiddlewareDoneRes{}
	if protoimpl.UnsafeEnabled {
		mi := &file_plugin_proto_msgTypes[12]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MiddlewareDoneRes) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MiddlewareDoneRes) ProtoMessage() {}

func (x *MiddlewareDoneRes) ProtoReflect() protoreflect.Message {
	mi := &file_plugin_proto_msgTypes[12]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MiddlewareDoneRes.ProtoReflect.Descriptor instead.
func (*MiddlewareDoneRes) Descriptor() ([]byte, []int) {
	return file_plugin_proto_rawDescGZIP(), []int{12}
}

var File_plugin_proto protoreflect.FileDescriptor

var file_plugin_proto_rawDesc = []byte{
	0x0a, 0x0c, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x06,
	0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x22, 0x31, 0x0a, 0x09, 0x53, 0x65, 0x6e, 0x64, 0x45, 0x76,
	0x65, 0x6e, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x12, 0x10, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x22, 0x89, 0x01, 0x0a, 0x08, 0x4c, 0x69,
	0x73, 0x74, 0x65, 0x6e, 0x65, 0x72, 0x12, 0x27, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x0e, 0x32, 0x11, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x45,
	0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x52, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x12,
	0x23, 0x0a, 0x0a, 0x6d, 0x69, 0x64, 0x64, 0x6c, 0x65, 0x77, 0x61, 0x72, 0x65, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x08, 0x48, 0x00, 0x52, 0x0a, 0x6d, 0x69, 0x64, 0x64, 0x6c, 0x65, 0x77, 0x61, 0x72,
	0x65, 0x88, 0x01, 0x01, 0x12, 0x17, 0x0a, 0x04, 0x6f, 0x6e, 0x63, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x08, 0x48, 0x01, 0x52, 0x04, 0x6f, 0x6e, 0x63, 0x65, 0x88, 0x01, 0x01, 0x42, 0x0d, 0x0a,
	0x0b, 0x5f, 0x6d, 0x69, 0x64, 0x64, 0x6c, 0x65, 0x77, 0x61, 0x72, 0x65, 0x42, 0x07, 0x0a, 0x05,
	0x5f, 0x6f, 0x6e, 0x63, 0x65, 0x22, 0x44, 0x0a, 0x05, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x12, 0x32,
	0x0a, 0x0a, 0x73, 0x65, 0x6e, 0x64, 0x5f, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x11, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x53, 0x65, 0x6e, 0x64,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x48, 0x00, 0x52, 0x09, 0x73, 0x65, 0x6e, 0x64, 0x45, 0x76, 0x65,
	0x6e, 0x74, 0x42, 0x07, 0x0a, 0x05, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x22, 0x32, 0x0a, 0x06, 0x43,
	0x6d, 0x64, 0x44, 0x65, 0x66, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x14, 0x0a, 0x05, 0x75, 0x73, 0x61,
	0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x75, 0x73, 0x61, 0x67, 0x65, 0x22,
	0x37, 0x0a, 0x0d, 0x43, 0x6d, 0x64, 0x49, 0x6e, 0x76, 0x6f, 0x6b, 0x61, 0x74, 0x69, 0x6f, 0x6e,
	0x12, 0x12, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x66, 0x72, 0x6f, 0x6d, 0x12, 0x12, 0x0a, 0x04, 0x61, 0x72, 0x67, 0x73, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x09, 0x52, 0x04, 0x61, 0x72, 0x67, 0x73, 0x22, 0x6e, 0x0a, 0x07, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x17, 0x0a, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x48, 0x00, 0x52, 0x04, 0x66, 0x72, 0x6f, 0x6d, 0x88, 0x01, 0x01, 0x12, 0x10, 0x0a, 0x03,
	0x6d, 0x73, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x12, 0x21,
	0x0a, 0x09, 0x65, 0x70, 0x68, 0x65, 0x6d, 0x65, 0x72, 0x61, 0x6c, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x08, 0x48, 0x01, 0x52, 0x09, 0x65, 0x70, 0x68, 0x65, 0x6d, 0x65, 0x72, 0x61, 0x6c, 0x88, 0x01,
	0x01, 0x42, 0x07, 0x0a, 0x05, 0x5f, 0x66, 0x72, 0x6f, 0x6d, 0x42, 0x0c, 0x0a, 0x0a, 0x5f, 0x65,
	0x70, 0x68, 0x65, 0x6d, 0x65, 0x72, 0x61, 0x6c, 0x22, 0x0c, 0x0a, 0x0a, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x52, 0x65, 0x73, 0x22, 0x7c, 0x0a, 0x05, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x12,
	0x35, 0x0a, 0x08, 0x73, 0x72, 0x63, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x0e, 0x32, 0x1a, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x49, 0x6d, 0x61, 0x67, 0x65,
	0x2e, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x53, 0x72, 0x63, 0x54, 0x79, 0x70, 0x65, 0x52, 0x07, 0x73,
	0x72, 0x63, 0x54, 0x79, 0x70, 0x65, 0x12, 0x10, 0x0a, 0x03, 0x73, 0x72, 0x63, 0x18, 0x02, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x03, 0x73, 0x72, 0x63, 0x22, 0x2a, 0x0a, 0x0c, 0x49, 0x6d, 0x61, 0x67,
	0x65, 0x53, 0x72, 0x63, 0x54, 0x79, 0x70, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x46, 0x49, 0x4c, 0x45,
	0x10, 0x00, 0x12, 0x07, 0x0a, 0x03, 0x55, 0x52, 0x4c, 0x10, 0x01, 0x12, 0x07, 0x0a, 0x03, 0x42,
	0x36, 0x34, 0x10, 0x02, 0x22, 0x0a, 0x0a, 0x08, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x52, 0x65, 0x73,
	0x22, 0x32, 0x0a, 0x11, 0x4d, 0x69, 0x64, 0x64, 0x6c, 0x65, 0x77, 0x61, 0x72, 0x65, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x15, 0x0a, 0x03, 0x6d, 0x73, 0x67, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x48, 0x00, 0x52, 0x03, 0x6d, 0x73, 0x67, 0x88, 0x01, 0x01, 0x42, 0x06, 0x0a, 0x04,
	0x5f, 0x6d, 0x73, 0x67, 0x22, 0x1a, 0x0a, 0x18, 0x4d, 0x69, 0x64, 0x64, 0x6c, 0x65, 0x77, 0x61,
	0x72, 0x65, 0x45, 0x64, 0x69, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x73,
	0x22, 0x17, 0x0a, 0x15, 0x4d, 0x69, 0x64, 0x64, 0x6c, 0x65, 0x77, 0x61, 0x72, 0x65, 0x44, 0x6f,
	0x6e, 0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x13, 0x0a, 0x11, 0x4d, 0x69, 0x64,
	0x64, 0x6c, 0x65, 0x77, 0x61, 0x72, 0x65, 0x44, 0x6f, 0x6e, 0x65, 0x52, 0x65, 0x73, 0x2a, 0x15,
	0x0a, 0x09, 0x45, 0x76, 0x65, 0x6e, 0x74, 0x54, 0x79, 0x70, 0x65, 0x12, 0x08, 0x0a, 0x04, 0x53,
	0x45, 0x4e, 0x44, 0x10, 0x00, 0x32, 0xfb, 0x02, 0x0a, 0x06, 0x50, 0x6c, 0x75, 0x67, 0x69, 0x6e,
	0x12, 0x35, 0x0a, 0x10, 0x52, 0x65, 0x67, 0x69, 0x73, 0x74, 0x65, 0x72, 0x4c, 0x69, 0x73, 0x74,
	0x65, 0x6e, 0x65, 0x72, 0x12, 0x10, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x4c, 0x69,
	0x73, 0x74, 0x65, 0x6e, 0x65, 0x72, 0x1a, 0x0d, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e,
	0x45, 0x76, 0x65, 0x6e, 0x74, 0x30, 0x01, 0x12, 0x36, 0x0a, 0x0b, 0x52, 0x65, 0x67, 0x69, 0x73,
	0x74, 0x65, 0x72, 0x43, 0x6d, 0x64, 0x12, 0x0e, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e,
	0x43, 0x6d, 0x64, 0x44, 0x65, 0x66, 0x1a, 0x15, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e,
	0x43, 0x6d, 0x64, 0x49, 0x6e, 0x76, 0x6f, 0x6b, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x30, 0x01, 0x12,
	0x32, 0x0a, 0x0b, 0x53, 0x65, 0x6e, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x0f,
	0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a,
	0x12, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x52, 0x65, 0x73, 0x12, 0x2c, 0x0a, 0x09, 0x53, 0x65, 0x6e, 0x64, 0x49, 0x6d, 0x61, 0x67, 0x65,
	0x12, 0x0d, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x1a,
	0x10, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x49, 0x6d, 0x61, 0x67, 0x65, 0x52, 0x65,
	0x73, 0x12, 0x54, 0x0a, 0x15, 0x4d, 0x69, 0x64, 0x64, 0x6c, 0x65, 0x77, 0x61, 0x72, 0x65, 0x45,
	0x64, 0x69, 0x74, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x12, 0x19, 0x2e, 0x70, 0x6c, 0x75,
	0x67, 0x69, 0x6e, 0x2e, 0x4d, 0x69, 0x64, 0x64, 0x6c, 0x65, 0x77, 0x61, 0x72, 0x65, 0x4d, 0x65,
	0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x20, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69, 0x6e, 0x2e, 0x4d,
	0x69, 0x64, 0x64, 0x6c, 0x65, 0x77, 0x61, 0x72, 0x65, 0x45, 0x64, 0x69, 0x74, 0x4d, 0x65, 0x73,
	0x73, 0x61, 0x67, 0x65, 0x52, 0x65, 0x73, 0x12, 0x4a, 0x0a, 0x0e, 0x4d, 0x69, 0x64, 0x64, 0x6c,
	0x65, 0x77, 0x61, 0x72, 0x65, 0x44, 0x6f, 0x6e, 0x65, 0x12, 0x1d, 0x2e, 0x70, 0x6c, 0x75, 0x67,
	0x69, 0x6e, 0x2e, 0x4d, 0x69, 0x64, 0x64, 0x6c, 0x65, 0x77, 0x61, 0x72, 0x65, 0x44, 0x6f, 0x6e,
	0x65, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x1a, 0x19, 0x2e, 0x70, 0x6c, 0x75, 0x67, 0x69,
	0x6e, 0x2e, 0x4d, 0x69, 0x64, 0x64, 0x6c, 0x65, 0x77, 0x61, 0x72, 0x65, 0x44, 0x6f, 0x6e, 0x65,
	0x52, 0x65, 0x73, 0x42, 0x22, 0x5a, 0x20, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x71, 0x75, 0x61, 0x63, 0x6b, 0x64, 0x75, 0x63, 0x6b, 0x2f, 0x64, 0x65, 0x76, 0x7a,
	0x61, 0x74, 0x2f, 0x6d, 0x61, 0x69, 0x6e, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_plugin_proto_rawDescOnce sync.Once
	file_plugin_proto_rawDescData = file_plugin_proto_rawDesc
)

func file_plugin_proto_rawDescGZIP() []byte {
	file_plugin_proto_rawDescOnce.Do(func() {
		file_plugin_proto_rawDescData = protoimpl.X.CompressGZIP(file_plugin_proto_rawDescData)
	})
	return file_plugin_proto_rawDescData
}

var file_plugin_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_plugin_proto_msgTypes = make([]protoimpl.MessageInfo, 13)
var file_plugin_proto_goTypes = []interface{}{
	(EventType)(0),                   // 0: plugin.EventType
	(Image_ImageSrcType)(0),          // 1: plugin.Image.ImageSrcType
	(*SendEvent)(nil),                // 2: plugin.SendEvent
	(*Listener)(nil),                 // 3: plugin.Listener
	(*Event)(nil),                    // 4: plugin.Event
	(*CmdDef)(nil),                   // 5: plugin.CmdDef
	(*CmdInvokation)(nil),            // 6: plugin.CmdInvokation
	(*Message)(nil),                  // 7: plugin.Message
	(*MessageRes)(nil),               // 8: plugin.MessageRes
	(*Image)(nil),                    // 9: plugin.Image
	(*ImageRes)(nil),                 // 10: plugin.ImageRes
	(*MiddlewareMessage)(nil),        // 11: plugin.MiddlewareMessage
	(*MiddlewareEditMessageRes)(nil), // 12: plugin.MiddlewareEditMessageRes
	(*MiddlewareDoneMessage)(nil),    // 13: plugin.MiddlewareDoneMessage
	(*MiddlewareDoneRes)(nil),        // 14: plugin.MiddlewareDoneRes
}
var file_plugin_proto_depIdxs = []int32{
	0,  // 0: plugin.Listener.event:type_name -> plugin.EventType
	2,  // 1: plugin.Event.send_event:type_name -> plugin.SendEvent
	1,  // 2: plugin.Image.src_type:type_name -> plugin.Image.ImageSrcType
	3,  // 3: plugin.Plugin.RegisterListener:input_type -> plugin.Listener
	5,  // 4: plugin.Plugin.RegisterCmd:input_type -> plugin.CmdDef
	7,  // 5: plugin.Plugin.SendMessage:input_type -> plugin.Message
	9,  // 6: plugin.Plugin.SendImage:input_type -> plugin.Image
	11, // 7: plugin.Plugin.MiddlewareEditMessage:input_type -> plugin.MiddlewareMessage
	13, // 8: plugin.Plugin.MiddlewareDone:input_type -> plugin.MiddlewareDoneMessage
	4,  // 9: plugin.Plugin.RegisterListener:output_type -> plugin.Event
	6,  // 10: plugin.Plugin.RegisterCmd:output_type -> plugin.CmdInvokation
	8,  // 11: plugin.Plugin.SendMessage:output_type -> plugin.MessageRes
	10, // 12: plugin.Plugin.SendImage:output_type -> plugin.ImageRes
	12, // 13: plugin.Plugin.MiddlewareEditMessage:output_type -> plugin.MiddlewareEditMessageRes
	14, // 14: plugin.Plugin.MiddlewareDone:output_type -> plugin.MiddlewareDoneRes
	9,  // [9:15] is the sub-list for method output_type
	3,  // [3:9] is the sub-list for method input_type
	3,  // [3:3] is the sub-list for extension type_name
	3,  // [3:3] is the sub-list for extension extendee
	0,  // [0:3] is the sub-list for field type_name
}

func init() { file_plugin_proto_init() }
func file_plugin_proto_init() {
	if File_plugin_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_plugin_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SendEvent); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plugin_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Listener); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plugin_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Event); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plugin_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CmdDef); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plugin_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*CmdInvokation); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plugin_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Message); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plugin_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MessageRes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plugin_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Image); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plugin_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ImageRes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plugin_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MiddlewareMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plugin_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MiddlewareEditMessageRes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plugin_proto_msgTypes[11].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MiddlewareDoneMessage); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_plugin_proto_msgTypes[12].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MiddlewareDoneRes); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	file_plugin_proto_msgTypes[1].OneofWrappers = []interface{}{}
	file_plugin_proto_msgTypes[2].OneofWrappers = []interface{}{
		(*Event_SendEvent)(nil),
	}
	file_plugin_proto_msgTypes[5].OneofWrappers = []interface{}{}
	file_plugin_proto_msgTypes[9].OneofWrappers = []interface{}{}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_plugin_proto_rawDesc,
			NumEnums:      2,
			NumMessages:   13,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_plugin_proto_goTypes,
		DependencyIndexes: file_plugin_proto_depIdxs,
		EnumInfos:         file_plugin_proto_enumTypes,
		MessageInfos:      file_plugin_proto_msgTypes,
	}.Build()
	File_plugin_proto = out.File
	file_plugin_proto_rawDesc = nil
	file_plugin_proto_goTypes = nil
	file_plugin_proto_depIdxs = nil
}
