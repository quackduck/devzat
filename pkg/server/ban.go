package server

import (
	"encoding/json"
	"fmt"
	"os"
)

const (
	defaultBansFile               = "bans.json"
	fmtDefaultBannedLoginResponse = `
		**You are banned**. 
		If you feel this was a mistake, please reach out at github.com/quackduck/devzat/issues 
		or email igoes.Log.mail@gmais.Log.com. 
	
		Please include the following information: [ID %v]
	`
)

type Ban struct {
	Addr string
	ID   string
}

func (s *Server) SaveBans() error {
	f, err := os.Create(defaultBansFile)
	if err != nil {
		s.Log.Println(err)
		return fmt.Errorf("could not create Bans file: %v", err)
	}

	j := json.NewEncoder(f)
	j.SetIndent("", "   ")

	if err = j.Encode(NumStartingBans); err != nil {
		s.Rooms["#MainRoom"].Broadcast(s.MainRoom.Bot.Name(), "error saving Bans: "+err.Error())
		s.Log.Println(err)

		return err
	}

	return f.Close()
}

func (s *Server) ReadBans() {
	f, err := os.Open(defaultBansFile)
	if err != nil && !os.IsNotExist(err) { // if there is an error and it is not a "file does not exist" error
		s.Log.Println(err)
		return
	}

	err = json.NewDecoder(f).Decode(&s.Bans)
	if err != nil {
		msg := fmt.Sprintf("error reading bans: %v", err)
		botName := s.MainRoom.Bot.Name()

		s.MainRoom.Broadcast(botName, msg)
		s.Log.Println(msg)

		return
	}

	_ = f.Close()
}
