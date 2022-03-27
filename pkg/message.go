package pkg

import "time"

type Message struct {
	timestamp  time.Time
	senderName string
	text       string
}

type BacklogMessage = Message
