package interfaces

import "github.com/gliderlabs/ssh"

type Server interface {
	serverManagement
	commandRegistry
	hasBot
	hasLog
	hasAdmins
	hasSlackIntegration

	Init() error
	MainRoom() Room
	NewUserFromSSH(session ssh.Session) (User, error)
	UniverseBroadcast(senderName, msg string)
	SendCurrentUsersTwitterMessage()
	SetUserColor(User, string) error
}

type serverManagement interface {
	managesRooms
	managesBans
	managesAdmins
	managesUsers
	managesColors
	managesProfanity
	managesSpam
}
