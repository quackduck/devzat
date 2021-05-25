package main

import (
	"bytes"
	"crypto/sha1"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode"

	"github.com/acarl005/stripansi"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	markdown "github.com/quackduck/go-term-markdown"
	ttt "github.com/shurcooL/tictactoe"
	"github.com/slack-go/slack"
	terminal "golang.org/x/term"
)

var (
	//go:embed slackAPI.txt
	slackAPI []byte
	//go:embed art.txt
	artBytes   []byte
	port       = 22
	scrollback = 16

	slackChan = getSendToSlackChan()
	api       = slack.New(string(slackAPI))
	rtm       = api.NewRTM()
	//twitterCreds = loadTwitterCreds()
	client = loadTwitterClient()

	red      = color.New(color.FgHiRed)
	green    = color.New(color.FgHiGreen)
	cyan     = color.New(color.FgHiCyan)
	magenta  = color.New(color.FgHiMagenta)
	yellow   = color.New(color.FgHiYellow)
	blue     = color.New(color.FgHiBlue)
	black    = color.New(color.FgHiBlack)
	white    = color.New(color.FgHiWhite)
	colorArr = []*color.Color{yellow, cyan, magenta, green, white, blue, red}

	devbot = ""

	users      = make([]*user, 0, 10)
	usersMutex = sync.Mutex{}

	allUsers      = make(map[string]string, 100) //map format is u.id => u.name
	allUsersMutex = sync.Mutex{}

	backlog      = make([]backlogmessage, 0, scrollback)
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
	// hasStartedGame = false
	hangGame = new(hangman)

	admins = []string{"d84447e08901391eb36aa8e6d9372b548af55bee3799cd3abb6cdd503fdf2d82", // Ishan Goel
		"f5c7f9826b6e143f6e9c3920767680f503f259570f121138b2465bb2b052a85d", // Ella Xu
		"6056734cc4d9fce31569167735e4808382004629a2d7fe6cb486e663714452fc", // Tommy Pujol
		"e9d47bb4522345d019086d0ed48da8ce491a491923a44c59fd6bfffe6ea73317", // Arav Narula
		"1eab2de20e41abed903ab2f22e7ff56dc059666dbe2ebbce07a8afeece8d0424"} // Shok
)

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
	backlog = append(backlog, backlogmessage{time.Now(), senderName, msg + "\n"})
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
		msg = strings.ReplaceAll(msg, "@"+stripansi.Strip(users[i].name), users[i].name)
		msg = strings.ReplaceAll(msg, "\\"+users[i].name, "@"+stripansi.Strip(users[i].name))
	}
	for i := range users {
		users[i].writeln(senderName, msg)
	}
}

type user struct {
	name            string
	session         ssh.Session
	term            *terminal.Terminal
	bell            bool
	color           color.Color
	id              string
	addr            string
	win             ssh.Window
	closeOnce       sync.Once
	lastMessageTime time.Time
	joinTime        time.Time
	timezone        *time.Location
}

type backlogmessage struct {
	timestamp  time.Time
	senderName string
	text       string
}

type hangman struct {
	word      string
	triesLeft int
	guesses   string // string containing all the guessed characters
}

// Twitter creds
type Credentials struct {
	ConsumerKey       string
	ConsumerSecret    string
	AccessToken       string
	AccessTokenSecret string
}

func sendCurrentUsersTwitterMessage() {
	l.Println("Sending twitter update")
	//broadcast(devbot, "sending twitter update", true)
	names := make([]string, 0, len(users))
	for _, us := range users {
		names = append(names, us.name)
	}
	t, _, err := client.Statuses.Update("People on Devzat rn: "+stripansi.Strip(fmt.Sprint(names))+"\nJoin em with \"ssh devzat.hackclub.com\"\n"+strconv.Itoa(rand.Intn(20)), nil)
	if err != nil {
		l.Println("Got twitter err", err)
		broadcast(devbot, "Sending a twitter update got an error"+fmt.Sprint(err), true)
		return
	}
	broadcast(devbot, "twitter.com/"+t.User.ScreenName+"/status/"+t.IDStr, true)
	//broadcast(devbot, tweet.Entities.Urls)
}

func loadTwitterClient() *twitter.Client {
	d, err := ioutil.ReadFile("twitter-creds.json")
	if err != nil {
		panic(err)
	}
	twitterCreds := new(Credentials)
	err = json.Unmarshal(d, twitterCreds)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%+v", twitterCreds)
	config := oauth1.NewConfig(twitterCreds.ConsumerKey, twitterCreds.ConsumerSecret)
	token := oauth1.NewToken(twitterCreds.AccessToken, twitterCreds.AccessTokenSecret)
	httpClient := config.Client(oauth1.NoContext, token)
	t := twitter.NewClient(httpClient)
	// Verify Credentials
	verifyParams := &twitter.AccountVerifyParams{
		SkipStatus:   twitter.Bool(true),
		IncludeEmail: twitter.Bool(true),
	}

	// we can retrieve the user and verify if the credentials
	// we have used successfully allow us to log in!
	_, _, err = t.Accounts.VerifyCredentials(verifyParams)
	if err != nil {
		panic(err)
	}

	// log.Printf("User's ACCOUNT:\n%+v\n", user)
	return t
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
	u := &user{s.User(), s, term, true, color.Color{}, hex.EncodeToString(hash.Sum(nil)), host, w, sync.Once{}, time.Now(), time.Now(), nil}
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
	if len(backlog) > 0 {
		lastStamp := backlog[0].timestamp
		u.term.Write([]byte(strings.TrimSuffix(u.joinTime.Sub(lastStamp).Round(time.Minute).String(), "0s") + " earlier\n"))
		for i := range backlog {
			//u.writeln("", backlog[i].timestamp.Format(time.Stamp))
			if backlog[i].timestamp.Sub(lastStamp) > time.Minute {
				lastStamp = backlog[i].timestamp
				//u.writeln("", strconv.Itoa(i))
				s := strings.TrimSuffix(u.joinTime.Sub(lastStamp).Round(time.Minute).String(), "0s")
				if s != "" {
					u.term.Write([]byte("\n" + s + " earlier\n"))
				} else {
					u.term.Write([]byte("\nJust now (< 1m earlier)\n"))
				}
			}
			u.writeln(backlog[i].senderName, backlog[i].text)
		}
	}

	u.pickUsername(s.User())
	if _, ok := allUsers[u.id]; !ok {
		broadcast(devbot, "You seem to be new here "+u.name+". Welcome to Devzat! Run /help to see what you can do.", true)
	}
	usersMutex.Lock()
	users = append(users, u)
	go sendCurrentUsersTwitterMessage()
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
	broadcast(devbot, u.name+green.Sprint(" has joined the chat"), true)
	return u
}

func (u *user) close(msg string) {
	u.closeOnce.Do(func() {
		usersMutex.Lock()
		users = remove(users, u)
		go sendCurrentUsersTwitterMessage()
		usersMutex.Unlock()
		broadcast(devbot, msg, true)
		u.session.Close()
	})
}

func (u *user) writeln(senderName string, msg string) {
	if u.bell {
		if strings.Contains(msg, u.name) { // is a ping
			msg += "\a"
		}
	}
	msg = strings.ReplaceAll(msg, `\n`, "\n")
	msg = strings.ReplaceAll(msg, `\`+"\n", `\n`) // let people escape newlines
	if senderName != "" {
		//msg = strings.TrimSpace(mdRender(msg, len(stripansi.Strip(senderName))+2, u.win.Width))
		if strings.HasSuffix(senderName, " <- ") || strings.HasSuffix(senderName, " -> ") {
			msg = strings.TrimSpace(mdRender(msg, len(stripansi.Strip(senderName)), u.win.Width))
			msg = senderName + msg
		} else {
			msg = strings.TrimSpace(mdRender(msg, len(stripansi.Strip(senderName))+2, u.win.Width))
			msg = senderName + ": " + msg
		}
	} else {
		msg = strings.TrimSpace(mdRender(msg, 0, u.win.Width)) // No sender
	}
	if time.Since(u.lastMessageTime) > time.Minute {
		if u.timezone == nil {
			u.term.Write([]byte("\n" + strings.TrimSuffix(time.Since(u.joinTime).Round(time.Minute).String(), "0s") + " in\n"))
		} else {
			u.term.Write([]byte("\n" + time.Now().In(u.timezone).Format("3:04 pm") + "\n"))
		}
		u.lastMessageTime = time.Now()
	}
	//if u.bell {
	//	u.term.Write([]byte(msg + "\a\n")) // "\a" is beep
	//} else {
	u.term.Write([]byte(msg + "\n"))
	//}
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
	u.changeColor(*colorArr[rand.Intn(len(colorArr))]) // also sets prompt
}

func (u *user) changeColor(color color.Color) {
	u.name = color.Sprint(stripansi.Strip(u.name))
	u.color = color
	u.term.SetPrompt(u.name + ": ")
	allUsersMutex.Lock()
	allUsers[u.id] = u.name
	allUsersMutex.Unlock()
	saveBansAndUsers()
}

func (u *user) repl() {
	for {
		line, err := u.term.ReadLine()
		line = strings.TrimSpace(line)

		if err == io.EOF {
			u.close(u.name + red.Sprint(" has left the chat"))
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
		return
	}

	toSlack := true
	if strings.HasPrefix(line, "/hide") && !isSlack {
		toSlack = false
	}
	if strings.HasPrefix(line, "#") && !isSlack && !strings.HasPrefix(line, "# ") {
		toSlack = false
		rest := strings.TrimSpace(strings.TrimPrefix(line, "#"))
		restSplit := strings.Fields(rest)
		if len(restSplit) < 2 {
			u.writeln(devbot, "You gotta have a message mate")
			return
		}
		peer, ok := findUserByName(restSplit[0])
		if !ok {
			u.writeln(devbot, "No such person lol, who you wanna dm?")
			return
		}
		msg := strings.TrimSpace(strings.TrimPrefix(rest, restSplit[0]))
		//u.writeln(u.name+" -> "+peer.name, msg)
		u.writeln(peer.name+" <- ", msg)
		if u == peer {
			go func() {
				time.Sleep(time.Second / 2)
				u.writeln(devbot, "You must be really lonely, DMing yourself. Don't worry, I won't judge :wink:")
			}()
			return
		}
		peer.writeln(u.name+" -> ", msg)
		//peer.writeln(u.name+" -> "+peer.name, msg)
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
	if strings.HasPrefix(line, "/tic") {
		rest := strings.TrimSpace(strings.TrimPrefix(line, "/tic"))
		if rest == "" {
			broadcast(devbot, "Starting a new game of Tic Tac Toe! The first player is always X.", toSlack)
			broadcast(devbot, "Play using /tic <cell num>", toSlack)
			currentPlayer = ttt.X
			tttGame = new(ttt.Board)
			//broadcast(devbot, "```\n"+"0│1│2\n3"+"\n```", toSlack)
			broadcast(devbot, "```\n"+tttPrint(tttGame.Cells)+"\n```", toSlack)
			return
		}

		// if !hasGameStarted(tttGame.Cells) {
		// 	broadcast(devbot, "You need to start a game first!", toSlack)
		// 	broadcast(devbot, "Start again using the /tic command.", toSlack)
		// 	return
		// }

		m, err := strconv.Atoi(rest)
		if err != nil {
			broadcast(devbot, "Make sure you're using a number lol", toSlack)
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
			// hasStartedGame = false
		}
		return
	}

	if !isSlack { // actually sends the message
		broadcast(u.name, line, toSlack)
	}

	devbotChat(line, toSlack)

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
		go func() {
			time.Sleep(time.Second)
			broadcast(devbot, "eggs?", toSlack)
		}()
		return
	}
	if line == "/exit" && !isSlack {
		u.close(u.name + red.Sprint(" has left the chat"))
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
	if strings.HasPrefix(line, "/tz") && !isSlack {
		var err error
		tz := strings.TrimSpace(strings.TrimPrefix(line, "/tz"))
		if tz == "" {
			u.timezone = nil
			return
		}
		u.timezone, err = time.LoadLocation(tz)
		if err != nil {
			broadcast(devbot, "Weird timezone you have there, use Continent/City, EST, PST or see nodatime.org/TimeZones!", toSlack)
			return
		}
		broadcast(devbot, "Done!", toSlack)
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
			broadcast(devbot, "Not authorized", toSlack)
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
			broadcast(devbot, "Not authorized", toSlack)
			return
		}
		bansMutex.Lock()
		bans = append(bans, victim.addr)
		bansMutex.Unlock()
		saveBansAndUsers()
		victim.close(victim.name + " has been banned by " + u.name)
		return
	}
	if strings.HasPrefix(line, "/kick") && !isSlack {
		victim, ok := findUserByName(strings.TrimSpace(strings.TrimPrefix(line, "/kick")))
		if !ok {
			broadcast("", "User not found", toSlack)
			return
		}
		if !auth(u) {
			broadcast(devbot, "Not authorized", toSlack)
			return
		}
		victim.close(victim.name + red.Sprint(" has been kicked by ") + u.name)
		return
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
Tommy Pujol, Sam Poder, Rishi Kothari,  
Amogh Chaubey, Ella Xu, Hugo Hu,  
Robert Goll  
_Possibly more people_


**From my school:**  
Kiyan, Riya, Georgie  
Rayed Hamayun, Aarush Kumar


**From Twitter:**  
Ayush Pathak    @ayshptk  
Bereket         @heybereket  
Sanketh         @SankethYS  
Tony Dinh       @tdinh\_me  
Srushti         @srushtiuniverse  
Surjith         @surjithctly  
Arav Narula     @HeyArav  
Krish Nerkar    @krishnerkar\_  
Amrit           @astro_shenava  
Mudrank Gupta   @mudrankgupta  
Harsh           @harshb__

**And many more have joined!**`, toSlack)
		return
	}

	if line == "/help" {
		broadcast("", `Welcome to Devzat! Devzat is chat over SSH: github.com/quackduck/devzat  
Because there's SSH apps on all platforms, even on mobile, you can join from anywhere.

Interesting features:
* Many, many commands. Run /commands.
* Markdown support! Tables, headers, italics and everything. Just use \\n in place of newlines.
* Code syntax highlighting. Use Markdown fences to send code. Run /example-code to see an example.
* Direct messages! Send a DM using #user <msg>.
* Timezone support, use /tz Continent/City to set your timezone.
* Built in Tic Tac Toe and Hangman! Run /tic or /hang <word> to start new games.
* Emoji replacements! \:rocket\: => :rocket: (like on Slack and Discord)

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
   #<user> <msg>           _DM <user> with <msg>_  
   /users                  _List users_  
   /nick   <name>          _Change your name_  
   /tic    <cell num>      _Play Tic Tac Toe!_  
   /hang   <char/word>     _Play Hangman!_  
   /people                 _See info about nice people who joined_  
   /tz     <zone>          _Change IANA timezone (eg: /tz Asia/Dubai)_  
   /color  <color>         _Change your name's color_  
   /all                    _Get a list of all users ever_  
   /emojis                 _See a list of emojis_  
   /exit                   _Leave the chat_  
   /hide                   _Hide messages from HC Slack_  
   /bell                   _Toggle the ANSI bell used in pings_  
   /id     <user>          _Get a unique ID for a user (hashed IP)_  
   /ban    <user>          _Ban <user> (admin)_  
   /kick   <user>          _Kick <user> (admin)_  
   /help                   _Show help_  
   /commands               _Show this message_`, toSlack)
	}
}

func devbotChat(line string, toSlack bool) {
	if strings.Contains(line, "devbot") {
		if strings.Contains(line, "thank") {
			devbotRespond([]string{"you're welcome",
				"no problem",
				"yeah dw about it",
				":smile:",
				"no worries",
				"you're welcome man!",
				"lol"}, 93, toSlack)
			return
		}
		if strings.Contains(line, "good") || strings.Contains(line, "cool") || strings.Contains(line, "awesome") || strings.Contains(line, "amazing") {
			devbotRespond([]string{"Thanks haha", ":sunglasses:", ":smile:", "lol", "haha", "Thanks lol", "yeeeeeeeee"}, 93, toSlack)
			return
		}
		if strings.Contains(line, "bad") || strings.Contains(line, "idiot") || strings.Contains(line, "stupid") {
			devbotRespond([]string{"what an idiot, bullying a bot", ":(", ":angry:", ":anger:", ":cry:", "I'm in the middle of something okay", "shut up", "Run /help, you need it."}, 60, toSlack)
			return
		}
		if strings.Contains(line, "shut up") {
			devbotRespond([]string{"NO YOU", "You shut up", "what an idiot, bullying a bot"}, 90, toSlack)
			return
		}
		devbotRespond([]string{"Hi I'm devbot", "Hey", "HALLO :rocket:", "Yes?", "Devbot to the rescue!", ":wave:"}, 90, toSlack)
	}
	if line == "help" || strings.Contains(line, "help me") {
		devbotRespond([]string{"Run /help to get help!",
			"Looking for /help?",
			"See available commands with /commands or see help with /help :star:"}, 100, toSlack)
	}

	if strings.Contains(line, "rocket") || strings.Contains(line, "spacex") || strings.Contains(line, "tesla") {
		devbotRespond([]string{"Doge to the mooooon :rocket:",
			"I should have bought ETH before it :rocket:ed to the :moon:",
			":rocket:",
			"I like rockets",
			"SpaceX",
			"Elon Musk OP"}, 80, toSlack)
	}

	if strings.Contains(line, "elon") {
		devbotRespond([]string{"When something is important enough, you do it even if the odds are not in your favor. - Elon",
			"I do think there is a lot of potential if you have a compelling product - Elon",
			"If you're trying to create a company, it's like baking a cake. You have to have all the ingredients in the right proportion. - Elon",
			"Patience is a virtue, and I'm learning patience. It's a tough lesson. - Elon"}, 75, toSlack)
	}

	if !strings.Contains(line, "start") && strings.Contains(line, "star") {
		devbotRespond([]string{"Someone say :star:?",
			"If you like Devzat, give it a star at github.com/quackduck/devzat!",
			":star: github.com/quackduck/devzat", ":star:"}, 90, toSlack)
	}
	if strings.Contains(line, "cool project") || strings.Contains(line, "this is cool") || strings.Contains(line, "this is so cool") {
		devbotRespond([]string{"Thank you :slight_smile:!",
			" If you like Devzat, do give it a star at github.com/quackduck/devzat!",
			"Star Devzat here: github.com/quackduck/devzat"}, 90, toSlack)
	}
}

func devbotRespond(messages []string, chance int, toSlack bool) {
	if chance == 100 || chance > rand.Intn(100) {
		go func() {
			time.Sleep(time.Second / 2)
			pick := messages[rand.Intn(len(messages))]
			broadcast(devbot, pick, toSlack)
		}()
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

func emptyCell(cell string) bool {
	return cell == " "
}

func hasGameStarted(cells [9]ttt.State) bool {
	for i := range cells {
		if !emptyCell(cells[i].String()) {
			return true
		}
	}

	return false
}

func tttPrint(cells [9]ttt.State) string {
	strcells := new([9]string)

	for i := range cells {
		if !hasGameStarted(cells) {
			strcells[i] = strconv.Itoa(i + 1)
		} else {
			strcells[i] = cells[i].String()
		}
	}

	var buf bytes.Buffer

	fmt.Fprintf(&buf, " %v │ %v │ %v \n", strcells[0], strcells[1], strcells[2])
	fmt.Fprintln(&buf, "───┼───┼───")
	fmt.Fprintf(&buf, " %v │ %v │ %v \n", strcells[3], strcells[4], strcells[5])
	fmt.Fprintln(&buf, "───┼───┼───")
	fmt.Fprintf(&buf, " %v │ %v │ %v ", strcells[6], strcells[7], strcells[8])

	result := buf.String()
	result = strings.ReplaceAll(result, ttt.X.String(), color.HiYellowString(ttt.X.String()))
	result = strings.ReplaceAll(result, ttt.O.String(), color.HiGreenString(ttt.O.String()))
	// hasStartedGame = true
	return result
}

func auth(u *user) bool {
	//return true
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
	s = strings.ReplaceAll(s, "<-", "")
	s = strings.ReplaceAll(s, "->", "")
	s = strings.ReplaceAll(s, " ", "-")
	return s
}

func mdRender(a string, beforeMessageLen int, lineWidth int) string {
	md := string(markdown.Render(a, lineWidth-(beforeMessageLen), 0))
	md = strings.TrimSuffix(md, "\n")
	split := strings.Split(md, "\n")
	for i := range split {
		if i == 0 {
			continue // the first line will automatically be padded
		}
		split[i] = strings.Repeat(" ", beforeMessageLen) + split[i]
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
				//i, _ := binary.Varint(h[:])
				i, _ := strconv.ParseInt(hex.EncodeToString(h[:1]), 16, 0)
				broadcast(color.HiYellowString("HC ")+(*colorArr[int(i)%len(colorArr)]).Sprint(strings.Fields(u.RealName)[0]), msg.Text, false)
				runCommands(msg.Text, nil, true)
			}
		case *slack.ConnectedEvent:
			l.Println("Connected to Slack")
		case *slack.InvalidAuthEvent:
			l.Println("Invalid token")
			return
		}
	}
}

func getSendToSlackChan() chan string {
	msgs := make(chan string, 100)
	go func() {
		for msg := range msgs {
			msg = strings.ReplaceAll(stripansi.Strip(msg), `\n`, "\n")
			if strings.HasPrefix(msg, "sshchat: ") { // just in case
				continue
			}
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
	l.Println("remove user was called")
	l.Println(string(debug.Stack()))
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
