package user

import (
	"devzat/pkg/server"
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"devzat/pkg"
	"devzat/pkg/room"
	"github.com/acarl005/stripansi"
	"github.com/gliderlabs/ssh"
	terminal "github.com/quackduck/term"
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
	Name     string
	Pronouns []string
	Session  ssh.Session
	Term     *terminal.Terminal

	*room.Room
	Messaging *User // currently Messaging this User in a DirectMessage

	Color struct {
		Foreground, Background string
	}

	ID   string
	Addr string

	Window        ssh.Window
	closeOnce     sync.Once
	LastTimestamp time.Time
	JoinTime      time.Time
	timezone      *time.Location

	UserSettings
}

func (u *User) Close(msg string) {
	u.closeOnce.Do(func() {
		u.CloseQuietly()
		go u.Server.SendCurrentUsersTwitterMessage()
		if time.Since(u.JoinTime) > time.Minute/2 {
			msg += ". They were online for " + pkg.PrintPrettyDuration(time.Since(u.JoinTime))
		}

		u.Room.Broadcast(u.Room.Bot.Name(), msg)
		u.Room.Users = remove(u.Room.Users, u)
		u.Room.Cleanup()
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
	u.Room.UsersMutex.Lock()
	u.Room.Users = remove(u.Room.Users, u)
	u.Room.UsersMutex.Unlock()
	_ = u.Session.Close()
}

func (u *User) Writeln(from string, srcMsg string) {
	if strings.Contains(srcMsg, u.Name) { // is a ping
		srcMsg += "\a"
	}

	srcMsg = strings.ReplaceAll(srcMsg, `\n`, "\n")
	srcMsg = strings.ReplaceAll(srcMsg, `\`+"\n", `\n`) // let people escape newlines

	dstMsg := strings.TrimSpace(pkg.MarkdownRender(srcMsg, 0, u.Window.Width)) // No sender
	useDmDetectionHack := strings.HasSuffix(from, hackDmRecieveDetect) || strings.HasSuffix(from, hackDmSendDetect)

	if from != "" {
		renderWidth := lenString(from) + 2
		fmtMsg := "%s: %s"

		if useDmDetectionHack {
			renderWidth = lenString(from)
			fmtMsg = "%s%s\a"
		}

		rendered := pkg.MarkdownRender(srcMsg, renderWidth, u.Window.Width)
		rendered = strings.TrimSpace(rendered)
		dstMsg = fmt.Sprintf(fmtMsg, from, rendered)
	}

	if time.Since(u.LastTimestamp) > time.Minute {
		if u.timezone == nil {
			u.RWriteln(pkg.PrintPrettyDuration(time.Since(u.JoinTime)) + " in")
		} else {
			if u.FormatTime24 {
				u.RWriteln(time.Now().In(u.timezone).Format("15:04"))
			} else {
				u.RWriteln(time.Now().In(u.timezone).Format("3:04 pm"))
			}
		}

		u.LastTimestamp = time.Now()
	}

	if u.PingEverytime && from != u.Name {
		dstMsg += "\a"
	}

	if !u.Bell {
		dstMsg = strings.ReplaceAll(dstMsg, "\a", "")
	}

	_, err := u.Term.Write([]byte(dstMsg + "\n"))
	if err != nil {
		u.Close(fmt.Sprintf(fmtUserLeftChatErr, u.Name, err))
	}
}

func (u *User) RWriteln(msg string) {
	if u.Window.Width-lenString(msg) > 0 {
		u.Term.Write([]byte(strings.Repeat(" ", u.Window.Width-lenString(msg)) + msg + "\n"))
	} else {
		u.Term.Write([]byte(msg + "\n"))
	}
}

func (u *User) PickUsername(possibleName string) error {
	oldName := u.Name
	err := u.PickUsernameQuietly(possibleName)
	if err != nil {
		return err
	}
	if stripansi.Strip(u.Name) != stripansi.Strip(oldName) && stripansi.Strip(u.Name) != possibleName { // did the Name change, and is it not what the User entered?
		botName := u.Room.Bot.Name()
		u.Room.Broadcast(botName, oldName+" is now called "+u.Name)
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
		} else if otherUser, dup := u.Room.UserDuplicate(possibleName); dup {
			if otherUser == u {
				break // allow selecting the same Name as before
			}
			u.Writeln("", "Your username is already in use. Pick a different one:")
		} else {
			possibleName = cleanName(possibleName)
			break
		}

		u.Term.SetPrompt("> ")
		possibleName, err = u.Term.ReadLine()
		if err != nil {
			return err
		}
		possibleName = cleanName(possibleName)
	}

	if u.Room.Server.IsProfane(possibleName) {
		u.Server.BanUser("DevBot [grow up]", u)
		return errors.New(u.Name + "'s username contained a bad word")
	}

	u.Name = possibleName
	//u.InitColor()

	if rand.Float64() <= 0.4 { // 40% chance of being a random Color
		// changeColor also sets prompt
		_ = u.Room.ChangeColor(u, randomColor) //nolint:errcheck // we know "random" is a valid Color
		return nil
	}

	styles := u.Room.Styles.Normal
	colorName := styles[rand.Intn(len(styles))].Name
	_ = u.Room.ChangeColor(u, colorName) //nolint:errcheck // we know this is a valid Color

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
	for i := 0; i < len(u.Pronouns); i++ {
		str, _ := u.Room.ApplyColorToData(u.Pronouns[i], u.Color.Foreground, u.Color.Background)
		result += "/" + str
	}
	if result == "" {
		return result
	}
	return result[1:]
}

func (u *User) ChangeRoom(r *room.Room) {
	if u.Room == r {
		return
	}

	blue := u.Room.Colors.Blue

	u.Room.Users = remove(u.Room.Users, u)
	u.Room.Broadcast("", u.Name+" is joining "+blue.Paint(r.Name)) // tell the old room
	u.Room.Cleanup()
	u.Room = r

	if _, dup := u.Room.UserDuplicate(u.Name); dup {
		_ = u.PickUsername("") //nolint:errcheck // if reading input failed the next repl will err out
	}

	u.Room.Users = append(u.Room.Users, u)
	botName := u.Room.Bot.Name()
	u.Room.Broadcast(botName, u.Name+" has joined "+blue.Paint(u.Room.Name))
}

func (u *User) Repl() {
	for {
		line, err := u.Term.ReadLine()
		if err == io.EOF {
			u.Close(u.Name + " has left the chat")
			return
		}
		line += "\n"
		hasNewlines := false
		//oldPrompt := u.Name + ": "
		for err == terminal.ErrPasteIndicator {
			hasNewlines = true
			//u.Term.SetPrompt(strings.Repeat(" ", lenString(u.Name)+2))
			u.Term.SetPrompt("")
			additionalLine := ""
			additionalLine, err = u.Term.ReadLine()
			additionalLine = strings.ReplaceAll(additionalLine, `\n`, `\\n`)
			//additionalLine = strings.ReplaceAll(additionalLine, "\t", strings.Repeat(" ", 8))
			line += additionalLine + "\n"
		}

		if err != nil {
			u.Room.Server.Log.Println(u.Name, err)
			u.Close(u.Name + " has left the chat due to an error: " + err.Error())
			return
		}

		if len(line) > maxMsgLen { // limit msg len as early as possible.
			line = line[0:maxMsgLen]
		}

		line = strings.TrimSpace(line)
		u.Term.SetPrompt(u.Name + ": ")

		//fmt.Println("window", u.Window)
		if hasNewlines {
			u.calculateLinesTaken(u.Name+": "+line, u.Window.Width)
		} else {
			u.Term.Write([]byte(strings.Repeat("\033[A\033[2K", int(math.Ceil(float64(lenString(u.Name+line)+2)/(float64(u.Window.Width))))))) // basically, ceil(length of line divided by term width)
		}
		//u.Term.Write([]byte(strings.Repeat("\033[A\033[2K", calculateLinesTaken(u.Name+": "+line, u.Window.Width))))

		if line == "" {
			continue
		}

		u.Room.Server.AntiSpamMessages[u.ID]++
		time.AfterFunc(5*time.Second, func() {
			u.Room.Server.AntiSpamMessages[u.ID]--
		})

		if u.Room.Server.AntiSpamMessages[u.ID] >= 30 {
			botName := u.Room.Bot.Name()
			u.Room.Broadcast(botName, u.Name+", stop spamming or you could get banned.")
		}

		if u.Room.Server.AntiSpamMessages[u.ID] >= 50 {
			if !u.Room.Server.BansContains(u.Addr, u.ID) {
				u.Room.Server.Bans = append(u.Room.Server.Bans, server.Ban{Addr: u.Addr, ID: u.ID})
				_ = u.Room.Server.SaveBans()
			}

			botName := u.Room.Bot.Name()
			u.Writeln(botName, "anti-spam triggered")
			u.Close(u.Room.Colors.Red.Paint(u.Name + " has been banned for spamming"))

			return
		}

		line = u.Room.Server.ReplaceSlackEmoji(line)

		_ = u.Room.RunCommands(line, u)
	}
}

// may contain a bug ("may" because it could be the terminal's fault)
func (u *User) calculateLinesTaken(str string, width int) {
	str = stripansi.Strip(str)
	//fmt.Println("`"+str+"`", "width", width)
	pos := 0
	//lines := 1
	_, _ = u.Term.Write([]byte("\033[A\033[2K"))
	currLine := ""
	for _, c := range str {
		pos++
		currLine += string(c)
		if c == '\t' {
			pos += 8
		}
		if c == '\n' || pos > width {
			pos = 1
			//lines++
			_, _ = u.Term.Write([]byte("\033[A\033[2K"))
		}
		//fmt.Println(string(c), "`"+currLine+"`", "pos", pos, "lines", lines)
	}
	//return lines
}

func lenString(a string) int {
	return len([]rune(stripansi.Strip(a)))
}
