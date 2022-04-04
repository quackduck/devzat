package interfaces

type hasPronouns interface {
	Pronouns() []string
	DisplayPronouns() string
}
