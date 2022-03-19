package main

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/chroma"
	chromastyles "github.com/alecthomas/chroma/styles"
	"github.com/mattn/go-sixel"
	markdown "github.com/quackduck/go-term-markdown"
	"github.com/shurcooL/tictactoe"
)

type cmd struct {
	name     string
	run      func(line string, u *user)
	argsInfo string
	info     string
}

var (
	allcmds = make([]cmd, 30)
	cmds    = []cmd{
		{"=<user>", dmCMD, "<msg>", "DM <user> with <msg>"}, // won't actually run, here just to show in docs
		{"users", usersCMD, "", "List users"},
		{"color", colorCMD, "<color>", "Change your name's color"},
		{"exit", exitCMD, "", "Leave the chat"},
		{"help", helpCMD, "", "Show help"},
		{"man", manCMD, "<cmd>", "Get help for a specific command"},
		{"emojis", emojisCMD, "", "See a list of emojis"},
		{"bell", bellCMD, "on|off|all", "ANSI bell on pings (on), never (off) or for every message (all)"},
		{"clear", clearCMD, "", "Clear the screen"},
		{"hang", hangCMD, "<char|word>", "Play hangman"}, // won't actually run, here just to show in docs
		{"tic", ticCMD, "<cell num>", "Play tic tac toe!"},
		{"cd", cdCMD, "#room|user", "Join #room, DM user or run cd to see a list"}, // won't actually run, here just to show in docs
		{"tz", tzCMD, "<zone> [24h]", "Set your IANA timezone (like tz Asia/Dubai) and optionally set 24h"},
		{"nick", nickCMD, "<name>", "Change your username"},
		{"pronouns", pronounsCMD, "<@user|pronoun...>", "Set your pronouns or get another user's"},
		{"theme", themeCMD, "<theme>|list", "Change the syntax highlighting theme"},
		{"rest", commandsRestCMD, "", "Uncommon commands list"}}
	cmdsRest = []cmd{
		{"people", peopleCMD, "", "See info about nice people who joined"},
		{"id", idCMD, "<user>", "Get a unique ID for a user (hashed key)"},
		{"eg-code", exampleCodeCMD, "[big]", "Example syntax-highlighted code"},
		{"lsbans", listBansCMD, "", "List banned IDs"},
		{"ban", banCMD, "<user>", "Ban <user> (admin)"},
		{"unban", unbanCMD, "<IP|ID>", "Unban a person (admin)"},
		{"kick", kickCMD, "<user>", "Kick <user> (admin)"},
		{"art", asciiArtCMD, "", "Show some panda art"},
		{"pwd", pwdCMD, "", "Show your current room"},
		//		{"sixel", sixelCMD, "<png url>", "Render an image in high quality"},
		{"shrug", shrugCMD, "", `¯\\_(ツ)_/¯`}} // won't actually run, here just to show in docs
	secretCMDs = []cmd{
		{"ls", lsCMD, "???", "???"},
		{"cat", catCMD, "???", "???"},
		{"rm", rmCMD, "???", "???"}}
)

func init() {
	cmds = append(cmds, cmd{"cmds", commandsCMD, "", "Show this message"}) // avoid initialization loop
	allcmds = append(append(append(allcmds,
		cmds...), cmdsRest...), secretCMDs...)
}

// runCommands parses a line of raw input from a user and sends a message as
// required, running any commands the user may have called.
// It also accepts a boolean indicating if the line of input is from slack, in
// which case some commands will not be run (such as ./tz and ./exit)
func runCommands(line string, u *user) {
	if detectBadWords(line) {
		banUser("devbot [grow up]", u)
		return
	}

	if line == "" {
		return
	}
	defer func() { // crash protection
		if i := recover(); i != nil {
			mainRoom.broadcast(devbot, "Slap the developers in the face for me, the server almost crashed, also tell them this: "+fmt.Sprint(i, "\n"+string(debug.Stack())))
		}
	}()
	currCmd := strings.Fields(line)[0]
	if u.messaging != nil && currCmd != "=" && currCmd != "cd" && currCmd != "exit" && currCmd != "pwd" { // the commands allowed in a private dm room
		dmRoomCMD(line, u)
		return
	}
	if strings.HasPrefix(line, "=") && !u.isSlack {
		dmCMD(strings.TrimSpace(strings.TrimPrefix(line, "=")), u)
		return
	}

	switch currCmd {
	case "hang":
		hangCMD(strings.TrimSpace(strings.TrimPrefix(line, "hang")), u)
		return
	case "cd":
		cdCMD(strings.TrimSpace(strings.TrimPrefix(line, "cd")), u)
		return
	case "shrug":
		shrugCMD(strings.TrimSpace(strings.TrimPrefix(line, "shrug")), u)
		return
	}

	if u.isSlack {
		u.room.broadcastNoSlack(u.name, line)
	} else {
		u.room.broadcast(u.name, line)
	}

	devbotChat(u.room, line)

	for _, c := range allcmds {
		if c.name == currCmd {
			c.run(strings.TrimSpace(strings.TrimPrefix(line, c.name)), u)
			return
		}
	}
}

// TODO: replacing with asterisks could be faster
func detectBadWords(text string) bool {
	text = strings.ToLower(text)
	badWords := []string{"nigger", "faggot", "tranny", "trannies"} // TODO: add more, it's sad that this is necessary, but the internet is harsh
	badWKillWords := []string{"tranny", "trannies", "transgender", "gay", "muslim", "jew"}
	for _, word := range badWords {
		if strings.Contains(text, word) {
			return true
		}
	}
	for _, word := range badWKillWords {
		if strings.Contains(text, word) && (strings.Contains(text, "kill") || strings.Contains(text, "death") || strings.Contains(text, "dead") || strings.Contains(text, "murder")) {
			return true
		}
	}
	return false
}

func dmCMD(rest string, u *user) {
	restSplit := strings.Fields(rest)
	if len(restSplit) < 2 {
		u.writeln(devbot, "You gotta have a message, mate")
		return
	}
	peer, ok := findUserByName(u.room, restSplit[0])
	if !ok {
		u.writeln(devbot, "No such person lol, who you wanna dm? (you might be in the wrong room)")
		return
	}
	msg := strings.TrimSpace(strings.TrimPrefix(rest, restSplit[0]))
	u.writeln(peer.name+" <- ", msg)
	if u == peer {
		devbotRespond(u.room, []string{"You must be really lonely, DMing yourself.",
			"Don't worry, I won't judge :wink:",
			"srsly?",
			"what an idiot"}, 30)
		return
	}
	peer.writeln(u.name+" -> ", msg)
}

func hangCMD(rest string, u *user) {
	if len(rest) > 1 {
		if !u.isSlack {
			u.writeln(u.name, "hang "+rest)
			u.writeln(devbot, "(that word won't show dw)")
		}
		hangGame = &hangman{rest, 15, " "} // default value of guesses so empty space is given away
		u.room.broadcast(devbot, u.name+" has started a new game of Hangman! Guess letters with hang <letter>")
		u.room.broadcast(devbot, "```\n"+hangPrint(hangGame)+"\nTries: "+strconv.Itoa(hangGame.triesLeft)+"\n```")
		return
	}
	if !u.isSlack {
		u.room.broadcast(u.name, "hang "+rest)
	}
	if strings.Trim(hangGame.word, hangGame.guesses) == "" {
		u.room.broadcast(devbot, "The game has ended. Start a new game with hang <word>")
		return
	}
	if len(rest) == 0 {
		u.room.broadcast(devbot, "Start a new game with hang <word> or guess with hang <letter>")
		return
	}
	if hangGame.triesLeft == 0 {
		u.room.broadcast(devbot, "No more tries! The word was "+hangGame.word)
		return
	}
	if strings.Contains(hangGame.guesses, rest) {
		u.room.broadcast(devbot, "You already guessed "+rest)
		return
	}
	hangGame.guesses += rest
	if !(strings.Contains(hangGame.word, rest)) {
		hangGame.triesLeft--
	}
	display := hangPrint(hangGame)
	u.room.broadcast(devbot, "```\n"+display+"\nTries: "+strconv.Itoa(hangGame.triesLeft)+"\n```")
	if strings.Trim(hangGame.word, hangGame.guesses) == "" {
		u.room.broadcast(devbot, "You got it! The word was "+hangGame.word)
	} else if hangGame.triesLeft == 0 {
		u.room.broadcast(devbot, "No more tries! The word was "+hangGame.word)
	}
}

func clearCMD(_ string, u *user) {
	u.term.Write([]byte("\033[H\033[2J"))
}

func usersCMD(_ string, u *user) {
	u.room.broadcast("", printUsersInRoom(u.room))
}

func dmRoomCMD(line string, u *user) {
	u.writeln(u.messaging.name+" <- ", line)
	if u == u.messaging {
		devbotRespond(u.room, []string{"You must be really lonely, DMing yourself.",
			"Don't worry, I won't judge :wink:",
			"srsly?",
			"what an idiot"}, 30)
		return
	}
	u.messaging.writeln(u.name+" -> ", line)
}

func ticCMD(rest string, u *user) {
	if rest == "" {
		u.room.broadcast(devbot, "Starting a new game of Tic Tac Toe! The first player is always X.")
		u.room.broadcast(devbot, "Play using tic <cell num>")
		currentPlayer = tictactoe.X
		tttGame = new(tictactoe.Board)
		u.room.broadcast(devbot, "```\n"+" 1 │ 2 │ 3\n───┼───┼───\n 4 │ 5 │ 6\n───┼───┼───\n 7 │ 8 │ 9\n"+"\n```")
		return
	}
	m, err := strconv.Atoi(rest)
	if err != nil {
		u.room.broadcast(devbot, "Make sure you're using a number lol")
		return
	}
	if m < 1 || m > 9 {
		u.room.broadcast(devbot, "Moves are numbers between 1 and 9!")
		return
	}
	err = tttGame.Apply(tictactoe.Move(m-1), currentPlayer)
	if err != nil {
		u.room.broadcast(devbot, err.Error())
		return
	}
	u.room.broadcast(devbot, "```\n"+tttPrint(tttGame.Cells)+"\n```")
	if currentPlayer == tictactoe.X {
		currentPlayer = tictactoe.O
	} else {
		currentPlayer = tictactoe.X
	}
	if !(tttGame.Condition() == tictactoe.NotEnd) {
		u.room.broadcast(devbot, tttGame.Condition().String())
		currentPlayer = tictactoe.X
		tttGame = new(tictactoe.Board)
	}
}

func exitCMD(_ string, u *user) {
	u.close(u.name + red.Paint(" has left the chat"))
}

func bellCMD(rest string, u *user) {
	switch rest {
	case "off":
		u.bell = false
		u.pingEverytime = false
		u.room.broadcast("", "bell off (never)")
	case "on":
		u.bell = true
		u.pingEverytime = false
		u.room.broadcast("", "bell on (pings)")
	case "all":
		u.pingEverytime = true
		u.room.broadcast("", "bell all (every message)")
	case "", "status":
		if u.bell {
			u.room.broadcast("", "bell on (pings)")
		} else if u.pingEverytime {
			u.room.broadcast("", "bell all (every message)")
		} else { // bell is off
			u.room.broadcast("", "bell off (never)")
		}
	default:
		u.room.broadcast(devbot, "your options are off, on and all")
	}
}

func sixelCMD(url string, u *user) {
	r, err := http.Get(url)
	if err != nil {
		u.room.broadcast(devbot, "huh, are you sure that's a working link?")
		return
	}
	i, _, err := image.Decode(r.Body)
	if err != nil {
		u.room.broadcast(devbot, "are you sure that's a link to a png or a jpeg?")
		return
	}
	b := new(bytes.Buffer)
	err = sixel.NewEncoder(b).Encode(i)
	if err != nil {
		u.room.broadcast(devbot, "uhhh I got this error trying to encode the image: "+err.Error())
	}
	for _, us := range u.room.users {
		us.term.Write(b.Bytes()) // TODO: won't shpw up in the backlog, is that okay?
	}
}

func cdCMD(rest string, u *user) {
	if u.messaging != nil {
		u.messaging = nil
		u.writeln(devbot, "Left private chat")
		if rest == "" || rest == ".." {
			return
		}
	}
	if strings.HasPrefix(rest, "#") {
		u.room.broadcast(u.name, "cd "+rest)
		if v, ok := rooms[rest]; ok {
			u.changeRoom(v)
		} else {
			rooms[rest] = &room{rest, make([]*user, 0, 10), sync.Mutex{}}
			u.changeRoom(rooms[rest])
		}
		return
	}
	if rest == "" {
		u.room.broadcast(u.name, "cd "+rest)
		type kv struct {
			roomName   string
			numOfUsers int
		}
		var ss []kv
		for k, v := range rooms {
			ss = append(ss, kv{k, len(v.users)})
		}
		sort.Slice(ss, func(i, j int) bool {
			return ss[i].numOfUsers > ss[j].numOfUsers
		})
		roomsInfo := ""
		for _, kv := range ss {
			roomsInfo += blue.Paint(kv.roomName) + ": " + printUsersInRoom(rooms[kv.roomName]) + "  \n"
		}
		u.room.broadcast("", "Rooms and users  \n"+strings.TrimSpace(roomsInfo))
		return
	}
	name := strings.Fields(rest)[0]
	if len(name) == 0 {
		u.writeln(devbot, "You think people have empty names?")
		return
	}
	peer, ok := findUserByName(u.room, name)
	if !ok {
		u.writeln(devbot, "No such person lol, who do you want to dm? (you might be in the wrong room)")
		return
	}
	u.messaging = peer
	u.writeln(devbot, "Now in DMs with "+peer.name+". To leave use cd ..")
}

func tzCMD(tzArg string, u *user) {
	var err error
	if tzArg == "" {
		u.timezone = nil
		return
	}
	tzArgList := strings.Fields(tzArg)
	tz := tzArgList[0]
	switch tz {
	case "PST", "PDT":
		tz = "PST8PDT"
	case "CST", "CDT":
		tz = "CST6CDT"
	case "EST", "EDT":
		tz = "EST5EDT"
	case "MT":
		tz = "America/Phoenix"
	}
	u.timezone, err = time.LoadLocation(tz)
	if err != nil {
		u.room.broadcast(devbot, "Weird timezone you have there, use the format Continent/City, the usual US timezones (PST, PDT, EST, EDT...) or check nodatime.org/TimeZones!")
		return
	}
	if len(tzArgList) == 2 {
		u.formatTime24 = tzArgList[1] == "24h"
	} else {
		u.formatTime24 = false
	}
	u.room.broadcast(devbot, "Changed your timezone!")
}

func idCMD(line string, u *user) {
	victim, ok := findUserByName(u.room, line)
	if !ok {
		u.room.broadcast("", "User not found")
		return
	}
	u.room.broadcast("", victim.id)
}

func nickCMD(line string, u *user) {
	u.pickUsername(line)
	return
}

func listBansCMD(_ string, u *user) {
	msg := "Printing bans by ID:  \n"
	for i := 0; i < len(bans); i++ {
		msg += cyan.Cyan(strconv.Itoa(i+1)) + ". " + bans[i].ID + "  \n"
	}
	u.room.broadcast(devbot, msg)
}

func unbanCMD(toUnban string, u *user) {
	if !auth(u) {
		u.room.broadcast(devbot, "Not authorized")
		return
	}

	for i := 0; i < len(bans); i++ {
		if bans[i].ID == toUnban || bans[i].Addr == toUnban { // allow unbanning by either ID or IP
			u.room.broadcast(devbot, "Unbanned person: "+bans[i].ID)
			// remove this ban
			bans = append(bans[:i], bans[i+1:]...)
		}
	}

	saveBans()
}

func banCMD(line string, u *user) {
	victim, ok := findUserByName(u.room, line)
	if !ok {
		u.room.broadcast("", "User not found")
		return
	}
	if !auth(u) {
		u.room.broadcast(devbot, "Not authorized")
		return
	}
	banUser(u.name, victim)
}

func banUser(banner string, victim *user) {
	bans = append(bans, ban{victim.addr, victim.id})
	saveBans()
	victim.close(victim.name + " has been banned by " + banner)
}

func kickCMD(line string, u *user) {
	victim, ok := findUserByName(u.room, line)
	if !ok {
		u.room.broadcast("", "User not found")
		return
	}
	if !auth(u) {
		u.room.broadcast(devbot, "Not authorized")
		return
	}
	victim.close(victim.name + red.Paint(" has been kicked by ") + u.name)
}

func colorCMD(rest string, u *user) {
	if rest == "which" {
		u.room.broadcast(devbot, "you're using "+u.color)
	} else if err := u.changeColor(rest); err != nil {
		u.room.broadcast(devbot, err.Error())
	}
}

func peopleCMD(_ string, u *user) {
	u.room.broadcast("", `
**Hack Club members**  
Zach Latta     - Founder of Hack Club  
Zachary Fogg   - Hack Club Game Designer  
Matthew        - Hack Club HQ  
Caleb Denio, Safin Singh, Eleeza A  
Jubril, Sarthak Mohanty, Anghe,  
Tommy Pujol, Sam Poder, Rishi Kothari,  
Amogh Chaubey, Ella Xu, Hugo Hu,  
Robert Goll, Tanishq Soni, Arash Nur Iman,  
Temi, Aiden Bai, Ivan Bowman, @epic  
Belle See, Fayd, Benjamin Smith  
Matt Gleich, Jason Appah  
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

**And many more have joined!**`)
}

func helpCMD(_ string, u *user) {
	u.room.broadcast("", `Welcome to Devzat! Devzat is chat over SSH: github.com/quackduck/devzat  
Because there's SSH apps on all platforms, even on mobile, you can join from anywhere.

Run cmds to see a list of commands.

Interesting features:
* Rooms! Run cd to see all rooms and use cd #foo to join a new room.
* Markdown support! Tables, headers, italics and everything. Just use \\n in place of newlines.
* Code syntax highlighting. Use Markdown fences to send code. Run eg-code to see an example.
* Direct messages! Send a quick DM using =user <msg> or stay in DMs by running cd @user.
* Timezone support, use tz Continent/City to set your timezone.
* Built in Tic Tac Toe and Hangman! Run tic or hang <word> to start new games.
* Emoji replacements! \:rocket\: => :rocket: (like on Slack and Discord)

For replacing newlines, I often use bulkseotools.com/add-remove-line-breaks.php.

Join the Devzat discord server: https://discord.gg/5AUjJvBHeT

Made by Ishan Goel with feature ideas from friends.  
Thanks to Caleb Denio for lending his server!`)
}

func catCMD(line string, u *user) {
	if line == "" {
		u.room.broadcast("", "usage: cat [-benstuv] [file ...]")
	} else if line == "README.md" {
		helpCMD(line, u)
	} else {
		u.room.broadcast("", "cat: "+line+": Permission denied")
	}
}

func rmCMD(line string, u *user) {
	if line == "" {
		u.room.broadcast("", `usage: rm [-f | -i] [-dPRrvW] file ...
unlink file`)
	} else {
		u.room.broadcast("", "rm: "+line+": Permission denied, sucker")
	}
}

func exampleCodeCMD(line string, u *user) {
	if line == "big" {
		u.room.broadcast(devbot, "```go\npackage main\n\nimport \"fmt\"\n\nfunc sum(nums ...int) {\n    fmt.Print(nums, \" \")\n    total := 0\n    for _, num := range nums {\n        total += num\n    }\n    fmt.Println(total)\n}\n\nfunc main() {\n\n    sum(1, 2)\n    sum(1, 2, 3)\n\n    nums := []int{1, 2, 3, 4}\n    sum(nums...)\n}\n```")
		return
	}
	u.room.broadcast(devbot, "\n```go\npackage main\nimport \"fmt\"\nfunc main() {\n   fmt.Println(\"Example!\")\n}\n```")
}

func init() { // add Matt Gleich's blackbird theme from https://github.com/blackbirdtheme/vscode/blob/master/themes/blackbird-midnight-color-theme.json#L175
	red := "#ff1131" // added saturation
	redItalic := "italic " + red
	white := "#fdf7cd"
	yellow := "#e1db3f"
	blue := "#268ef8"  // added saturation
	green := "#22e327" // added saturation
	gray := "#5a637e"
	teal := "#00ecd8"
	tealItalic := "italic " + teal

	chromastyles.Register(chroma.MustNewStyle("blackbird", chroma.StyleEntries{
		chroma.Text:                white,
		chroma.Error:               red,
		chroma.Comment:             gray,
		chroma.Keyword:             redItalic,
		chroma.KeywordNamespace:    redItalic,
		chroma.KeywordType:         tealItalic,
		chroma.Operator:            blue,
		chroma.Punctuation:         white,
		chroma.Name:                white,
		chroma.NameAttribute:       white,
		chroma.NameClass:           green,
		chroma.NameConstant:        tealItalic,
		chroma.NameDecorator:       green,
		chroma.NameException:       red,
		chroma.NameFunction:        green,
		chroma.NameOther:           white,
		chroma.NameTag:             yellow,
		chroma.LiteralNumber:       blue,
		chroma.Literal:             yellow,
		chroma.LiteralDate:         yellow,
		chroma.LiteralString:       yellow,
		chroma.LiteralStringEscape: teal,
		chroma.GenericDeleted:      red,
		chroma.GenericEmph:         "italic",
		chroma.GenericInserted:     green,
		chroma.GenericStrong:       "bold",
		chroma.GenericSubheading:   yellow,
		chroma.Background:          "bg:#000000",
	}))
}

func themeCMD(line string, u *user) {
	if line == "list" {
		u.room.broadcast(devbot, "Available themes: "+strings.Join(chromastyles.Names(), ", "))
		return
	}
	for _, name := range chromastyles.Names() {
		if name == line {
			markdown.CurrentTheme = chromastyles.Get(name)
			u.room.broadcast(devbot, "Theme set to "+name)
			return
		}
	}
	u.room.broadcast(devbot, "What theme is that? Use theme list to see what's available.")
}

func asciiArtCMD(_ string, u *user) {
	u.room.broadcast("", art)
}

func pwdCMD(_ string, u *user) {
	if u.messaging != nil {
		u.writeln("", u.messaging.name)
		u.messaging.writeln("", u.messaging.name)
	} else {
		u.room.broadcast("", u.room.name)
	}
}

func shrugCMD(line string, u *user) {
	u.room.broadcast(u.name, line+` ¯\\_(ツ)_/¯`)
}

func pronounsCMD(line string, u *user) {
	args := strings.Fields(line)

	if line == "" {
		u.room.broadcast(devbot, "Set pronouns by providing em or query a user's pronouns!")
		return
	}

	if len(args) == 1 && strings.HasPrefix(args[0], "@") {
		victim, ok := findUserByName(u.room, args[0][1:])
		if !ok {
			u.room.broadcast(devbot, "Who's that?")
			return
		}
		u.room.broadcast(devbot, victim.name+"'s pronouns are "+victim.displayPronouns())
		return
	}

	u.pronouns = strings.Fields(strings.ReplaceAll(strings.ToLower(line), "\n", ""))
	//u.changeColor(u.color) // refresh pronouns
	u.room.broadcast(devbot, u.name+" now goes by "+u.displayPronouns())
}

func emojisCMD(_ string, u *user) {
	u.room.broadcast(devbot, "Check out github.com/ikatyang/emoji-cheat-sheet")
}

func commandsRestCMD(_ string, u *user) {
	u.room.broadcast("", "The rest  \n"+autogenCommands(cmdsRest))
}

func manCMD(rest string, u *user) {
	if rest == "" {
		u.room.broadcast(devbot, "What command do you want help with?")
		return
	}

	for _, c := range allcmds {
		if c.name == rest {
			u.room.broadcast(devbot, "Usage: "+c.name+" "+c.argsInfo+"  \n"+c.info)
			return
		}
	}
	u.room.broadcast("", "This system has been minimized by removing packages and content that are not required on a system that users do not log into.\n\nTo restore this content, including manpages, you can run the 'unminimize' command. You will still need to ensure the 'man-db' package is installed.")
}

func lsCMD(rest string, u *user) {
	if rest != "" {
		u.room.broadcast("", "ls: "+rest+" Permission denied")
		return
	}
	roomList := ""
	for _, r := range rooms {
		roomList += blue.Paint(r.name + "/ ")
	}
	usersList := ""
	for _, us := range u.room.users {
		usersList += us.name + blue.Paint("/ ")
	}
	u.room.broadcast("", "README.md "+usersList+roomList)
}

func commandsCMD(_ string, u *user) {
	u.room.broadcast("", "Commands  \n"+autogenCommands(cmds))
}
