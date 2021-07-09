package main

import (
	"fmt"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shurcooL/tictactoe"
)

type cmd struct {
	name     string
	run      func(line string, u *user, isSlack bool)
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
		{"emojis", emojisCMD, "", "See a list of emojis"},
		{"bell", bellCMD, "on|off|all", "ANSI bell on pings (on), never (off) or for every message (all)"},
		{"clear", clearCMD, "", "Clear the screen"},
		{"hang", hangCMD, "<char|word>", "Play hangman"}, // won't actually run, here just to show in docs
		{"tic", ticCMD, "<cell num>", "Play tic tac toe!"},
		{"cd", cdCMD, "#room/user", "Join #room, DM user or run cd to see a list"}, // won't actually run, here just to show in docs
		{"tz", tzCMD, "<zone>", "Set your IANA timezone (like tz Asia/Dubai)"},
		{"nick", nickCMD, "<name>", "Change your username"},
		{"rest", commandsRestCMD, "", "Uncommon commands list"}}
	cmdsRest = []cmd{
		{"people", peopleCMD, "", "See info about nice people who joined"},
		{"id", idCMD, "<user>", "Get a unique ID for a user (hashed IP)"},
		{"eg-code", exampleCodeCMD, "", "Example syntax-highlighted code"},
		{"banIP", banIPCMD, "<IP>", "Ban an IP (admin)"},
		{"ban", banCMD, "<user>", "Ban <user> (admin)"},
		{"kick", kickCMD, "<user>", "Kick <user> (admin)"},
		{"art", asciiArtCMD, "", "Show some panda art"},
		{"shrug", shrugCMD, "", `¯\\_(ツ)_/¯`}} // won't actually run, here just to show in docs
	secretCMDs = []cmd{
		{"ls", lsCMD, "", ""},
		{"cat", catCMD, "", ""},
		{"rm", rmCMD, "", ""}}
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
func runCommands(line string, u *user, isUserSlack bool) {
	if line == "" {
		return
	}
	defer func() { // crash protection
		if i := recover(); i != nil {
			mainRoom.broadcast(devbot, "Slap the developers in the face for me, the server almost crashed, also tell them this: "+fmt.Sprint(i, "\n"+string(debug.Stack())))
		}
	}()
	currCmd := strings.Fields(line)[0]
	if u.messaging != nil && currCmd != "=" && currCmd != "cd" && currCmd != "exit" { // the commands allowed in a private dm room
		dmRoomCMD(line, u, isUserSlack)
		return
	}
	if strings.HasPrefix(line, "=") && !isUserSlack {
		dmCMD(strings.TrimSpace(strings.TrimPrefix(line, "=")), u, isUserSlack)
		return
	}

	switch currCmd {
	case "hang":
		hangCMD(strings.TrimSpace(strings.TrimPrefix(line, "hang")), u, isUserSlack)
		return
	case "cd":
		cdCMD(strings.TrimSpace(strings.TrimPrefix(line, "cd")), u, isUserSlack)
		return
	case "shrug":
		shrugCMD(strings.TrimSpace(strings.TrimPrefix(line, "shrug")), u, isUserSlack)
		return
	}

	if isUserSlack {
		u.room.broadcastNoSlack(u.name, line)
	} else {
		u.room.broadcast(u.name, line)
	}

	devbotChat(u.room, line)

	for _, c := range allcmds {
		if c.name == currCmd {
			c.run(strings.TrimSpace(strings.TrimPrefix(line, c.name)), u, isUserSlack)
			return
		}
	}
}

func dmCMD(rest string, u *user, _ bool) {
	restSplit := strings.Fields(rest)
	if len(restSplit) < 2 {
		u.writeln(devbot, "You gotta have a message mate")
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

func hangCMD(rest string, u *user, isSlack bool) {
	if len(rest) > 1 {
		if !isSlack {
			u.writeln(u.name, "hang "+rest)
			u.writeln(devbot, "(that word won't show dw)")
		}
		hangGame = &hangman{rest, 15, " "} // default value of guesses so empty space is given away
		u.room.broadcast(devbot, u.name+" has started a new game of Hangman! Guess letters with hang <letter>")
		u.room.broadcast(devbot, "```\n"+hangPrint(hangGame)+"\nTries: "+strconv.Itoa(hangGame.triesLeft)+"\n```")
		return
	}
	if !isSlack {
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

func clearCMD(_ string, u *user, _ bool) {
	u.term.Write([]byte("\033[H\033[2J"))
}

func usersCMD(_ string, u *user, _ bool) {
	u.room.broadcast("", printUsersInRoom(u.room))
}

func dmRoomCMD(line string, u *user, _ bool) {
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

func ticCMD(rest string, u *user, _ bool) {
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

func exitCMD(_ string, u *user, _ bool) {
	u.close(u.name + red.Paint(" has left the chat"))
}

func bellCMD(rest string, u *user, _ bool) {
	switch rest {
	case "off":
		u.bell = false
		u.room.broadcast("", "bell off (never)")
	case "on":
		u.bell = true
		u.room.broadcast("", "bell on (pings)")
	case "all":
		u.pingEverytime = true
		u.room.broadcast("", "bell all (every message)")
	}
}

func cdCMD(rest string, u *user, _ bool) {
	if u.messaging != nil {
		u.messaging = nil
		u.writeln(devbot, "Left private chat")
		if rest == "" || rest == ".." {
			return
		}
	}
	if strings.HasPrefix(rest, "#") {
		u.writeln(u.name, "cd "+rest)
		if v, ok := rooms[rest]; ok {
			u.changeRoom(v)
		} else {
			rooms[rest] = &room{rest, make([]*user, 0, 10), sync.Mutex{}}
			u.changeRoom(rooms[rest])
		}
		return
	}
	if rest == "" {
		u.writeln(u.name, "cd "+rest)
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
		u.writeln(devbot, "No such person lol, who you wanna dm? (you might be in the wrong room)")
		return
	}
	u.messaging = peer
	u.writeln(devbot, "Now in DMs with "+peer.name+". To leave use cd ..")
}

func tzCMD(tz string, u *user, _ bool) {
	var err error
	if tz == "" {
		u.timezone = nil
		return
	}
	u.timezone, err = time.LoadLocation(tz)
	if err != nil {
		u.room.broadcast(devbot, "Weird timezone you have there, use Continent/City, EST, PST or see nodatime.org/TimeZones!")
		return
	}
	u.room.broadcast(devbot, "Done!")
}

func idCMD(line string, u *user, _ bool) {
	victim, ok := findUserByName(u.room, line)
	if !ok {
		u.room.broadcast("", "User not found")
		return
	}
	u.room.broadcast("", victim.id)
}

func nickCMD(line string, u *user, _ bool) {
	u.pickUsername(line)
	return
}

func banIPCMD(line string, u *user, _ bool) {
	if !auth(u) {
		u.room.broadcast(devbot, "Not authorized")
		return
	}
	bans = append(bans, line)
	saveBans()
}

func banCMD(line string, u *user, _ bool) {
	victim, ok := findUserByName(u.room, line)
	if !ok {
		u.room.broadcast("", "User not found")
		return
	}
	if !auth(u) {
		u.room.broadcast(devbot, "Not authorized")
		return
	}
	//bansMutex.Lock()
	bans = append(bans, victim.addr)
	//bansMutex.Unlock()
	saveBans()
	victim.close(victim.name + " has been banned by " + u.name)
}

func kickCMD(line string, u *user, _ bool) {
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

func colorCMD(rest string, u *user, _ bool) {
	if rest == "which" {
		u.room.broadcast(devbot, "you're using "+u.color)
	} else if err := u.changeColor(rest); err != nil {
		u.room.broadcast(devbot, err.Error())
	}
}

func peopleCMD(_ string, u *user, _ bool) {
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
Belle See, Fayd  
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

func helpCMD(_ string, u *user, _ bool) {
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

Made by Ishan Goel with feature ideas from friends.  
Thanks to Caleb Denio for lending his server!`)
}

func catCMD(line string, u *user, isSlack bool) {
	if line == "" {
		u.room.broadcast("", "usage: cat [-benstuv] [file ...]")
	} else if line == "README.md" {
		helpCMD(line, u, isSlack)
	} else {
		u.room.broadcast("", "cat: "+line+": Permission denied")
	}
}

func rmCMD(line string, u *user, _ bool) {
	if line == "" {
		u.room.broadcast("", `usage: rm [-f | -i] [-dPRrvW] file ...
unlink file`)
	} else {
		u.room.broadcast("", "rm: "+line+": Permission denied, sucker")
	}
}

func exampleCodeCMD(_ string, u *user, _ bool) {
	u.room.broadcast(devbot, "\n```go\npackage main\nimport \"fmt\"\nfunc main() {\n   fmt.Println(\"Example!\")\n}\n```")
}

func asciiArtCMD(_ string, u *user, _ bool) {
	u.room.broadcast("", art)
}

func shrugCMD(line string, u *user, _ bool) {
	u.room.broadcast(u.name, line+` ¯\\_(ツ)_/¯`)
}

func emojisCMD(_ string, u *user, _ bool) {
	u.room.broadcast(devbot, "Check out github.com/ikatyang/emoji-cheat-sheet")
}

func commandsRestCMD(_ string, u *user, _ bool) {
	u.room.broadcast("", "The rest  \n"+autogenCommands(cmdsRest))
}

func lsCMD(rest string, u *user, _ bool) {
	if rest != "" {
		u.room.broadcast("", "ls: "+rest+" Permission denied")
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

func commandsCMD(_ string, u *user, _ bool) {
	u.room.broadcast("", "Commands  \n"+autogenCommands(cmds))
}
