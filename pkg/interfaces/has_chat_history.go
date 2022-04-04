package interfaces

import "devzat/pkg/models"

type hasChatHistory interface {
	Backlogs() []models.BacklogMessage
	SetBacklogs([]models.BacklogMessage)
}
