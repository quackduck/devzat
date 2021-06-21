package main

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/shurcooL/tictactoe"
)

type cmd struct {
	name string
	run  func(line string, u *user, isSlack bool)
}

var cmds = []cmd{
	{"./users", usersCMD},
	{"./all", allCMD},
	{"./color", colorCMD},
	{"./exit", exitCMD},
	{"./bell", bellCMD},
	{"./people", peopleCMD},
	{"./help", helpCMD},
	{"./example-code", exampleCodeCMD},
	{"./ascii-art", asciiArtCMD},
	//{"./shrug", shrugCMD},
	{"./emojis", emojisCMD},
	{"./commands", commandsCMD},
	{"./commands-rest", commandsRestCMD},
	{"./clear", clearCMD},
	{"./tic", ticCMD},
	{"./room", roomCMD},
	{"./tz", tzCMD},
	{"./id", idCMD},
	{"./nick", nickCMD},
	{"./banIP", banIPCMD},
	{"./ban", banCMD},
	{"./kick", kickCMD},
	{"ls", lsCMD},
	{"cat README.md", helpCMD},
	{"cat", catCMD},
	{"rm", rmCMD},
	{"easter", easterCMD},
}

// runCommands parses a line of raw input from a user and sends a message as
// required, running any commands the user may have called.
// It also accepts a boolean indicating if the line of input is from slack, in
// which case some commands will not be run (such as ./tz and ./exit)
func runCommands(line string, u *user, isUserSlack bool) {
	defer func() { // crash protection
		if i := recover(); i != nil {
			mainRoom.broadcast(devbot, "Slap the developers in the face for me, the server almost crashed, also tell them this: "+fmt.Sprint(i))
		}
	}()
	if line == "" {
		return
	}
	currCmd := strings.Fields(line)[0]
	if u.messaging != nil && !strings.HasPrefix(line, "=") && !strings.HasPrefix(line, "./room") { // the commands allowed in a private dm room
		dmRoomCMD(line, u, isUserSlack)
		return
	}
	if strings.HasPrefix(line, "=") && !isUserSlack {
		dmCMD(strings.TrimSpace(strings.TrimPrefix(line, "=")), u, isUserSlack)
		return
	}
	if strings.HasPrefix(line, "./hang") { // handles whether or not to print line too
		hangCMD(strings.TrimSpace(strings.TrimPrefix(line, "./hang")), u, isUserSlack)
		return
	}
	if currCmd == "./shrug" {
		shrugCMD(strings.TrimSpace(strings.TrimPrefix(line, "./shrug")), u, isUserSlack)
		return
	}
	if isUserSlack {
		u.room.broadcastNoSlack(u.name, line)
	} else {
		u.room.broadcast(u.name, line)
	}
	devbotChat(u.room, line)
	for _, c := range cmds {
		if c.name == currCmd {
			c.run(strings.TrimSpace(strings.TrimPrefix(line, c.name)), u, isUserSlack)
			return
		}
	}
	if strings.HasPrefix(currCmd, "./") {
		u.room.broadcast(devbot, "wrong command, what's \""+currCmd+"\"")
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
		u.writeln(u.name, "./hang "+rest)
		u.writeln(devbot, "(that word won't show dw)")
		hangGame = &hangman{rest, 15, " "} // default value of guesses so empty space is given away
		u.room.broadcast(devbot, u.name+" has started a new game of Hangman! Guess letters with ./hang <letter>")
		u.room.broadcast(devbot, "```\n"+hangPrint(hangGame)+"\nTries: "+strconv.Itoa(hangGame.triesLeft)+"\n```")
		return
	}
	if !isSlack {
		u.room.broadcast(u.name, "./hang "+rest)
	}

	if strings.Trim(hangGame.word, hangGame.guesses) == "" {
		u.room.broadcast(devbot, "The game has ended. Start a new game with /hang <word>")
		return
	}
	if len(rest) == 0 {
		u.room.broadcast(devbot, "Start a new game with ./hang <word> or guess with /hang <letter>")
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
		u.room.broadcast(devbot, "Play using ./tic <cell num>")
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
		// hasStartedGame = false
	}
}

func allCMD(_ string, u *user, _ bool) {
	names := make([]string, 0, len(allUsers))
	for _, name := range allUsers {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool {
		return strings.ToLower(stripansi.Strip(names[i])) < strings.ToLower(stripansi.Strip(names[j]))
	})
	u.room.broadcast("", fmt.Sprint(names))
}

func easterCMD(_ string, u *user, _ bool) {
	go func() {
		time.Sleep(time.Second)
		u.room.broadcast(devbot, "eggs?")
	}()
}

func exitCMD(_ string, u *user, _ bool) {
	u.close(u.name + red.Paint(" has left the chat"))
}

func bellCMD(_ string, u *user, _ bool) {
	u.bell = !u.bell
	if u.bell {
		u.room.broadcast("", "bell on")
	} else {
		u.room.broadcast("", "bell off")
	}
}

func roomCMD(rest string, u *user, _ bool) {
	if u.messaging != nil {
		u.messaging = nil
		u.writeln(devbot, "Left private chat")
		return
	}
	if rest == "" || rest == "s" { // s so "./rooms" works too
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
	if strings.HasPrefix(rest, "#") {
		if v, ok := rooms[rest]; ok {
			u.changeRoom(v)
		} else {
			rooms[rest] = &room{rest, make([]*user, 0, 10), sync.Mutex{}}
			u.changeRoom(rooms[rest])
		}
		return
	}
	if strings.HasPrefix(rest, "@") {
		restSplit := strings.Fields(strings.TrimPrefix(rest, "@"))
		if len(restSplit) == 0 {
			u.writeln(devbot, "You think people have empty names?")
			return
		}
		peer, ok := findUserByName(u.room, restSplit[0])
		if !ok {
			u.writeln(devbot, "No such person lol, who you wanna dm? (you might be in the wrong room)")
			return
		}
		u.messaging = peer
		u.writeln(devbot, "Now in DMs with "+peer.name+". To leave use ./room")
		return
	}
	u.writeln(devbot, "Rooms need to start with # (public rooms) or @ (dms)")
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
	bansMutex.Lock()
	bans = append(bans, line)
	bansMutex.Unlock()
	saveBansAndUsers()
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
	bansMutex.Lock()
	bans = append(bans, victim.addr)
	bansMutex.Unlock()
	saveBansAndUsers()
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
Belle See  
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

Interesting features:
* Many, many commands. Run ./commands.
* Rooms! Run ./room to see all rooms and use ./room #foo to join a new room.
* Markdown support! Tables, headers, italics and everything. Just use \\n in place of newlines.
* Code syntax highlighting. Use Markdown fences to send code. Run ./example-code to see an example.
* Direct messages! Send a quick DM using =user <msg> or stay in DMs by running ./room @user.
* Timezone support, use ./tz Continent/City to set your timezone.
* Built in Tic Tac Toe and Hangman! Run ./tic or ./hang <word> to start new games.
* Emoji replacements! \:rocket\: => :rocket: (like on Slack and Discord)

For replacing newlines, I often use bulkseotools.com/add-remove-line-breaks.php.

Made by Ishan Goel with feature ideas from friends.  
Thanks to Caleb Denio for lending his server!`)
}

func catCMD(line string, u *user, _ bool) {
	if line == "" {
		u.room.broadcast("", "usage: cat [-benstuv] [file ...]")
	} else {
		u.room.broadcast("", "cat: "+line+": Permission denied")
	}
}

func rmCMD(line string, u *user, _ bool) {
	if line == "" {
		u.room.broadcast("", `usage: rm [-f | -i] [-dPRrvW] file ...
       unlink file`)
	} else {
		u.room.broadcast("", "rm: "+line+": Permission denied, you troll")
	}
}

func exampleCodeCMD(_ string, u *user, _ bool) {
	u.room.broadcast(devbot, "\n```go\npackage main\nimport \"fmt\"\nfunc main() {\n   fmt.Println(\"Example!\")\n}\n```")
}

func asciiArtCMD(_ string, u *user, _ bool) {
	u.room.broadcast("", string(artBytes))
}

func shrugCMD(line string, u *user, _ bool) {
	u.room.broadcast(u.name, line+` ¯\\_(ツ)_/¯`)
}

func emojisCMD(_ string, u *user, _ bool) {
	u.room.broadcast(devbot, "Check out github.com/ikatyang/emoji-cheat-sheet")
}

func commandsRestCMD(_ string, u *user, _ bool) {
	u.room.broadcast("", `All Commands
   ./bell                   _Toggle the ANSI bell used in pings_  
   ./id     <user>          _Get a unique ID for a user (hashed IP)_  
   ./ban    <user>          _Ban <user> (admin)_  
   ./kick   <user>          _Kick <user> (admin)_  
   ./ascii-art              _Show some panda art_  
   ./shrug                  _¯\\_(ツ)_/¯_  
   ./example-code           _Example syntax-highlighted code_  
   ./banIP  <IP/ID>         _Ban by IP or ID (admin)_`)
}

func lsCMD(_ string, u *user, _ bool) {
	u.room.broadcast("", `README.md *users *nick *room *tic *hang *people
*tz *color *all *emojis *exit *help *commands *commands-rest 
*hide *bell *id *ban *kick *ascii-art *shrug *example-code *banIP`)
}

func commandsCMD(_ string, u *user, _ bool) {
	u.room.broadcast("", `Available commands  
   =<user> <msg>            _DM <user> with <msg>_  
   ./users                  _List users_  
   ./nick   <name>          _Change your name_  
   ./room   #<room>         _Join a room or use /room to see all rooms_  
   ./tic    <cell num>      _Play Tic Tac Toe!_  
   ./hang   <char/word>     _Play Hangman!_  
   ./people                 _See info about nice people who joined_  
   ./tz     <zone>          _Change IANA timezone (eg: /tz Asia/Dubai)_  
   ./color  <color>         _Change your name's color_  
   ./all                    _Get a list of all users ever_  
   ./emojis                 _See a list of emojis_  
   ./exit                   _Leave the chat_  
   ./help                   _Show help_  
   ./commands               _Show this message_  
   ./commands-rest          _Uncommon commands list_`)
}
