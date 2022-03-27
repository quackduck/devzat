package pkg

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"

	"devchat/pkg/bots/devbot"
	"devchat/pkg/colors"
	"encoding/json"
	"github.com/acarl005/stripansi"
	"github.com/gliderlabs/ssh"
	terminal "github.com/quackduck/term"
)

const (
	scrollback       = 16
	numStartingBans  = 10
	defaultUserCount = 10
	hugeTerminalSize = 10000
)

const (
	defaultBansFile   = "bans.json"
	tooManyLogins     = 6
	defaultAdminsFile = "admins.json"
)

const (
	fmtDefaultBannedLoginResponse = `
		**You are banned**. 
		If you feel this was a mistake, please reach out at github.com/quackduck/devzat/issues 
		or email igoes.Log.mail@gmais.Log.com. 
	
		Please include the following information: [ID %v]
	`
)

type Server struct {
	Admins AdminInfoMap

	Port        int
	scrollback  int
	ProfilePort int

	// should this instance run offline? (should it not connect to slack or twitter?)
	OfflineSlack   bool
	OfflineTwitter bool

	MainRoom *Room // name of MainRoom room
	Rooms    map[string]*Room

	backlog          []BacklogMessage
	bans             []ban
	idsInMinToTimes  map[string]int
	antispamMessages map[string]int

	Logfile io.WriteCloser
	Log     *log.Logger // prints to stdout as well as the logfile

	startupTime time.Time
}

func (s *Server) Init() error {
	s.Port = 22
	s.scrollback = 16
	s.ProfilePort = 5555

	// should this instance run offline? (should it not connect to slack or twitter?)
	s.OfflineSlack = os.Getenv(EnvOfflineSlack) != ""
	s.OfflineTwitter = os.Getenv(EnvOfflineTwitter) != ""

	bot := &devbot.Bot{}

	s.MainRoom = &Room{
		Server:    s,
		Name:      "#MainRoom",
		users:     make([]*User, 0, defaultUserCount),
		Formatter: colors.NewFormatter(),
	}

	bot.SetRoom(s.MainRoom)
	if err := bot.Init(); err != nil {
		return fmt.Errorf("there was an error initializing the bot: %v", err)
	}

	s.Rooms = map[string]*Room{s.MainRoom.Name: s.MainRoom}

	s.backlog = make([]BacklogMessage, 0, scrollback)
	s.bans = make([]ban, 0, numStartingBans)
	s.idsInMinToTimes = make(map[string]int) // TODO: maybe add some IP-based factor to disallow rapid key-gen attempts
	s.antispamMessages = make(map[string]int)

	f, err := os.OpenFile("log.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}

	s.Logfile = f
	s.Log = log.New(io.MultiWriter(s.Logfile, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)

	s.startupTime = time.Now()

	return nil
}

func (s *Server) SaveBans() error {
	f, err := os.Create(defaultBansFile)
	if err != nil {
		s.Log.Println(err)
		return fmt.Errorf("could not create bans file: %v", err)
	}

	j := json.NewEncoder(f)
	j.SetIndent("", "   ")

	if err = j.Encode(numStartingBans); err != nil {
		s.Rooms["#MainRoom"].Broadcast(s.MainRoom.Bot.Name(), "error saving bans: "+err.Error())
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

	err = json.NewDecoder(f).Decode(&s.bans)
	if err != nil {
		msg := fmt.Sprintf("error reading bans: %v", err)
		botName := s.MainRoom.Bot.Name()

		s.MainRoom.Broadcast(botName, msg)
		s.Log.Println(msg)

		return
	}

	_ = f.Close()
}

func (s *Server) UniverseBroadcast(senderName, msg string) {
	for _, r := range s.Rooms {
		r.Broadcast(senderName, msg)
	}
}

func (s *Server) autocompleteCallback(u *User, line string, pos int, key rune) (string, int, bool) {
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

func (s *Server) userMentionAutocomplete(u *User, words []string) string {
	if len(words) < 1 {
		return ""
	}

	// Check the last word and see if it's trying to refer to a User
	if words[len(words)-1][0] == '@' || (len(words)-1 == 0 && words[0][0] == '=') { // mentioning someone or dm-ing someone
		inputWord := words[len(words)-1][1:] // slice the @ or = off

		for _, users := range u.Room.users {
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

func (s *Server) NewUser(session ssh.Session) (*User, error) {
	term := terminal.NewTerminal(session, "> ")

	pty, winChan, accepted := session.Pty()
	if !accepted {
		return nil, fmt.Errorf("PTY for ssh session not accepted: %v")
	}

	// disable any formatting done by term
	_ = term.SetSize(hugeTerminalSize, hugeTerminalSize)

	// definitely should not give an err
	host, _, _ := net.SplitHostPort(session.RemoteAddr().String())

	toHash := host // If we can't get the public key fall back to the IP.
	if pubKey := session.PublicKey(); pubKey != nil {
		toHash = string(pubKey.Marshal())
	}

	u := &User{
		Name:          "",
		Pronouns:      []string{"unset"},
		Session:       session,
		Term:          term,
		id:            shasum(toHash),
		addr:          host,
		Window:        pty.Window,
		lastTimestamp: time.Now(),
		joinTime:      time.Now(),
		Room:          s.MainRoom,
	}

	u.bell = true
	u.Color.Background = colors.NoBackground // the FG will be set randomly

	go func() {
		for u.Window = range winChan {
		}
	}()

	s.Log.Printf("Connected %v [%v]", u.Name, u.id)

	if s.bansContains(u.addr, u.id) {
		banResponse := fmt.Sprintf(fmtDefaultBannedLoginResponse, u.id)
		botName := s.MainRoom.Bot.Name()

		s.Log.Printf("Rejected %v [%v]", u.Name, host)
		u.Writeln(botName, banResponse)

		u.CloseQuietly()

		return nil, nil
	}

	s.idsInMinToTimes[u.id]++
	time.AfterFunc(60*time.Second, func() {
		s.idsInMinToTimes[u.id]--
	})

	if s.idsInMinToTimes[u.id] > tooManyLogins {
		s.bans = append(s.bans, ban{u.addr, u.id})
		msg := fmt.Sprintf("`%v` has been banned automatically. ID: %v", u.Name, u.id)

		s.MainRoom.Broadcast(s.MainRoom.Bot.Name(), msg)

		return nil, nil
	}

	// always clear the screen on connect
	if err := clearCMD("", u); err != nil {
		return nil, err
	}

	u.handleValentinesDay()

	if len(s.backlog) > 0 {
		lastStamp := s.backlog[0].timestamp
		u.RWriteln(printPrettyDuration(u.joinTime.Sub(lastStamp)) + " earlier")

		for i := range s.backlog {
			if s.backlog[i].timestamp.Sub(lastStamp) > time.Minute {
				lastStamp = s.backlog[i].timestamp
				u.RWriteln(printPrettyDuration(u.joinTime.Sub(lastStamp)) + " earlier")
			}
			u.Writeln(s.backlog[i].senderName, s.backlog[i].text)
		}
	}

	if err := u.PickUsernameQuietly(session.User()); err != nil { // User exited or had some error
		s.Log.Println(err)
		return nil, session.Close()
	}

	s.MainRoom.usersMutex.Lock()
	s.MainRoom.users = append(s.MainRoom.users, u)
	go sendCurrentUsersTwitterMessage()
	s.MainRoom.usersMutex.Unlock()

	u.Term.SetBracketedPasteMode(true) // experimental paste bracketing support
	term.AutoCompleteCallback = func(line string, pos int, key rune) (string, int, bool) {
		return s.autocompleteCallback(u, line, pos, key)
	}

	switch len(s.MainRoom.users) - 1 {
	case 0:
		u.Writeln("", s.MainRoom.Colors.Blue.Paint("Welcome to the chat. There are no more users"))
	case 1:
		u.Writeln("", s.MainRoom.Colors.Yellow.Paint("Welcome to the chat. There is one more User"))
	default:
		u.Writeln("", s.MainRoom.Colors.Green.Paint("Welcome to the chat. There are", strconv.Itoa(len(s.MainRoom.users)-1), "more users"))
	}

	botName := s.MainRoom.Bot.Name()
	s.MainRoom.Broadcast(botName, fmt.Sprintf("%s has joined the chat", u.Name))

	return u, nil
}

func (s *Server) replaceSlackEmoji(input string) string {
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

// bansContains reports if the addr or id is found in the bans list
func (s *Server) bansContains(addr string, id string) bool {
	for i := 0; i < len(s.bans); i++ {
		if s.bans[i].Addr == addr || s.bans[i].ID == id {
			return true
		}
	}

	return false
}

func (s *Server) getAdmins() (map[AdminID]AdminIndo, error) {
	if _, err := os.Stat(defaultAdminsFileName); err == os.ErrNotExist {
		return nil, errors.New("make an admins.json file to add admins")
	}

	data, err := ioutil.ReadFile("admins.json")
	if err != nil {
		return nil, fmt.Errorf("error reading admins.json: %s", err)
	}

	adminsList := make(AdminInfoMap)

	if err = json.Unmarshal(data, &adminsList); err != nil {
		return nil, fmt.Errorf("bad json: %v", err)
	}

	return adminsList, nil
}
