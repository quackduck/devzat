package v2

import (
	"crypto/sha256"
	"devzat/pkg/colors"
	"devzat/pkg/util"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/gliderlabs/ssh"
	terminal "github.com/quackduck/term"

	i "devzat/pkg/interfaces"
)

const (
	maxMsgLen = 5120

	termCommandClear = "\033[A\033[2K"
)

const (
	// TODO: change this hacky DirectMessage detection
	hackDmRecieveDetect = " <- "
	hackDmSendDetect    = " -> "
)

const (
	fmtUserLeftChatErr = "%v has left the chat because of an error writing to their terminal: %v"
)

const (
	randomColor = "random"
)

type User struct {
	settings
	colors.Formatter
	room     i.Room
	dmTarget i.User
	session  ssh.Session
	window   ssh.Window
	term     *terminal.Terminal
	once     sync.Once
	id       string
	name     string
	time     struct {
		joined, lastSeen time.Time
		zone             *time.Location
	}
	colors struct{ fg, bg string }
}

func (u *User) Bell() bool     { return u.settings.Bell }
func (u *User) SetBell(b bool) { u.settings.Bell = b }

func (u *User) PingEverytime() bool     { return u.settings.Ping }
func (u *User) SetPingEverytime(b bool) { u.settings.Ping = b }

func (u *User) IsSlack() bool      { return u.settings.IsSlack }
func (u *User) FormatTime24() bool { return u.settings.FmtHour24 }

func (u *User) DMTarget() i.User        { return u.dmTarget }
func (u *User) SetDMTarget(user i.User) { u.dmTarget = user }

func (u *User) Session() ssh.Session     { return u.session }
func (u *User) Term() *terminal.Terminal { return u.term }

func (u *User) ForegroundColor() string { return u.colors.fg }
func (u *User) SetForegroundColor(s string) error {
	u.colors.bg = s

	return u.updatePrompt()
}

func (u *User) BackgroundColor() string { return u.colors.bg }
func (u *User) SetBackgroundColor(s string) error {
	u.colors.bg = fmt.Sprintf("%s-%s", "bg", s) // ugly hacky bullshit

	return u.updatePrompt()
}

func (u *User) updatePrompt() error {
	updated, err := u.ApplyColorToData(u.Name(), u.colors.fg, u.colors.bg)
	if err != nil {
		return nil
	}

	u.name = updated
	u.Term().SetPrompt(fmt.Sprintf("%s: ", u.name))

	return nil
}

func (u *User) IsAdmin() bool {
	b, _ := u.room.Server().IsAdmin(u)

	return b
}

func (u *User) Close(msg string) {
	u.once.Do(func() {
		go u.Room().Server().SendCurrentUsersTwitterMessage()

		if time.Since(u.time.joined) > time.Minute/2 {
			msg += ". They were online for " + util.PrintPrettyDuration(time.Since(u.time.joined))
		}

		u.Disconnect()
		u.Room().BotCast(msg)
		u.Room().Cleanup()
	})
}

func (u *User) Disconnect()   { u.Room().Disconnect(u) }
func (u *User) CloseQuietly() { _ = u.session.Close() }

func (u *User) Writeln(from string, srcMsg string) {
	if strings.Contains(srcMsg, u.name) { // is a ping
		srcMsg += "\a"
	}

	srcMsg = strings.ReplaceAll(srcMsg, `\n`, "\n")
	srcMsg = strings.ReplaceAll(srcMsg, `\`+"\n", `\n`) // let people escape newlines

	dstMsg := strings.TrimSpace(util.MarkdownRender(srcMsg, 0, u.window.Width)) // No sender
	useDmDetectionHack := strings.HasSuffix(from, hackDmRecieveDetect) || strings.HasSuffix(from, hackDmSendDetect)

	if from != "" {
		renderWidth := lenString(from) + 2
		fmtMsg := "%s: %s"

		if useDmDetectionHack {
			renderWidth = lenString(from)
			fmtMsg = "%s%s\a"
		}

		rendered := util.MarkdownRender(srcMsg, renderWidth, u.window.Width)
		rendered = strings.TrimSpace(rendered)
		dstMsg = fmt.Sprintf(fmtMsg, from, rendered)
	}

	if time.Since(u.time.lastSeen) > time.Minute {
		if u.time.zone == nil {
			u.RWriteln(util.PrintPrettyDuration(time.Since(u.time.joined)) + " in")
		} else {
			if u.settings.FmtHour24 {
				u.RWriteln(time.Now().In(u.time.zone).Format("15:04"))
			} else {
				u.RWriteln(time.Now().In(u.time.zone).Format("3:04 pm"))
			}
		}

		u.time.lastSeen = time.Now()
	}

	if u.settings.Ping && from != u.name {
		dstMsg += "\a"
	}

	if !u.settings.Bell {
		dstMsg = strings.ReplaceAll(dstMsg, "\a", "")
	}

	_, err := u.Term().Write([]byte(dstMsg + "\n"))
	if err != nil {
		u.Close(fmt.Sprintf(fmtUserLeftChatErr, u.name, err))
	}
}

func (u *User) RWriteln(msg string) {
	line := []byte(msg + "\n")

	if u.window.Width-lenString(msg) > 0 {
		line = []byte(strings.Repeat(" ", u.window.Width-lenString(msg)) + msg + "\n")
	}

	_, _ = u.Term().Write(line)
}

func (u *User) Addr() string {
	return u.session.LocalAddr().String()
}

func (u *User) ID() string {
	h := sha256.Sum256([]byte(fmt.Sprintf("%s:%s", u.name, u.Nick())))
	return hex.EncodeToString(h[:])
}

func (u *User) Room() i.Room {
	return u.room
}

func (u *User) SetRoom(room i.Room) {
	u.room = room
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Nick() string {
	if u.settings.Nick == "" {
		u.settings.Nick = u.name
	}

	return u.settings.Nick
}

func (u *User) SetNick(s string) error {
	if u.room.Server().IsProfane(s) {
		return errors.New("nickname can not contain profanity")
	}

	u.settings.Nick = s

	return nil
}

func (u *User) ChangeColor(colorName string) error {
	return u.room.Server().SetUserColor(u, colorName)
}

func (u *User) Pronouns() []string {
	return append(make([]string, 0), u.settings.Pronouns...)
}

func (u *User) DisplayPronouns() string {
	if len(u.settings.Pronouns) < 1 {
		return ""
	}

	const fmtConcat = "%s/%s"

	concat := u.settings.Pronouns[0]
	for _, pronoun := range u.settings.Pronouns {
		concat = fmt.Sprintf(fmtConcat, concat, pronoun)
	}

	return fmt.Sprintf("(%s)", concat)
}

func (u *User) Repl() {
	for {
		line, err := u.Term().ReadLine()
		if err == io.EOF {
			u.Close(u.name + " has left the chat")
			return
		}

		line += "\n"
		hasNewlines := false

		for err == terminal.ErrPasteIndicator {
			hasNewlines = true
			//u.Term().SetPrompt(strings.Repeat(" ", lenString(u.name)+2))
			u.Term().SetPrompt("")
			additionalLine := ""
			additionalLine, err = u.Term().ReadLine()
			additionalLine = strings.ReplaceAll(additionalLine, `\n`, `\\n`)
			//additionalLine = strings.ReplaceAll(additionalLine, "\t", strings.Repeat(" ", 8))
			line += additionalLine + "\n"
		}

		if err != nil {
			u.room.Server().Log().Println(u.name, err)
			u.Close(u.name + " has left the chat due to an error: " + err.Error())
			return
		}

		if len(line) > maxMsgLen { // limit msg len as early as possible.
			line = line[0:maxMsgLen]
		}

		line = strings.TrimSpace(line)
		u.Term().SetPrompt(u.name + ": ")

		if hasNewlines {
			u.calculateLinesTaken(u.name+": "+line, u.window.Width)
		} else {
			// basically, ceil(length of line divided by term width)
			cmdRepeat := int(math.Ceil(float64(lenString(u.name+line)+2) / (float64(u.window.Width))))
			_, _ = u.Term().Write([]byte(strings.Repeat(termCommandClear, cmdRepeat)))
		}

		if line == "" {
			continue
		}

		u.room.Server().Antispam(u)

		line = u.room.Server().ReplaceSlackEmoji(line)

		_ = u.room.ParseUserInput(line, u)
	}
}

func (u *User) calculateLinesTaken(str string, width int) {
	pos := 0
	str = stripansi.Strip(str)
	currLine := ""

	_, _ = u.Term().Write([]byte(termCommandClear))

	for _, c := range str {
		pos++
		currLine += string(c)
		if c == '\t' {
			pos += 8
		}
		if c == '\n' || pos > width {
			pos = 1

			_, _ = u.Term().Write([]byte(termCommandClear))
		}
	}
}
