package interfaces

import (
	"github.com/gliderlabs/ssh"
	terminal "github.com/quackduck/term"
)

type hasShell interface {
	hasPrivilege
	hasSettings
	hasPrivateChat
	Session() ssh.Session
	Term() *terminal.Terminal
	Close(msg string)
	CloseQuietly()
	Disconnect()
	Writeln(from string, srcMsg string)
	RWriteln(msg string)
	Addr() string
	ID() string
}

type hasSettings interface {
	Bell() bool
	SetBell(bool)
	PingEverytime() bool
	SetPingEverytime(bool)
	IsSlack() bool
	FormatTime24() bool
	SetFormatTime24(bool)
	TimeZone() string
	SetTimeZone(string)
}
