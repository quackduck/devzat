package plugin

// IsEvent_Event Need to copy because not exported from the gRPC generated code
//goland:noinspection GoSnakeCaseUsage
//type IsEvent_Event interface {
//	isEvent_Event()
//}

//type MiddlewareRes interface {
//	_MiddlewareRes()
//}
//
//func (*MiddlewareMessage) _MiddlewareRes() {}
//func (*MiddlewareDoneMessage) _MiddlewareRes() {}

type MiddlewareChannelMessage interface {
	_MiddlewareChannelMessage()
}

// Members of IsEvent_Event
func (*Event) _MiddlewareChannelMessage() {}

// Members of isListenerClientData_Data
// TODO
//func (*ListenerClientData_Listener) _MiddlewareChannelMessage() {}
func (*ListenerClientData_Response) _MiddlewareChannelMessage() {}
