package pkg

import (
	"errors"
	"fmt"
	"io"
	"math"
	"math/rand"
	"strings"
	"sync"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/gliderlabs/ssh"
	terminal "github.com/quackduck/term"
)

const (
	// TODO: kinda hacky DM detection
	hackDmRecieveDetect = " <- "
	hackDmSendDetect    = " -> "
)

const (
	fmtUserLeftChatErr = "%v has left the chat because of an error writing to their terminal: %v"
)

const (
	randomColor = "random"
)

type UserSettings struct {
	bell          bool
	pingEverytime bool
	isSlack       bool
	formatTime24  bool
}

type User struct {
	Name     string
	Pronouns []string
	Session  ssh.Session
	Term     *terminal.Terminal

	Room      *Room
	Messaging *User // currently Messaging this User in a DM

	Color struct {
		Foreground, Background string
	}

	id   string
	addr string

	Window        ssh.Window
	closeOnce     sync.Once
	lastTimestamp time.Time
	joinTime      time.Time
	timezone      *time.Location

	UserSettings
}

func (u *User) Close(msg string) {
	u.closeOnce.Do(func() {
		u.CloseQuietly()
		go sendCurrentUsersTwitterMessage()
		if time.Since(u.joinTime) > time.Minute/2 {
			msg += ". They were online for " + printPrettyDuration(time.Since(u.joinTime))
		}

		u.Room.Broadcast(u.Room.Bot.Name(), msg)
		u.Room.users = remove(u.Room.users, u)
		u.Room.cleanupRoom()
	})
}

func (u *User) CloseQuietly() {
	u.Room.usersMutex.Lock()
	u.Room.users = remove(u.Room.users, u)
	u.Room.usersMutex.Unlock()
	_ = u.Session.Close()
}

func (u *User) Writeln(from string, srcMsg string) {
	if strings.Contains(srcMsg, u.Name) { // is a ping
		srcMsg += "\a"
	}

	srcMsg = strings.ReplaceAll(srcMsg, `\n`, "\n")
	srcMsg = strings.ReplaceAll(srcMsg, `\`+"\n", `\n`) // let people escape newlines

	dstMsg := strings.TrimSpace(mdRender(srcMsg, 0, u.Window.Width)) // No sender
	useDmDetectionHack := strings.HasSuffix(from, hackDmRecieveDetect) || strings.HasSuffix(from, hackDmSendDetect)

	if from != "" {
		renderWidth := lenString(from) + 2
		fmtMsg := "%s: %s"

		if useDmDetectionHack {
			renderWidth = lenString(from)
			fmtMsg = "%s%s\a"
		}

		rendered := mdRender(srcMsg, renderWidth, u.Window.Width)
		rendered = strings.TrimSpace(rendered)
		dstMsg = fmt.Sprintf(fmtMsg, from, rendered)
	}

	if time.Since(u.lastTimestamp) > time.Minute {
		if u.timezone == nil {
			u.RWriteln(printPrettyDuration(time.Since(u.joinTime)) + " in")
		} else {
			if u.formatTime24 {
				u.RWriteln(time.Now().In(u.timezone).Format("15:04"))
			} else {
				u.RWriteln(time.Now().In(u.timezone).Format("3:04 pm"))
			}
		}

		u.lastTimestamp = time.Now()
	}

	if u.pingEverytime && from != u.Name {
		dstMsg += "\a"
	}

	if !u.bell {
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

	if detectBadWords(possibleName) { // sadly this is necessary
		banUser("botName [grow up]", u)
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

func (u *User) ChangeRoom(r *Room) {
	if u.Room == r {
		return
	}

	blue := u.Room.Colors.Blue

	u.Room.users = remove(u.Room.users, u)
	u.Room.Broadcast("", u.Name+" is joining "+blue.Paint(r.Name)) // tell the old room
	u.Room.cleanupRoom()
	u.Room = r

	if _, dup := u.Room.UserDuplicate(u.Name); dup {
		_ = u.PickUsername("") //nolint:errcheck // if reading input failed the next repl will err out
	}

	u.Room.users = append(u.Room.users, u)
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
		u.Term.SetPrompt(u.Name + ": ")
		line = strings.TrimSpace(line)

		if err != nil {
			u.Room.Server.Log.Println(u.Name, err)
			u.Close(u.Name + " has left the chat due to an error: " + err.Error())
			return
		}

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

		u.Room.Server.antispamMessages[u.id]++
		time.AfterFunc(5*time.Second, func() {
			u.Room.Server.antispamMessages[u.id]--
		})

		if u.Room.Server.antispamMessages[u.id] >= 30 {
			botName := u.Room.Bot.Name()
			u.Room.Broadcast(botName, u.Name+", stop spamming or you could get banned.")
		}

		if u.Room.Server.antispamMessages[u.id] >= 50 {
			if !u.Room.Server.bansContains(u.addr, u.id) {
				u.Room.Server.bans = append(u.Room.Server.bans, ban{u.addr, u.id})
				u.Room.Server.SaveBans()
			}

			botName := u.Room.Bot.Name()
			u.Writeln(botName, "anti-spam triggered")
			u.Close(u.Room.Colors.Red.Paint(u.Name + " has been banned for spamming"))

			return
		}

		line = u.Room.Server.replaceSlackEmoji(line)

		_ = runCommands(line, u)
	}
}

func (u *User) handleValentinesDay() {
	if time.Now().Month() == time.February &&
		(time.Now().Day() == 14 || time.Now().Day() == 15 || time.Now().Day() == 13) {
		// TODO: add a few more random images
		u.Writeln("", "![❤️](https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/160/apple/81/heavy-black-heart_2764.png)")
		//u.term.Write([]byte("\u001B[A\u001B[2K\u001B[A\u001B[2K")) // delete last line of rendered markdown
		time.Sleep(time.Second)
		// clear screen
		_ = clearCMD("", u)
	}
}
