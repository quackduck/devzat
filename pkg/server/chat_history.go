package server

import "devzat/pkg/models"

const (
	defaultBacklogWindow = 16
)

type chatHistory struct {
	window  int
	backlog []models.BacklogMessage
}

func (sl *chatHistory) init() {
	sl.window = defaultBacklogWindow
	sl.backlog = make([]models.BacklogMessage, 0, defaultBacklogWindow)
}

func (sl *chatHistory) Backlogs() []models.BacklogMessage {
	return sl.backlog
}

func (sl *chatHistory) SetBacklogs(bl []models.BacklogMessage) {
	sl.backlog = bl
}
