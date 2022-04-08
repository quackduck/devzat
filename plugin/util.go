package plugin

// Need to copy because not exported from the gRPC generated code
type IsEvent_Event interface {
	isEvent_Event()
}

type MiddlewareRes interface {
	_MiddlewareRes()
}

func (*MiddlewareMessage) _MiddlewareRes() {}
func (*MiddlewareDoneMessage) _MiddlewareRes() {}