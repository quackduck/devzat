package interfaces

type hasName interface {
	Name() string
}

type hasNickname interface {
	Nick() string
	SetNick(string) error
}
