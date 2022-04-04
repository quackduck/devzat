package interfaces

type hasPronouns interface {
	Pronouns() []string
	SetPronouns(...string)
	DisplayPronouns() string
}
