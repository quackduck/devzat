package plugin

type MiddlewareChannelMessage interface {
	_MiddlewareChannelMessage()
}

func (*Event) _MiddlewareChannelMessage()                       {}
func (*ListenerClientData_Response) _MiddlewareChannelMessage() {}
