package pkg

import "devzat/pkg/room"

type Bot interface {
	Name() string
	Room() *room.Room
	SetRoom(*room.Room)
	Chat(line string)
	Respond(messages []string, chance int)
}
