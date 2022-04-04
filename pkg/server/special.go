package server

import (
	"devzat/pkg/commands/clear"
	v2 "devzat/pkg/user"
	"time"
)

func (s *Server) handleValentinesDay(u *v2.User) {
	if time.Now().Month() == time.February &&
		(time.Now().Day() == 14 || time.Now().Day() == 15 || time.Now().Day() == 13) {
		// TODO: add a few more random images
		u.Writeln("", "![❤️](https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/160/apple/81/heavy-black-heart_2764.png)")
		//u.Term().Write([]byte("\u001B[A\u001B[2K\u001B[A\u001B[2K")) // delete last line of rendered markdown
		time.Sleep(time.Second)
		// clear screen
		if cmd, found := s.GetCommand(clear.Name); found {
			_ = cmd.Fn("", u)
		}
	}
}
