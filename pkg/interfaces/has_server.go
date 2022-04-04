package interfaces

type hasServer interface {
	Server() Server
	SetServer(Server)
}
