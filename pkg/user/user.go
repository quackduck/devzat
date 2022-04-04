package user

import (
	"devzat/pkg/interfaces"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/gliderlabs/ssh"
	terminal "github.com/quackduck/term"

	"devzat/pkg/util"
)

const (
	// TODO: kinda hacky DirectMessage detection
	hackDmRecieveDetect = " <- "
	hackDmSendDetect    = " -> "
)

const (
	fmtUserLeftChatErr = "%v has left the chat because of an error writing to their terminal: %v"
)

const (
	randomColor = "random"
)

const (
	maxMsgLen = 5120
)

type UserSettings struct {
	Bell          bool
	PingEverytime bool
	IsSlack       bool
	FormatTime24  bool
}

type User struct {
	name     string
	pronouns []string
	session  ssh.Session
	term     *terminal.Terminal

	room     interfaces.Room
	dmTarget interfaces.User // currently DMTarget this User in a DirectMessage

	Color struct {
		Foreground, Background string
	}

	id   string
	addr string

	Window        ssh.Window
	closeOnce     sync.Once
	LastTimestamp time.Time
	JoinTime      time.Time
	timezone      *time.Location

	UserSettings
}

func (u *User) Name() string {
	return u.name
}

func (u *User) Pronouns() []string {
	return u.pronouns
}

func (u *User) Session() ssh.Session {
	return u.session
}

func (u *User) Term() *terminal.Terminal {
	return u.term
}

func (u *User) Room() interfaces.Room {
	return u.room
}

func (u *User) Addr() string {
	return u.addr
}

func (u *User) ID() string {
	return u.id
}

func (u *User) DMTarget() interfaces.User {
	return u.dmTarget
}

func (u *User) SetDMTarget(target interfaces.User) {
	u.dmTarget = target
}

func (u *User) Bell() bool {
	return u.UserSettings.Bell
}

func (u *User) SetBell(b bool) {
	u.UserSettings.Bell = b
}

func (u *User) PingEverytime() bool {
	return u.UserSettings.PingEverytime
}

func (u *User) SetPingEverytime(b bool) {
	u.UserSettings.PingEverytime = b
}

func (u *User) IsSlack() bool {
	return u.UserSettings.IsSlack
}

func (u *User) FormatTime24() bool {
	return u.UserSettings.FormatTime24
}

func (u *User) Close(msg string) {
	u.closeOnce.Do(func() {
		u.CloseQuietly()
		go u.Room().Server.SendCurrentUsersTwitterMessage()
		if time.Since(u.JoinTime) > time.Minute/2 {
			msg += ". They were online for " + util.PrintPrettyDuration(time.Since(u.JoinTime))
		}

		u.Room().BotCast(msg)
		u.Room().Users = remove(u.Room().Users, u)
		u.Room().Cleanup()
	})
}

func remove(s []*User, a *User) []*User {
	for j := range s {
		if s[j] == a {
			return append(s[:j], s[j+1:]...)
		}
	}
	return s
}

func (u *User) CloseQuietly() {
	u.Room().UsersMutex.Lock()
	u.Room().Users = remove(u.Room().Users, u)
	u.Room().UsersMutex.Unlock()
	_ = u.session.Close()
}

func (u *User) Writeln(from string, srcMsg string) {
	if strings.Contains(srcMsg, u.name) { // is a ping
		srcMsg += "\a"
	}

	srcMsg = strings.ReplaceAll(srcMsg, `\n`, "\n")
	srcMsg = strings.ReplaceAll(srcMsg, `\`+"\n", `\n`) // let people escape newlines

	dstMsg := strings.TrimSpace(util.MarkdownRender(srcMsg, 0, u.Window.Width)) // No sender
	useDmDetectionHack := strings.HasSuffix(from, hackDmRecieveDetect) || strings.HasSuffix(from, hackDmSendDetect)

	if from != "" {
		renderWidth := lenString(from) + 2
		fmtMsg := "%s: %s"

		if useDmDetectionHack {
			renderWidth = lenString(from)
			fmtMsg = "%s%s\a"
		}

		rendered := util.MarkdownRender(srcMsg, renderWidth, u.Window.Width)
		rendered = strings.TrimSpace(rendered)
		dstMsg = fmt.Sprintf(fmtMsg, from, rendered)
	}

	if time.Since(u.LastTimestamp) > time.Minute {
		if u.timezone == nil {
			u.RWriteln(util.PrintPrettyDuration(time.Since(u.JoinTime)) + " in")
		} else {
			if u.UserSettings.FormatTime24 {
				u.RWriteln(time.Now().In(u.timezone).Format("15:04"))
			} else {
				u.RWriteln(time.Now().In(u.timezone).Format("3:04 pm"))
			}
		}

		u.LastTimestamp = time.Now()
	}

	if u.UserSettings.PingEverytime && from != u.name {
		dstMsg += "\a"
	}

	if !u.UserSettings.Bell {
		dstMsg = strings.ReplaceAll(dstMsg, "\a", "")
	}

	_, err := u.Term().Write([]byte(dstMsg + "\n"))
	if err != nil {
		u.Close(fmt.Sprintf(fmtUserLeftChatErr, u.name, err))
	}
}

func (u *User) RWriteln(msg string) {
	if u.Window.Width-lenString(msg) > 0 {
		u.Term().Write([]byte(strings.Repeat(" ", u.Window.Width-lenString(msg)) + msg + "\n"))
	} else {
		u.Term().Write([]byte(msg + "\n"))
	}
}

func (u *User) PickUsername(possibleName string) error {
	oldName := u.name
	err := u.PickUsernameQuietly(possibleName)
	if err != nil {
		return err
	}
	if stripansi.Strip(u.name) != stripansi.Strip(oldName) && stripansi.Strip(u.name) != possibleName { // did the Name change, and is it not what the User entered?
		botName := u.Room().Bot().Name()
		u.Room().Broadcast(botName, oldName+" is now called "+u.name)
	}
	return nil
}

func (u *User) PickUsernameQuietly(possibleName string) error {
	possibleName = cleanName(possibleName)
	var err error
	for {
		if possibleName == "" {
		} else if strings.HasPrefix(possibleName, "#") || possibleName == "botName" {
			u.Writeln("", "Your username is invalid. Pick a different one:")
		} else if otherUser, dup := u.Room().UserDuplicate(possibleName); dup {
			if otherUser == u {
				break // allow selecting the same Name as before
			}
			u.Writeln("", "Your username is already in use. Pick a different one:")
		} else {
			possibleName = cleanName(possibleName)
			break
		}

		u.Term().SetPrompt("> ")
		possibleName, err = u.Term().ReadLine()
		if err != nil {
			return err
		}
		possibleName = cleanName(possibleName)
	}

	if u.Room().Server.IsProfane(possibleName) {
		u.Room().Server.BanUser("DevBot [grow up]", u)
		return errors.New(u.name + "'s username contained a bad word")
	}

	u.name = possibleName
	//u.InitColor()

	if rand.Float64() <= 0.4 { // 40% chance of being a random Color
		// changeColor also sets prompt
		_ = u.ChangeColor(randomColor) //nolint:errcheck // we know "random" is a valid Color
		return nil
	}

	styles := u.Room().Styles.Normal
	colorName := styles[rand.Intn(len(styles))].Name
	_ = u.ChangeColor(colorName) //nolint:errcheck // we know this is a valid Color

	return nil
}

// removes arrows, spaces and non-ascii-printable characters
func cleanName(name string) string {
	s := ""
	name = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
		strings.TrimSpace(strings.Split(name, "\n")[0]), // use one trimmed line
		"<-", ""),
		"->", ""),
		" ", "-")
	if len([]rune(name)) > 27 {
		name = string([]rune(name)[:27])
	}
	for i := 0; i < len(name); i++ {
		if 33 <= name[i] && name[i] <= 126 { // ascii printables only: '!' to '~'
			s += string(name[i])
		}
	}
	return s
}

func (u *User) DisplayPronouns() string {
	result := ""
	for i := 0; i < len(u.pronouns); i++ {
		str, _ := u.Room().ApplyColorToData(u.pronouns[i], u.Color.Foreground, u.Color.Background)
		result += "/" + str
	}
	if result == "" {
		return result
	}
	return result[1:]
}

func (u *User) ChangeRoom(r interfaces.Room) {
	if u.room == r {
		return
	}

	blue := u.Room().Colors.Blue

	u.Room().Users = remove(u.Room().Users, u)
	u.Room().Broadcast("", u.name+" is joining "+blue.Paint(r.Name)) // tell the old room
	u.Room().Cleanup()
	u.room = r

	if _, dup := u.Room().UserDuplicate(u.name); dup {
		_ = u.PickUsername("") //nolint:errcheck // if reading input failed the next repl will err out
	}

	u.Room().Users = append(u.Room().Users, u)
	botName := u.Room().Bot().Name()
	u.Room().Broadcast(botName, u.name+" has joined "+blue.Paint(u.Room().Name))
}

// Applies color from name
func (u *User) ChangeColor(colorName string) error {

}
