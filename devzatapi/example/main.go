package main

import (
	"os"

	api "github.com/quackduck/devzat/devzatapi"
)

func main() {
	s, err := api.NewSession("devzat.hackclub.com:5556", os.Getenv("DEVZAT_TOKEN"))
	if err != nil {
		panic(err)
	}
	messageChan, middlewareResponseChan, err := s.RegisterListener(true, false, "")
	if err != nil {
		panic(err)
	}
	for {
		msg := <-messageChan
		if err = <-s.ErrorChan; err != nil {
			panic(err)
		}
		if msg.Data == "ping" {
			middlewareResponseChan <- "pong"
			err = s.SendMessage(api.Message{Room: msg.Room, From: "examplebot", Data: "Did you say ping? I think you meant pong."})
			if err != nil {
				panic(err)
			}
		} else {
			middlewareResponseChan <- msg.Data
		}
	}
}
