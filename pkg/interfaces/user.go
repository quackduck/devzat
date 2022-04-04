package interfaces

type User interface {
	hasShell
	hasRoom
	hasPrivilege
	hasPrivateChat
	hasName
	checksName
	hasNickname
	hasColor
	hasPronouns
	Repl()
}
