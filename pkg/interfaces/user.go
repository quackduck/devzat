package interfaces

type User interface {
	hasShell
	hasRoom
	hasPrivilege
	hasPrivateChat
	hasName
	hasNickname
	hasColor
	hasPronouns
	Repl()
}
