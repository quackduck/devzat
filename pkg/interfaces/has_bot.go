package interfaces

type hasBot interface {
	Bot() Bot
	SetBot(Bot)
	BotCast(msg string)
}
