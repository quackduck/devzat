package interfaces

type Bot interface {
	hasRoom
	hasName
	responder
	Say(line string)
}
