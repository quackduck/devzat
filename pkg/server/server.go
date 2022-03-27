package server

import (
	"crypto/sha256"
	"devzat/pkg"
	"devzat/pkg/commands/dm"
	"devzat/pkg/room"
	"devzat/pkg/user"
	_ "embed"
	"encoding/hex"
	"fmt"
	goaway "github.com/TwiN/go-away"
	"github.com/dghubble/go-twitter/twitter"
	"io"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"

	"devzat/pkg/bots/devbot"
	"devzat/pkg/colors"
	"github.com/acarl005/stripansi"
	"github.com/gliderlabs/ssh"
	terminal "github.com/quackduck/term"
)

const (
	defaultLogFileName      = "log.txt"
	defaultAsciiArtFileName = "art.txt"
)

const (
	Scrollback       = 16
	NumStartingBans  = 10
	DefaultUserCount = 10
	HugeTerminalSize = 10000
)

const (
	tooManyLogins = 6
)

type Server struct {
	Admins AdminInfoMap

	Port        int
	scrollback  int
	ProfilePort int

	// should this instance run offline? (should it not connect to slack or twitter?)
	OfflineSlack   bool
	OfflineTwitter bool

	MainRoom *room.Room
	Rooms    map[string]*room.Room

	Backlog          []pkg.BacklogMessage
	Bans             []Ban
	idsInMinToTimes  map[string]int
	AntiSpamMessages map[string]int

	Logfile io.WriteCloser
	Log     *log.Logger // prints to stdout as well as the logfile

	startupTime time.Time

	goaway.ProfanityDetector

	twitterClient *twitter.Client
	Commands      map[string]pkg.CommandFunc
}

func (s *Server) Init() error {
	s.initCommands()

	s.Port = 22
	s.scrollback = 16
	s.ProfilePort = 5555

	// should this instance run offline? (should it not connect to slack or twitter?)
	s.OfflineSlack = os.Getenv(pkg.EnvOfflineSlack) != ""
	s.OfflineTwitter = os.Getenv(pkg.EnvOfflineTwitter) != ""

	bot := &devbot.Bot{}

	s.MainRoom = &room.Room{
		Server:    s,
		Name:      "#MainRoom",
		Users:     make([]*user.User, 0, DefaultUserCount),
		Formatter: colors.NewFormatter(),
	}

	bot.SetRoom(s.MainRoom)
	if err := bot.Init(); err != nil {
		return fmt.Errorf("there was an error initializing the bot: %v", err)
	}

	s.Rooms = map[string]*room.Room{s.MainRoom.Name: s.MainRoom}

	s.Backlog = make([]pkg.BacklogMessage, 0, Scrollback)
	s.Bans = make([]Ban, 0, NumStartingBans)
	s.idsInMinToTimes = make(map[string]int) // TODO: maybe add some IP-based factor to disallow rapid key-gen attempts
	s.AntiSpamMessages = make(map[string]int)

	f, err := os.OpenFile(defaultLogFileName, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	s.Logfile = f
	s.Log = log.New(io.MultiWriter(s.Logfile, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)

	s.startupTime = time.Now()

	s.loadTwitterClient()

	return nil
}

func (s *Server) initCommands() {
	commands := []pkg.CommandRegistration{
		&dm.Command{},
	}

	commandMap := make(map[string]commands.CommandFunc)

	for _, c := range commands {
		commandMap[c.Name()] = c.Fn
	}
}

func (s *Server) UniverseBroadcast(senderName, msg string) {
	for _, r := range s.Rooms {
		r.Broadcast(senderName, msg)
	}
}

func (s *Server) autocompleteCallback(u *user.User, line string, pos int, key rune) (string, int, bool) {
	if key != '\t' {
		return "", pos, false
	}

	// Autocomplete a username

	// Split the input string to look for @<name>
	words := strings.Fields(line)

	toAdd := s.userMentionAutocomplete(u, words)
	if toAdd != "" {
		return line + toAdd, pos + len(toAdd), true
	}

	toAdd = s.roomAutocomplete(words)
	if toAdd != "" {
		return line + toAdd, pos + len(toAdd), true
	}

	//return line + toAdd + " ", pos + len(toAdd) + 1, true
	return "", pos, false
}

func (s *Server) userMentionAutocomplete(u *user.User, words []string) string {
	if len(words) < 1 {
		return ""
	}

	// Check the last word and see if it's trying to refer to a User
	if words[len(words)-1][0] == '@' || (len(words)-1 == 0 && words[0][0] == '=') { // mentioning someone or dm-ing someone
		inputWord := words[len(words)-1][1:] // slice the @ or = off

		for _, users := range u.Room.Users {
			strippedName := stripansi.Strip(users.Name)
			toAdd := strings.TrimPrefix(strippedName, inputWord)
			if toAdd != strippedName { // there was a match, and some text got trimmed!
				return toAdd + " "
			}
		}
	}

	return ""
}

func (s *Server) roomAutocomplete(words []string) string {
	// trying to refer to a room?
	if len(words) > 0 && words[len(words)-1][0] == '#' {
		// don't slice the # off, since the room name includes it
		for name := range s.Rooms {
			toAdd := strings.TrimPrefix(name, words[len(words)-1])
			if toAdd != name { // there was a match, and some text got trimmed!
				return toAdd + " "
			}
		}
	}

	return ""
}

func (s *Server) NewUser(session ssh.Session) (*user.User, error) {
	term := terminal.NewTerminal(session, "> ")

	pty, winChan, accepted := session.Pty()
	if !accepted {
		return nil, fmt.Errorf("PTY for ssh session not accepted: %v")
	}

	// disable any formatting done by term
	_ = term.SetSize(HugeTerminalSize, HugeTerminalSize)

	// definitely should not give an err
	host, _, _ := net.SplitHostPort(session.RemoteAddr().String())

	toHash := host // If we can't get the public key fall back to the IP.
	if pubKey := session.PublicKey(); pubKey != nil {
		toHash = string(pubKey.Marshal())
	}

	u := &user.User{
		Name:          "",
		Pronouns:      []string{"unset"},
		Session:       session,
		Term:          term,
		ID:            shasum(toHash),
		Addr:          host,
		Window:        pty.Window,
		LastTimestamp: time.Now(),
		JoinTime:      time.Now(),
		Room:          s.MainRoom,
	}

	u.Bell = true
	u.Color.Background = colors.NoBackground // the FG will be set randomly

	go func() {
		for u.Window = range winChan {
		}
	}()

	s.Log.Printf("Connected %v [%v]", u.Name, u.ID)

	if s.BansContains(u.Addr, u.ID) {
		banResponse := fmt.Sprintf(fmtDefaultBannedLoginResponse, u.ID)
		botName := s.MainRoom.Bot.Name()

		s.Log.Printf("Rejected %v [%v]", u.Name, host)
		u.Writeln(botName, banResponse)

		u.CloseQuietly()

		return nil, nil
	}

	s.idsInMinToTimes[u.ID]++
	time.AfterFunc(60*time.Second, func() {
		s.idsInMinToTimes[u.ID]--
	})

	if s.idsInMinToTimes[u.ID] > tooManyLogins {
		s.Bans = append(s.Bans, Ban{u.Addr, u.ID})
		msg := fmt.Sprintf("`%v` has been banned automatically. ID: %v", u.Name, u.ID)

		s.MainRoom.Broadcast(s.MainRoom.Bot.Name(), msg)

		return nil, nil
	}

	// always clear the screen on connect
	if err := s.Commands["Clear"]("", u); err != nil {
		return nil, err
	}

	s.handleValentinesDay(u)

	if len(s.Backlog) > 0 {
		lastStamp := s.Backlog[0].Time
		u.RWriteln(pkg.PrintPrettyDuration(u.JoinTime.Sub(lastStamp)) + " earlier")

		for i := range s.Backlog {
			if s.Backlog[i].Time.Sub(lastStamp) > time.Minute {
				lastStamp = s.Backlog[i].Time
				u.RWriteln(pkg.PrintPrettyDuration(u.JoinTime.Sub(lastStamp)) + " earlier")
			}
			u.Writeln(s.Backlog[i].SenderName, s.Backlog[i].Text)
		}
	}

	if err := u.PickUsernameQuietly(session.User()); err != nil { // User exited or had some error
		s.Log.Println(err)
		return nil, session.Close()
	}

	s.MainRoom.UsersMutex.Lock()
	s.MainRoom.Users = append(s.MainRoom.Users, u)
	go s.SendCurrentUsersTwitterMessage()
	s.MainRoom.UsersMutex.Unlock()

	u.Term.SetBracketedPasteMode(true) // experimental paste bracketing support
	term.AutoCompleteCallback = func(line string, pos int, key rune) (string, int, bool) {
		return s.autocompleteCallback(u, line, pos, key)
	}

	switch len(s.MainRoom.Users) - 1 {
	case 0:
		u.Writeln("", s.MainRoom.Colors.Blue.Paint("Welcome to the chat. There are no more users"))
	case 1:
		u.Writeln("", s.MainRoom.Colors.Yellow.Paint("Welcome to the chat. There is one more User"))
	default:
		u.Writeln("", s.MainRoom.Colors.Green.Paint("Welcome to the chat. There are", strconv.Itoa(len(s.MainRoom.Users)-1), "more users"))
	}

	botName := s.MainRoom.Bot.Name()
	s.MainRoom.Broadcast(botName, fmt.Sprintf("%s has joined the chat", u.Name))

	return u, nil
}

func (s *Server) ReplaceSlackEmoji(input string) string {
	if len(input) < 4 {
		return input
	}
	emojiName := ""
	result := make([]byte, 0, len(input))
	inEmojiName := false
	for i := 0; i < len(input)-1; i++ {
		if inEmojiName {
			emojiName += string(input[i]) // end result: if input contains "::lol::", emojiName will contain ":lol:". "::lol:: ::cat::" => ":lol::cat:"
		}
		if input[i] == ':' && input[i+1] == ':' {
			inEmojiName = !inEmojiName
		}
		//if !inEmojiName {
		result = append(result, input[i])
		//}
	}
	result = append(result, input[len(input)-1])
	if emojiName != "" {
		toAdd := s.fetchEmoji(strings.Split(strings.ReplaceAll(emojiName[1:len(emojiName)-1], "::", ":"), ":")) // cut the ':' at the start and end

		result = append(result, toAdd...)
	}
	return string(result)
}

// accepts a ':' separated list of emoji
func (s *Server) fetchEmoji(names []string) string {
	if s.OfflineSlack {
		return ""
	}
	result := ""
	for _, name := range names {
		result += s.fetchEmojiSingle(name)
	}
	return result
}

func (s *Server) fetchEmojiSingle(name string) string {
	if s.OfflineSlack {
		return ""
	}

	r, err := http.Get("https://e.benjaminsmith.dev/" + name)
	if err != nil {
		return ""
	}

	defer r.Body.Close()

	if r.StatusCode != 200 {
		return ""
	}

	return "![" + name + "](https://e.benjaminsmith.dev/" + name + ")"
}

// BansContains reports if the addr or id is found in the bans list
func (s *Server) BansContains(addr string, id string) bool {
	for i := 0; i < len(s.Bans); i++ {
		if s.Bans[i].Addr == addr || s.Bans[i].ID == id {
			return true
		}
	}

	return false
}

func (s *Server) BanUser(banner string, victim *user.User) {
	s.Bans = append(s.Bans, Ban{victim.Addr, victim.ID})
	_ = s.SaveBans()
	victim.Close(victim.Name + " has been banned by " + banner)
}

func (s *Server) handleValentinesDay(u *user.User) {
	if time.Now().Month() == time.February &&
		(time.Now().Day() == 14 || time.Now().Day() == 15 || time.Now().Day() == 13) {
		// TODO: add a few more random images
		u.Writeln("", "![❤️](https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/160/apple/81/heavy-black-heart_2764.png)")
		//u.term.Write([]byte("\u001B[A\u001B[2K\u001B[A\u001B[2K")) // delete last line of rendered markdown
		time.Sleep(time.Second)
		// clear screen
		_ = s.Commands["Clear"]("", u)
	}
}

func shasum(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
