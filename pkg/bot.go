package pkg

type Bot interface {
	Name() string
	Room() *Room
	SetRoom(*Room)
	Chat(line string)
	Respond(messages []string, chance int)
}
