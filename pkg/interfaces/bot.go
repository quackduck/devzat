package interfaces

type Bot interface {
	hasRoom
	hasName
	hasColor
	Interpret(line string)
}
