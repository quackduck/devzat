package interfaces

type hasColor interface {
	ChangeColor(string) error

	ForegroundColor() string
	SetForegroundColor(string) error

	BackgroundColor() string
	SetBackgroundColor(string) error
}
