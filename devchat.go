package main

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	_ "embed"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

	markdown "github.com/MichaelMure/go-term-markdown"
	"github.com/acarl005/stripansi"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	ttt "github.com/shurcooL/tictactoe"
	"github.com/slack-go/slack"
	terminal "golang.org/x/term"
)

var (
	//go:embed slackAPI.txt
	slackAPI []byte
	//go:embed adminPass.txt
	adminPass []byte
	//go:embed art.txt
	artBytes   []byte
	port       = 22
	scrollback = 16

	slackChan = getSendToSlackChan()
	api       = slack.New(string(slackAPI))
	rtm       = api.NewRTM()

	red      = color.New(color.FgHiRed)
	green    = color.New(color.FgHiGreen)
	cyan     = color.New(color.FgHiCyan)
	magenta  = color.New(color.FgHiMagenta)
	yellow   = color.New(color.FgHiYellow)
	blue     = color.New(color.FgHiBlue)
	black    = color.New(color.FgHiBlack)
	white    = color.New(color.FgHiWhite)
	colorArr = []*color.Color{yellow, cyan, magenta, green, white, blue, black, red}

	devbot = ""

	users      = make([]*user, 0, 10)
	usersMutex = sync.Mutex{}

	allUsers      = make(map[string]string, 100) //map format is u.id => u.name
	allUsersMutex = sync.Mutex{}

	backlog      = make([]message, 0, scrollback)
	backlogMutex = sync.Mutex{}

	logfile, _ = os.OpenFile("log.txt", os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	l          = log.New(io.MultiWriter(logfile, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)

	bans      = make([]string, 0, 10)
	bansMutex = sync.Mutex{}

	// stores the ids which have joined in 20 seconds and how many times this happened
	idsIn20ToTimes = make(map[string]int, 10)
	idsIn20Mutex   = sync.Mutex{}

	tttGame       = new(ttt.Board)
	currentPlayer = ttt.X

	hangGame = new(hangman)

	admins = []string{"d84447e08901391eb36aa8e6d9372b548af55bee3799cd3abb6cdd503fdf2d82"}
)

// TODO: devbot hangman game
// TODO: email people on ping word idea
func main() {
	color.NoColor = false
	devbot = green.Sprint("devbot")
	var err error
	rand.Seed(time.Now().Unix())
	readBansAndUsers()
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL)
	go func() {
		<-c
		fmt.Println("Shutting down...")
		saveBansAndUsers()
		logfile.Close()
		broadcast(devbot, "Server going down! This is probably because it is being updated. Try joining back immediately.  \n"+
			"If you still can't join, try joining back in 2 minutes. If you _still_ can't join, make an issue at github.com/quackduck/devzat/issues", true)
		os.Exit(0)
	}()

	ssh.Handle(func(s ssh.Session) {
		u := newUser(s)
		if u == nil {
			return
		}
		u.repl()
	})
	if os.Getenv("PORT") != "" {
		port, err = strconv.Atoi(os.Getenv("PORT"))
		if err != nil {
			fmt.Println(err)
			return
		}
	}

	fmt.Println(fmt.Sprintf("Starting chat server on port %d", port))
	go getMsgsFromSlack()
	go func() {
		if port == 22 {
			fmt.Println("Also starting chat server on port 443")
			err = ssh.ListenAndServe(
				":443",
				nil,
				ssh.HostKeyFile(os.Getenv("HOME")+"/.ssh/id_rsa"))
		}
	}()
	err = ssh.ListenAndServe(
		fmt.Sprintf(":%d", port),
		nil,
		ssh.HostKeyFile(os.Getenv("HOME")+"/.ssh/id_rsa"))
	if err != nil {
		fmt.Println(err)
	}
}

func broadcast(senderName, msg string, toSlack bool) {
	if msg == "" {
		return
	}
	backlogMutex.Lock()
	backlog = append(backlog, message{senderName, msg + "\n"})
	backlogMutex.Unlock()
	if toSlack {
		if senderName != "" {
			slackChan <- senderName + ": " + msg
		} else {
			slackChan <- msg
		}
	}
	for len(backlog) > scrollback { // for instead of if just in case
		backlog = backlog[1:]
	}
	for i := range users {
		users[i].writeln(senderName, msg)
	}
}

type user struct {
	name      string
	session   ssh.Session
	term      *terminal.Terminal
	bell      bool
	color     color.Color
	id        string
	addr      string
	win       ssh.Window
	closeOnce sync.Once
}

type message struct {
	senderName string
	text       string
}

type hangman struct {
	word      string
	triesLeft int
	guesses   string // string containing all the guessed characters
}

func newUser(s ssh.Session) *user {
	term := terminal.NewTerminal(s, "> ")
	_ = term.SetSize(10000, 10000) // disable any formatting done by term
	pty, winchan, _ := s.Pty()
	w := pty.Window
	host, _, err := net.SplitHostPort(s.RemoteAddr().String()) // definitely should not give an err
	if err != nil {
		term.Write([]byte(fmt.Sprintln(err) + "\n"))
		s.Close()
		return nil
	}
	hash := sha256.New()
	hash.Write([]byte(host))
	u := &user{s.User(), s, term, true, color.Color{}, hex.EncodeToString(hash.Sum(nil)), host, w, sync.Once{}}
	go func() {
		for u.win = range winchan {
		}
	}()
	l.Println("Connected " + u.name + " [" + u.id + "]")
	for i := range bans {
		if u.addr == bans[i] || u.id == bans[i] { // allow banning by ID
			if u.id == bans[i] { // then replace the ID in the ban with the actual IP
				bans[i] = u.addr
				saveBansAndUsers()
			}
			l.Println("Rejected " + u.name + " [" + u.addr + "]")
			u.writeln(devbot, "**You are banned**. If you feel this was done wrongly, please reach out at github.com/quackduck/devzat/issues. Please include the following information: [IP "+u.addr+"]")
			u.close("")
			return nil
		}
	}
	idsIn20Mutex.Lock()
	idsIn20ToTimes[u.id]++
	idsIn20Mutex.Unlock()
	time.AfterFunc(30*time.Second, func() {
		idsIn20Mutex.Lock()
		idsIn20ToTimes[u.id]--
		idsIn20Mutex.Unlock()
	})
	if idsIn20ToTimes[u.id] > 3 { // 10 minute ban
		bansMutex.Lock()
		bans = append(bans, u.addr)
		bansMutex.Unlock()
		broadcast(devbot, u.name+" has been banned automatically. IP: "+u.addr, true)
		return nil
	}
	u.pickUsername(s.User())
	usersMutex.Lock()
	users = append(users, u)
	usersMutex.Unlock()
	switch len(users) - 1 {
	case 0:
		u.writeln("", "**"+cyan.Sprint("Welcome to the chat. There are no more users")+"**")
	case 1:
		u.writeln("", "**"+cyan.Sprint("Welcome to the chat. There is one more user")+"**")
	default:
		u.writeln("", "**"+cyan.Sprint("Welcome to the chat. There are ", len(users)-1, " more users")+"**")
	}
	//_, _ = term.Write([]byte(strings.Join(backlog, ""))) // print out backlog
	for i := range backlog {
		u.writeln(backlog[i].senderName, backlog[i].text)
	}
	broadcast(devbot, "**"+u.name+"** **"+green.Sprint("has joined the chat")+"**", true)
	return u
}

func (u *user) close(msg string) {
	u.closeOnce.Do(func() {
		usersMutex.Lock()
		users = remove(users, u)
		usersMutex.Unlock()
		broadcast(devbot, msg, true)
		u.session.Close()
	})
}

func (u *user) writeln(senderName string, msg string) {
	msg = strings.ReplaceAll(msg, `\n`, "\n")
	msg = strings.ReplaceAll(msg, `\`+"\n", `\n`) // let people escape newlines
	if senderName != "" {
		msg = strings.TrimSpace(mdRender(msg, len([]rune(stripansi.Strip(senderName))), u.win.Width))
		msg = senderName + ": " + msg
	} else {
		msg = strings.TrimSpace(mdRender(msg, -2, u.win.Width)) // -2 so linewidth is used as is
	}
	if u.bell {
		u.term.Write([]byte(msg + "\a\n")) // "\a" is beep
	} else {
		u.term.Write([]byte(msg + "\n"))
	}
}

func (u *user) pickUsername(possibleName string) {
	possibleName = cleanName(possibleName)
	var err error
	for userDuplicate(possibleName) || possibleName == "" || possibleName == "devbot" {
		u.writeln("", "Pick a different username")
		u.term.SetPrompt("> ")
		possibleName, err = u.term.ReadLine()
		if err != nil {
			l.Println(err)
			return
		}
		possibleName = cleanName(possibleName)
	}
	u.name = possibleName
	u.changeColor(*colorArr[rand.Intn(len(colorArr))])
	allUsersMutex.Lock()
	if _, ok := allUsers[u.id]; !ok {
		broadcast(devbot, "You seem to be new here "+u.name+". Welcome to Devzat! Run /help to see what you can do.", true)
	}
	allUsers[u.id] = u.name
	allUsersMutex.Unlock()
	saveBansAndUsers()
}

func (u *user) changeColor(color color.Color) {
	u.name = color.Sprint(stripansi.Strip(u.name))
	u.color = color
	u.term.SetPrompt(u.name + ": ")
}

func (u *user) repl() {
	for {
		line, err := u.term.ReadLine()
		line = strings.TrimSpace(line)

		if err == io.EOF {
			u.close("**" + u.name + "** **" + red.Sprint("has left the chat") + "**")
			return
		}
		if err != nil {
			l.Println(u.name, err)
			continue
		}
		u.term.Write([]byte(strings.Repeat("\033[A\033[2K", int(math.Ceil(float64(len([]rune(u.name+line))+2)/(float64(u.win.Width))))))) // basically, ceil(length of line divided by term width)

		runCommands(line, u, false)
	}
}

func runCommands(line string, u *user, isSlack bool) {
	if line == "" {
		u.writeln("", "An empty message? Send some content!")
		return
	}

	toSlack := true
	if strings.HasPrefix(line, "/hide") && !isSlack {
		toSlack = false
	}
	if strings.HasPrefix(line, "/dm") && !isSlack {
		toSlack = false
		rest := strings.TrimSpace(strings.TrimPrefix(line, "/dm"))
		restSplit := strings.Fields(rest)
		if len(restSplit) < 2 {
			u.writeln("", "Not enough arguments to /dm. Use /dm <user> <msg>")
			return
		}
		peer, ok := findUserByName(restSplit[0])
		if !ok {
			u.writeln("", "User not found")
			return
		}
		msg := strings.TrimSpace(strings.TrimPrefix(rest, restSplit[0]))
		u.writeln(u.name+" -> "+peer.name, msg)
		//peer.writeln(u.name+" -> "+peer.name, msg)
		if u == peer {
			u.writeln(devbot, "You must be really lonely, DMing yourself. Don't worry, I won't judge :wink:")
		} else {
			//peer.writeln(peer.name+" <- "+u.name, msg)
			peer.writeln(u.name+" -> "+peer.name, msg)
		}
		return
	}
	if strings.HasPrefix(line, "/hang") {
		rest := strings.TrimSpace(strings.TrimPrefix(line, "/hang"))
		if len(rest) > 1 {
			u.writeln(u.name, line)
			hangGame = &hangman{rest, 15, " "} // default value of guesses so empty space is given away
			broadcast(devbot, u.name+" has started a new game of Hangman! Guess letters with /hang <letter>", toSlack)
			broadcast(devbot, "```\n"+hangPrint(hangGame)+"\nTries: "+strconv.Itoa(hangGame.triesLeft)+"\n```", toSlack)
			return
		} else { // allow message to show to everyone
			if !isSlack {
				broadcast(u.name, line, toSlack)
			}
		}
		if strings.Trim(hangGame.word, hangGame.guesses) == "" {
			broadcast(devbot, "The game has ended. Start a new game with /hang <word>", toSlack)
			return
		}
		if len(rest) == 0 {
			broadcast(devbot, "Start a new game with /hang <word> or guess with /hang <letter>", toSlack)
			return
		}
		if hangGame.triesLeft == 0 {
			broadcast(devbot, "No more tries! The word was "+hangGame.word, toSlack)
			return
		}
		if strings.Contains(hangGame.guesses, rest) {
			broadcast(devbot, "You already guessed "+rest, toSlack)
			return
		} else {
                        hangGame.guesses += rest
                }
		if !(strings.Contains(hangGame.word, rest)) {
			hangGame.triesLeft--
		}

		display := hangPrint(hangGame)
		broadcast(devbot, "```\n"+display+"\nTries: "+strconv.Itoa(hangGame.triesLeft)+"\n```", toSlack)

		if strings.Trim(hangGame.word, hangGame.guesses) == "" {
			broadcast(devbot, "You got it! The word was "+hangGame.word, toSlack)
		} else if hangGame.triesLeft == 0 {
			broadcast(devbot, "No more tries! The word was "+hangGame.word, toSlack)
		}
		return
	}
	if !isSlack {
		broadcast(u.name, line, toSlack)
	}

	if strings.Contains(line, "devbot") {
		devbotMessages := []string{"Hi I'm devbot", "Hey", "HALLO :rocket:", "Yes?", "I'm in the middle of something can you not", "Devbot to the rescue!", "Run /help, you need it."}
		if strings.Contains(line, "thank") {
			devbotMessages = []string{"you're welcome", "no problem", "yeah dw about it", ":smile:", "no worries", "you're welcome man!"}
		}
		pick := devbotMessages[rand.Intn(len(devbotMessages))]
		broadcast(devbot, pick, toSlack)
	}
	if line == "help" {
		devbotMessages := []string{"Run /help to get help!", "Looking for /help?", "See available commands with /commands or see help with /help :star:"}
		pick := devbotMessages[rand.Intn(len(devbotMessages))]
		broadcast(devbot, pick, toSlack)
		return
	}

	if strings.Contains(line, "rocket") {
		devbotMessages := []string{"Doge to the mooooon :rocket:", ":rocket:", "I like rockets", "SpaceX", ":dog2:"}
		pick := devbotMessages[rand.Intn(len(devbotMessages))]
		broadcast(devbot, pick, toSlack)
		return
	}

	if strings.Contains(line, "elon") {
		devbotMessages := []string{"When something is important enough, you do it even if the odds are not in your favor.", "I do think there is a lot of potential if you have a compelling product", "If you're trying to create a company, it's like baking a cake. You have to have all the ingredients in the right proportion.", "Patience is a virtue, and I'm learning patience. It's a tough lesson."}
		pick := devbotMessages[rand.Intn(len(devbotMessages))]
		broadcast(devbot, "> "+pick+"\n\n\\- Elon Musk", toSlack)
		return
	}

	if strings.Contains(line, "star") {
		devbotMessages := []string{"Someone say :star:? If you like Devzat, do give it a star at github.com/quackduck/devzat!", "You should :star: github.com/quackduck/devzat", ":star:"}
		pick := devbotMessages[rand.Intn(len(devbotMessages))]
		broadcast(devbot, pick, toSlack)
	}
	if strings.Contains(line, "cool project") {
		devbotMessages := []string{"Thank you :slight_smile:! If you like Devzat, do give it a star at github.com/quackduck/devzat!", "Star Devzat here: github.com/quackduck/devzat"}
		pick := devbotMessages[rand.Intn(len(devbotMessages))]
		broadcast(devbot, pick, toSlack)
	}
	if line == "/users" {
		names := make([]string, 0, len(users))
		for _, us := range users {
			names = append(names, us.name)
		}
		broadcast("", fmt.Sprint(names), toSlack)
		return
	}
	if line == "/all" {
		names := make([]string, 0, len(allUsers))
		for _, name := range allUsers {
			names = append(names, name)
		}
		//sort.Strings(names)
		sort.Slice(names, func(i, j int) bool {
			return strings.ToLower(stripansi.Strip(names[i])) < strings.ToLower(stripansi.Strip(names[j]))
		})
		broadcast("", fmt.Sprint(names), toSlack)
		return
	}
	if line == "easter" {
		broadcast(devbot, "eggs?", toSlack)
		return
	}
	if line == "/exit" && !isSlack {
		u.close("**" + u.name + "** **" + red.Sprint("has left the chat") + "**")
		return
	}
	if line == "/bell" && !isSlack {
		u.bell = !u.bell
		if u.bell {
			broadcast("", fmt.Sprint("bell on"), toSlack)
		} else {
			broadcast("", fmt.Sprint("bell off"), toSlack)
		}
		return
	}
	if strings.HasPrefix(line, "/id") {
		victim, ok := findUserByName(strings.TrimSpace(strings.TrimPrefix(line, "/id")))
		if !ok {
			broadcast("", "User not found", toSlack)
			return
		}
		broadcast("", victim.id, toSlack)
		return
	}
	if strings.HasPrefix(line, "/nick") && !isSlack {
		u.pickUsername(strings.TrimSpace(strings.TrimPrefix(line, "/nick")))
		return
	}
	if strings.HasPrefix(line, "/banIP") && !isSlack {
		if !auth(u) {
			return
		}
		bansMutex.Lock()
		bans = append(bans, strings.TrimSpace(strings.TrimPrefix(line, "/banIP")))
		bansMutex.Unlock()
		saveBansAndUsers()
		return
	}

	if strings.HasPrefix(line, "/ban") && !isSlack {
		victim, ok := findUserByName(strings.TrimSpace(strings.TrimPrefix(line, "/ban")))
		if !ok {
			broadcast("", "User not found", toSlack)
			return
		}
		if !auth(u) {
			return
		}
		bansMutex.Lock()
		bans = append(bans, victim.addr)
		bansMutex.Unlock()
		saveBansAndUsers()
		victim.close(victim.name + " has been banned by " + u.name)

	}
	if strings.HasPrefix(line, "/kick") && !isSlack {
		victim, ok := findUserByName(strings.TrimSpace(strings.TrimPrefix(line, "/kick")))
		if !ok {
			broadcast("", "User not found", toSlack)
			return
		}
		if !auth(u) {
			return
		}
		victim.close(victim.name + red.Sprint(" has been kicked by ") + u.name)
	}
	if strings.HasPrefix(line, "/color") && !isSlack {
		colorMsg := "Which color? Choose from green, cyan, blue, red/orange, magenta/purple/pink, yellow/beige, white/cream and black/gray/grey.  \nThere's also a few secret colors :)"
		switch strings.TrimSpace(strings.TrimPrefix(line, "/color")) {
		case "green":
			u.changeColor(*green)
		case "cyan":
			u.changeColor(*cyan)
		case "blue":
			u.changeColor(*blue)
		case "red", "orange":
			u.changeColor(*red)
		case "magenta", "purple", "pink":
			u.changeColor(*magenta)
		case "yellow", "beige":
			u.changeColor(*yellow)
		case "white", "cream":
			u.changeColor(*white)
		case "black", "gray", "grey":
			u.changeColor(*black)
			// secret colors
		case "easter":
			u.changeColor(*color.New(color.BgMagenta, color.FgHiYellow))
		case "baby":
			u.changeColor(*color.New(color.BgBlue, color.FgHiMagenta))
		case "l33t":
			u.changeColor(*u.color.Add(color.BgHiBlack))
		case "whiten":
			u.changeColor(*u.color.Add(color.BgWhite))
		case "hacker":
			u.changeColor(*color.New(color.FgHiGreen, color.BgBlack))
		default:
			broadcast(devbot, colorMsg, toSlack)
		}
		return
	}
	if line == "/people" {
		broadcast("", `
**Hack Club members**  
Zach Latta     - Founder of Hack Club  
Zachary Fogg   - Hack Club Game Designer  
Matthew        - Hack Club HQ  
Caleb Denio, Safin Singh, Eleeza A  
Jubril, Sarthak Mohanty, Anghe,  
Tommy Pujol, Sam Poder, Rishi Kothari  
Amogh Chaubey, Ella Xu, Hugo Hu  
_Possibly more people_


**From my school:**  
Kiyan, Riya, Georgie  
Rayed Hamayun, Aarush Kumar


**From Twitter:**  
Ayush Pathak    @ayshptk  
Bereket         @heybereket  
Srushti         @srushtiuniverse  
Surjith         @surjithctly  
Arav Narula     @HeyArav  
Krish Nerkar    @krishnerkar_  
Amrit           @astro_shenava  
Mudrank Gupta   @mudrankgupta

**And many more have joined!**`, toSlack)
		return
	}

	if strings.HasPrefix(line, "/tic") {
		rest := strings.TrimSpace(strings.TrimPrefix(line, "/tic"))
		if rest == "" {
			broadcast(devbot, "Starting a new game of Tic Tac Toe! The first player is always X.", toSlack)
			currentPlayer = ttt.X
			tttGame = new(ttt.Board)
			//broadcast(devbot, "```\n"+"0│1│2\n3"+"\n```", toSlack)
			broadcast(devbot, "```\n"+tttPrint(tttGame.Cells)+"\n```", toSlack)
			return
		}
		m, err := strconv.Atoi(rest)
		if err != nil {
			broadcast(devbot, "There's something wrong with that command :thinking:", toSlack)
			return
		}
		if m < 1 || m > 9 {
			broadcast(devbot, "Moves are numbers between 1 and 9!", toSlack)
			return
		}
		err = tttGame.Apply(ttt.Move(m-1), currentPlayer)
		if err != nil {
			broadcast(devbot, err.Error(), toSlack)
			return
		}
		broadcast(devbot, "```\n"+tttPrint(tttGame.Cells)+"\n```", toSlack)
		if currentPlayer == ttt.X {
			currentPlayer = ttt.O
		} else {
			currentPlayer = ttt.X
		}
		if !(tttGame.Condition() == ttt.NotEnd) {
			broadcast(devbot, tttGame.Condition().String(), toSlack)
			currentPlayer = ttt.X
			tttGame = new(ttt.Board)
		}
		return
	}

	if line == "/help" {
		broadcast("", `Welcome to Devzat! Devzat is chat over SSH: github.com/quackduck/devzat  
Because there's SSH apps on all platforms, even on mobile, you can join from anywhere.

Interesting features:
* Many, many commands. Check em out by using /commands.
* Tic Tac Toe! Run /tic
* Markdown support! Tables, headers, italics and everything. Just use "\\n" in place of newlines.  
   You can even send _ascii art_ with code fences. Run /ascii-art to see an example.
* Emoji replacements :fire:! \:rocket\: => :rocket: (like on Slack and Discord)
* Code syntax highlighting. Use Markdown fences to send code. Run /example-code to see an example.

For replacing newlines, I often use bulkseotools.com/add-remove-line-breaks.php.

Made by Ishan Goel with feature ideas from friends.  
Thanks to Caleb Denio for lending his server!`, toSlack)
		return
	}
	if line == "/example-code" {
		broadcast(devbot, "\n```go\npackage main\nimport \"fmt\"\nfunc main() {\n   fmt.Println(\"Example!\")\n}\n```", toSlack)
		return
	}
	if line == "/ascii-art" {
		broadcast("", string(artBytes), toSlack)
		return
	}
	if line == "/emojis" {
		broadcast(devbot, "Check out github.com/ikatyang/emoji-cheat-sheet", toSlack)
		return
	}
	if line == "/commands" {
		broadcast("", `**Available commands**  
   **/dm**    <user> <msg>   _Privately message people_  
   **/users**                _List users_  
   **/nick**  <name>         _Change your name_  
   **/color** <color>        _Change your name color_  
   **/tic**   <move>         _Play Tic Tac Toe!_  
   **/all**                  _Get a list of all unique users ever_  
   **/emojis**               _See a list of emojis_  
   **/people**               _See info about nice people who joined_  
   **/exit**                 _Leave the chat_  
   **/hide**                 _Hide messages from HC Slack_  
   **/bell**                 _Toggle the ansi bell_  
   **/id**    <user>         _Get a unique identifier for a user_  
   **/ban**   <user>         _Ban a user, requires an admin pass_  
   **/kick**  <user>         _Kick a user, requires an admin pass_  
   **/help**                 _Show help_  
   **/commands**             _Show this message_`, toSlack)
	}
}

func hangPrint(hangGame *hangman) string {
	display := ""
	for _, c := range []rune(hangGame.word) {
		if strings.ContainsRune(hangGame.guesses, c) {
			display += string(c)
		} else {
			display += "_"
		}
	}
	return display
}

func tttPrint(cells [9]ttt.State) string {
	strcells := new([9]string)
	for i := range cells {
		//if cells[i].String() == " " {
		//	strcells[i] = supsub.ToSub(strconv.Itoa(i + 1))
		//	continue
		//}
		strcells[i] = cells[i].String()
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, " %v │ %v │ %v \n", strcells[0], strcells[1], strcells[2])
	fmt.Fprintln(&buf, "───┼───┼───")
	fmt.Fprintf(&buf, " %v │ %v │ %v \n", strcells[3], strcells[4], strcells[5])
	fmt.Fprintln(&buf, "───┼───┼───")
	fmt.Fprintf(&buf, " %v │ %v │ %v ", strcells[6], strcells[7], strcells[8])
	return buf.String()
}

func auth(u *user) bool {
	for _, id := range admins {
		if u.id == id || u.addr == id {
			return true
		}
	}
	return false
}

func cleanName(name string) string {
	var s string
	s = ""
	name = strings.TrimSpace(name)
	name = strings.Split(name, "\n")[0] // use only one line
	for _, r := range name {
		if unicode.IsGraphic(r) {
			s += string(r)
		}
	}
	return s
}

func mdRender(a string, nameLen int, lineWidth int) string {
	md := string(markdown.Render(a, lineWidth-(nameLen+2), 0))
	md = strings.TrimSuffix(md, "\n")
	split := strings.Split(md, "\n")
	for i := range split {
		if i == 0 {
			continue // the first line will automatically be padded
		}
		split[i] = strings.Repeat(" ", nameLen+2) + split[i]
	}
	if len(split) == 1 {
		return md
	}
	return strings.Join(split, "\n")
}

// Returns true if the username is taken, false otherwise
func userDuplicate(a string) bool {
	for i := range users {
		if stripansi.Strip(users[i].name) == stripansi.Strip(a) {
			return true
		}
	}
	return false
}

func saveBansAndUsers() {
	f, err := os.Create("allusers.json")
	if err != nil {
		l.Println(err)
		return
	}
	j := json.NewEncoder(f)
	j.SetIndent("", "   ")
	j.Encode(allUsers)
	f.Close()

	f, err = os.Create("bans.json")
	if err != nil {
		l.Println(err)
		return
	}
	j = json.NewEncoder(f)
	j.SetIndent("", "   ")
	j.Encode(bans)
	f.Close()
}

func readBansAndUsers() {
	f, err := os.Open("allusers.json")
	if err != nil {
		l.Println(err)
		return
	}
	allUsersMutex.Lock()
	json.NewDecoder(f).Decode(&allUsers)
	allUsersMutex.Unlock()
	f.Close()

	f, err = os.Open("bans.json")
	if err != nil {
		l.Println(err)
		return
	}
	bansMutex.Lock()
	json.NewDecoder(f).Decode(&bans)
	bansMutex.Unlock()
	f.Close()
}

func getMsgsFromSlack() {
	go rtm.ManageConnection()
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			msg := ev.Msg
			if msg.SubType != "" {
				break // We're only handling normal messages.
			}
			u, _ := api.GetUserInfo(msg.User)
			if !strings.HasPrefix(msg.Text, "hide") {
				h := sha1.Sum([]byte(u.ID))
				i, _ := binary.Varint(h[:])
				broadcast(color.HiYellowString("HC ")+(*colorArr[rand.New(rand.NewSource(i)).Intn(len(colorArr))]).Sprint(strings.Fields(u.RealName)[0]), msg.Text, false)
				runCommands(msg.Text, nil, true)
			}
		case *slack.ConnectedEvent:
			fmt.Println("Connected to Slack")
		case *slack.InvalidAuthEvent:
			fmt.Println("Invalid token")
			return
		}
	}
}

func getSendToSlackChan() chan string {
	msgs := make(chan string, 100)
	go func() {
		for msg := range msgs {
			//if strings.HasPrefix(msg, "HC: ") { // just in case
			//	continue
			//}
			msg = strings.ReplaceAll(stripansi.Strip(msg), `\n`, "\n")
			rtm.SendMessage(rtm.NewOutgoingMessage(msg, "C01T5J557AA"))
		}
	}()
	return msgs
}

func findUserByName(name string) (*user, bool) {
	usersMutex.Lock()
	defer usersMutex.Unlock()
	for _, u := range users {
		if stripansi.Strip(u.name) == name {
			return u, true
		}
	}
	return nil, false
}

func remove(s []*user, a *user) []*user {
	var i int
	for i = range s {
		if s[i] == a {
			break // i is now where it is
		}
	}
	if i == 0 {
		return make([]*user, 0)
	}
	return append(s[:i], s[i+1:]...)
}
