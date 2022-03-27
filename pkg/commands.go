package pkg

import (
	"devchat/pkg/commands/dm"
	"errors"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/chroma"
	chromastyles "github.com/alecthomas/chroma/styles"
	markdown "github.com/quackduck/go-term-markdown"
	"github.com/shurcooL/tictactoe"
)

type CommandType = int

const (
	CommandTypeNormal = iota
	CommandTypeRest   // whatever the fuck this means...
	CommandTypeSecret
)

type Command struct {
	name     string
	run      func(line string, u *User) error
	argsInfo string
	info     string
}

const (
	maxLengthRoomName = 30
)

const (
	fmtRecover = "Slap the developers in the face for me, the server almost crashed, also tell them this: %v, stack: %v"
)

type commandRegistry = map[string]CommandFunc

type Registrar struct {
	commandRegistry
}

func (r *Registrar) Register(cr CommandRegistration) error {
	if cr == nil {
		return errors.New("empty command registration given")
	}

	if cr.Name() == "" {
		return errors.New("empty command name given")
	}

	if cr.Command == nil {
		return errors.New("nil command func given")
	}

	if r.commandRegistry == nil {
		r.commandRegistry = make(commandRegistry)
	}

	r.commandRegistry[cr.Name()] = cr.Command

	return nil
}

func (c *Registrar) init() error {
	commandsToRegister := []CommandRegistration{
		&dm.Command{},
	}

	for _, cr := range commandsToRegister {
		if err := c.Register(cr); err != nil {
			return err
		}
	}

	return nil
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

func hangCMD(rest string, u *User) error {
	if len(rest) > 1 {
		if !u.isSlack {
			u.writeln(u.name, "hang "+rest)
			u.writeln(devbot, "(that word won't show dw)")
		}
		hangGame = &hangman{rest, 15, " "} // default value of guesses so empty space is given away
		u.Room.Broadcast(devbot, u.name+" has started a new game of Hangman! Guess letters with hang <letter>")
		u.Room.Broadcast(devbot, "```\n"+hangPrint(hangGame)+"\nTries: "+strconv.Itoa(hangGame.triesLeft)+"\n```")
		return
	}
	if !u.isSlack {
		u.Room.Broadcast(u.name, "hang "+rest)
	}
	if strings.Trim(hangGame.word, hangGame.guesses) == "" {
		u.Room.Broadcast(devbot, "The game has ended. Start a new game with hang <word>")
		return
	}
	if len(rest) == 0 {
		u.Room.Broadcast(devbot, "Start a new game with hang <word> or guess with hang <letter>")
		return
	}
	if hangGame.triesLeft == 0 {
		u.Room.Broadcast(devbot, "No more tries! The word was "+hangGame.word)
		return
	}
	if strings.Contains(hangGame.guesses, rest) {
		u.Room.Broadcast(devbot, "You already guessed "+rest)
		return
	}
	hangGame.guesses += rest
	if !(strings.Contains(hangGame.word, rest)) {
		hangGame.triesLeft--
	}
	display := hangPrint(hangGame)
	u.Room.Broadcast(devbot, "```\n"+display+"\nTries: "+strconv.Itoa(hangGame.triesLeft)+"\n```")
	if strings.Trim(hangGame.word, hangGame.guesses) == "" {
		u.Room.Broadcast(devbot, "You got it! The word was "+hangGame.word)
	} else if hangGame.triesLeft == 0 {
		u.Room.Broadcast(devbot, "No more tries! The word was "+hangGame.word)
	}
}

func clearCMD(_ string, u *User) error {
	u.term.Write([]byte("\033[H\033[2J"))
}

func usersCMD(_ string, u *User) error {
	u.Room.Broadcast("", printUsersInRoom(u.Room))
}

func dmRoomCMD(line string, u *User) error {
	u.writeln(u.Messaging.name+" <- ", line)
	if u == u.Messaging {
		devbotRespond(u.Room, []string{"You must be really lonely, DMing yourself.",
			"Don't worry, I won't judge :wink:",
			"srsly?",
			"what an idiot"}, 30)
		return
	}
	u.Messaging.writeln(u.name+" -> ", line)
}

func ticCMD(rest string, u *User) error {
	if rest == "" {
		u.Room.Broadcast(devbot, "Starting a new game of Tic Tac Toe! The first player is always X.")
		u.Room.Broadcast(devbot, "Play using tic <cell num>")
		currentPlayer = tictactoe.X
		tttGame = new(tictactoe.Board)
		u.Room.Broadcast(devbot, "```\n"+" 1 │ 2 │ 3\n───┼───┼───\n 4 │ 5 │ 6\n───┼───┼───\n 7 │ 8 │ 9\n"+"\n```")
		return
	}
	m, err := strconv.Atoi(rest)
	if err != nil {
		u.Room.Broadcast(devbot, "Make sure you're using a number lol")
		return
	}
	if m < 1 || m > 9 {
		u.Room.Broadcast(devbot, "Moves are numbers between 1 and 9!")
		return
	}
	err = tttGame.Apply(tictactoe.Move(m-1), currentPlayer)
	if err != nil {
		u.Room.Broadcast(devbot, err.Error())
		return
	}
	u.Room.Broadcast(devbot, "```\n"+tttPrint(tttGame.Cells)+"\n```")
	if currentPlayer == tictactoe.X {
		currentPlayer = tictactoe.O
	} else {
		currentPlayer = tictactoe.X
	}
	if !(tttGame.Condition() == tictactoe.NotEnd) {
		u.Room.Broadcast(devbot, tttGame.Condition().String())
		currentPlayer = tictactoe.X
		tttGame = new(tictactoe.Board)
	}
}

func exitCMD(_ string, u *User) error {
	u.close(u.name + red.Paint(" has left the chat"))
}

func bellCMD(rest string, u *User) error {
	switch rest {
	case "off":
		u.bell = false
		u.pingEverytime = false
		u.Room.Broadcast("", "bell off (never)")
	case "on":
		u.bell = true
		u.pingEverytime = false
		u.Room.Broadcast("", "bell on (pings)")
	case "all":
		u.pingEverytime = true
		u.Room.Broadcast("", "bell all (every message)")
	case "", "status":
		if u.bell {
			u.Room.Broadcast("", "bell on (pings)")
		} else if u.pingEverytime {
			u.Room.Broadcast("", "bell all (every message)")
		} else { // bell is off
			u.Room.Broadcast("", "bell off (never)")
		}
	default:
		u.Room.Broadcast(devbot, "your options are off, on and all")
	}
}

func cdCMD(rest string, u *User) error {
	if u.Messaging != nil {
		u.Messaging = nil
		u.writeln(devbot, "Left private chat")
		if rest == "" || rest == ".." {
			return
		}
	}
	if strings.HasPrefix(rest, "#") {
		u.Room.Broadcast(u.name, "cd "+rest)
		if len(rest) > maxLengthRoomName {
			rest = rest[0:maxLengthRoomName]
			u.Room.Broadcast(devbot, "room name lengths are limited, so I'm shortening it to "+rest+".")
		}
		if v, ok := rooms[rest]; ok {
			u.changeRoom(v)
		} else {
			rooms[rest] = &Room{rest, make([]*User, 0, 10), sync.Mutex{}}
			u.changeRoom(rooms[rest])
		}
		return
	}
	if rest == "" {
		u.Room.Broadcast(u.name, "cd "+rest)
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
		u.Room.Broadcast("", "Rooms and users  \n"+strings.TrimSpace(roomsInfo))
		return
	}
	name := strings.Fields(rest)[0]
	if len(name) == 0 {
		u.writeln(devbot, "You think people have empty names?")
		return
	}
	peer, ok := findUserByName(u.Room, name)
	if !ok {
		u.writeln(devbot, "No such person lol, who do you want to dm? (you might be in the wrong room)")
		return
	}
	u.Messaging = peer
	u.writeln(devbot, "Now in DMs with "+peer.name+". To leave use cd ..")
}

func tzCMD(tzArg string, u *User) error {
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
		u.Room.Broadcast(devbot, "Weird timezone you have there, use the format Continent/City, the usual US timezones (PST, PDT, EST, EDT...) or check nodatime.org/TimeZones!")
		return
	}
	if len(tzArgList) == 2 {
		u.formatTime24 = tzArgList[1] == "24h"
	} else {
		u.formatTime24 = false
	}
	u.Room.Broadcast(devbot, "Changed your timezone!")
}

func idCMD(line string, u *User) error {
	victim, ok := findUserByName(u.Room, line)
	if !ok {
		u.Room.Broadcast("", "User not found")
		return
	}
	u.Room.Broadcast("", victim.id)
}

func nickCMD(line string, u *User) error {
	u.pickUsername(line) //nolint:errcheck // if reading input fails, the next repl will err out
}

func listBansCMD(_ string, u *User) error {
	msg := "Printing bans by ID:  \n"
	for i := 0; i < len(bans); i++ {
		msg += cyan.Cyan(strconv.Itoa(i+1)) + ". " + bans[i].ID + "  \n"
	}
	u.Room.Broadcast(devbot, msg)
}

func unbanCMD(toUnban string, u *User) error {
	isAdmin, errCheckAdmin := checkIsAdmin(u)
	if errCheckAdmin != nil {
		return fmt.Errorf("could not unban: %v", errCheckAdmin)
	}

	if !isAdmin {
		u.Room.Broadcast(devbot, "Not authorized")
		return nil
	}

	if unbanIDorIP(toUnban) {
		u.Room.Broadcast(devbot, "Unbanned person: "+toUnban)
		saveBans()

		return nil
	}

	u.Room.Broadcast(devbot, "I couldn't find that person")

	return nil
}

// unbanIDorIP unbans an ID or an IP, but does NOT save bans to the bans file.
// It returns whether the person was found, and so, whether the bans slice was modified.
func unbanIDorIP(toUnban string) bool {
	for i := 0; i < len(bans); i++ {
		if bans[i].ID == toUnban || bans[i].Addr == toUnban { // allow unbanning by either ID or IP
			// remove this ban
			bans = append(bans[:i], bans[i+1:]...)
			return true
		}
	}
	return false
}

func banCMD(line string, u *User) error {
	if !checkIsAdmin(u) {
		u.Room.Broadcast(devbot, "Not authorized")
		return
	}
	split := strings.Split(line, " ")
	if len(split) == 0 {
		u.Room.Broadcast(devbot, "Which User do you want to ban?")
		return
	}
	victim, ok := findUserByName(u.Room, split[0])
	if !ok {
		u.Room.Broadcast("", "User not found")
		return
	}
	// check if the ban is for a certain duration
	if len(split) > 1 {
		dur, err := time.ParseDuration(split[1])
		if err != nil {
			u.Room.Broadcast(devbot, "I couldn't parse that as a duration")
			return
		}
		bans = append(bans, ban{victim.addr, victim.id})
		victim.close(victim.name + " has been banned by " + u.name + " for " + dur.String())
		go func(id string) {
			time.Sleep(dur)
			unbanIDorIP(id)
		}(victim.id) // evaluate id now, call unban with that value later
		return
	}
	banUser(u.name, victim)
}

func banUser(banner string, victim *User) {
	bans = append(bans, ban{victim.addr, victim.id})
	saveBans()
	victim.close(victim.name + " has been banned by " + banner)
}

func kickCMD(line string, u *User) error {
	victim, ok := findUserByName(u.Room, line)
	if !ok {
		u.Room.Broadcast("", "User not found")
		return
	}
	if !checkIsAdmin(u) {
		u.Room.Broadcast(devbot, "Not authorized")
		return
	}
	victim.close(victim.name + red.Paint(" has been kicked by ") + u.name)
}

func colorCMD(rest string, u *User) error {
	if rest == "which" {
		u.Room.Broadcast(devbot, "fg: "+u.color+" & bg: "+u.colorBG)
	} else if err := u.changeColor(rest); err != nil {
		u.Room.Broadcast(devbot, err.Error())
	}
}

func adminsCMD(_ string, u *User) error {
	msg := "Admins:  \n"
	for i := range admins {
		msg += admins[i] + ": " + i + "  \n"
	}
	u.Room.Broadcast(devbot, msg)
}

func peopleCMD(_ string, u *User) error {
	u.Room.Broadcast("", `
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

func helpCMD(_ string, u *User) error {
	u.Room.Broadcast("", `Welcome to Devzat! Devzat is chat over SSH: github.com/quackduck/devzat  
Because there's SSH apps on all platforms, even on mobile, you can join from anywhere.

Run cmds to see a list of commands.

Interesting features:
* Rooms! Run cd to see all Rooms and use cd #foo to join a new room.
* Markdown support! Tables, headers, italics and everything. Just use \\n in place of newlines.
* Code syntax highlighting. Use Markdown fences to send code. Run eg-code to see an example.
* Direct messages! Send a quick DM using =User <msg> or stay in DMs by running cd @User.
* Timezone support, use tz Continent/City to set your timezone.
* Built in Tic Tac Toe and Hangman! Run tic or hang <word> to start new games.
* Emoji replacements! \:rocket\: => :rocket: (like on Slack and Discord)

For replacing newlines, I often use bulkseotools.com/add-remove-line-breaks.php.

Join the Devzat discord server: https://discord.gg/5AUjJvBHeT

Made by Ishan Goel with feature ideas from friends.  
Thanks to Caleb Denio for lending his server!`)
}

func catCMD(line string, u *User) error {
	if line == "" {
		u.Room.Broadcast("", "usage: cat [-benstuv] [file ...]")
	} else if line == "README.md" {
		helpCMD(line, u)
	} else {
		u.Room.Broadcast("", "cat: "+line+": Permission denied")
	}
}

func rmCMD(line string, u *User) error {
	if line == "" {
		u.Room.Broadcast("", `usage: rm [-f | -i] [-dPRrvW] file ...
unlink file`)
	} else {
		u.Room.Broadcast("", "rm: "+line+": Permission denied, sucker")
	}
}

func exampleCodeCMD(line string, u *User) error {
	if line == "big" {
		u.Room.Broadcast(devbot, "```go\npackage MainRoom\n\nimport \"fmt\"\n\nfunc sum(nums ...int) {\n    fmt.Print(nums, \" \")\n    total := 0\n    for _, num := range nums {\n        total += num\n    }\n    fmt.Println(total)\n}\n\nfunc MainRoom() {\n\n    sum(1, 2)\n    sum(1, 2, 3)\n\n    nums := []int{1, 2, 3, 4}\n    sum(nums...)\n}\n```")
		return
	}
	u.Room.Broadcast(devbot, "\n```go\npackage MainRoom\nimport \"fmt\"\nfunc MainRoom() {\n   fmt.Println(\"Example!\")\n}\n```")
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

func themeCMD(line string, u *User) error {
	if line == "list" {
		u.Room.Broadcast(devbot, "Available themes: "+strings.Join(chromastyles.Names(), ", "))
		return
	}
	for _, name := range chromastyles.Names() {
		if name == line {
			markdown.CurrentTheme = chromastyles.Get(name)
			u.Room.Broadcast(devbot, "Theme set to "+name)
			return
		}
	}
	u.Room.Broadcast(devbot, "What theme is that? Use theme list to see what's available.")
}

func asciiArtCMD(_ string, u *User) error {
	u.Room.Broadcast("", art)
}

func pwdCMD(_ string, u *User) error {
	if u.Messaging != nil {
		u.writeln("", u.Messaging.name)
		u.Messaging.writeln("", u.Messaging.name)
	} else {
		u.Room.Broadcast("", u.Room.name)
	}
}

func shrugCMD(line string, u *User) error {
	u.Room.Broadcast(u.name, line+` ¯\\\_(ツ)\_/¯`)
}

func pronounsCMD(line string, u *User) error {
	args := strings.Fields(line)

	if line == "" {
		u.Room.Broadcast(devbot, "Set pronouns by providing em or query a User's pronouns!")
		return
	}

	if len(args) == 1 && strings.HasPrefix(args[0], "@") {
		victim, ok := findUserByName(u.Room, args[0][1:])
		if !ok {
			u.Room.Broadcast(devbot, "Who's that?")
			return
		}
		u.Room.Broadcast(devbot, victim.name+"'s pronouns are "+victim.displayPronouns())
		return
	}

	u.pronouns = strings.Fields(strings.ReplaceAll(strings.ToLower(line), "\n", ""))
	//u.changeColor(u.color) // refresh pronouns
	u.Room.Broadcast(devbot, u.name+" now goes by "+u.displayPronouns())
}

func emojisCMD(_ string, u *User) error {
	u.Room.Broadcast(devbot, "Check out https\\://github.com/ikatyang/emoji-cheat-sheet")
}

func commandsRestCMD(_ string, u *User) error {
	u.Room.Broadcast("", "The rest  \n"+autogenCommands(cmdsRest))
}

func manCMD(rest string, u *User) error {
	if rest == "" {
		u.Room.Broadcast(devbot, "What command do you want help with?")
		return
	}

	for _, c := range allcmds {
		if c.name == rest {
			u.Room.Broadcast(devbot, "Usage: "+c.name+" "+c.argsInfo+"  \n"+c.info)
			return
		}
	}
	u.Room.Broadcast("", "This system has been minimized by removing packages and content that are not required on a system that users do not log into.\n\nTo restore this content, including manpages, you can run the 'unminimize' command. You will still need to ensure the 'man-db' package is installed.")
}

func lsCMD(rest string, u *User) error {
	if rest != "" {
		u.Room.Broadcast("", "ls: "+rest+" Permission denied")
		return
	}
	roomList := ""
	for _, r := range rooms {
		roomList += blue.Paint(r.name + "/ ")
	}
	usersList := ""
	for _, us := range u.Room.users {
		usersList += us.name + blue.Paint("/ ")
	}
	u.Room.Broadcast("", "README.md "+usersList+roomList)
}

func commandsCMD(_ string, u *User) error {
	u.Room.Broadcast("", "Commands  \n"+autogenCommands(cmds))
}
