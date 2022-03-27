package pkg

import "time"

type Message struct {
	Time       time.Time
	SenderName string
	Text       string
}

type BacklogMessage = Message
