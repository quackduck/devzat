package v2

import "devzat/pkg/models"

const (
	Scrollback = 16
)

type chatHistory struct {
	backlog []models.BacklogMessage
}

func (sl chatHistory) init() {
	sl.backlog = make([]models.BacklogMessage, 0, Scrollback)
}
