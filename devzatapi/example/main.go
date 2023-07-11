package main

import (
	"github.com/quackduck/devzat/devzatapi"
	"os"
)

func main() {
	s, err := devzatapi.NewSession("devzat.hackclub.com:5556", os.Getenv("DEVZAT_TOKEN"))
	if err != nil {
		panic(err)
	}
	messageChan, middlewareResponseChan, err := s.RegisterListener(true, false, "")
	if err != nil {
		panic(err)
	}
	for {
		msg := <-messageChan
		if msg.Error != nil {
			panic(err)
		}
		if msg.Data == "ping" {
			middlewareResponseChan <- "pong"
			err = s.SendMessage(msg.Room, "examplebot", "Did you say ping? I think you meant pong.", "")
			if err != nil {
				panic(err)
			}
		} else {
			middlewareResponseChan <- msg.Data
		}
	}
}
