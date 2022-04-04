package interfaces

type hasBot interface {
	Bot() Bot
	BotCast(msg string)
}
