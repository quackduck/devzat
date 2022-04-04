package interfaces

type managesProfanity interface {
	IsProfane(string) bool
	Censor(string) string
}
