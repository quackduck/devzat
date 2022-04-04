package room

import i "devzat/pkg/interfaces"

func (r *Room) Bot() i.Bot {
	return r.bot
}

func (r *Room) SetBot(bot i.Bot) {
	r.bot = bot
}
