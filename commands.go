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

// runCommands parses a line of raw input from a user and sends a message as
// required, running any commands the user may have called.
// It also accepts a boolean indicating if the line of input is from slack, in
// which case some commands will not be run (such as /tz and /exit)
func runCommands(line string, u *user, isSlack bool) {
	if line == "" {
		return
	}
	if u.messaging != nil && !strings.HasPrefix(line, "=") && !strings.HasPrefix(line, "./room") {
		u.writeln(u.messaging.name+" <- ", line)
		if u == u.messaging {
			devbotRespond(u.room, []string{"You must be really lonely, DMing yourself.",
				"Don't worry, I won't judge :wink:",
				"srsly?",
				"what an idiot"}, 30, false)
			return
		}
		u.messaging.writeln(u.name+" -> ", line)
		return
	}

	sendToSlack := true
	b := func(senderName, msg string) {
		u.room.broadcast(senderName, msg, true)
	}

	if strings.HasPrefix(line, "./hide") && !isSlack {
		sendToSlack = false
		b = func(senderName, msg string) {
			u.room.broadcast(senderName, msg, false)
		}
	}
	if strings.HasPrefix(line, "=") && !isSlack {
		sendToSlack = false
		b = func(senderName, msg string) {
			u.room.broadcast(senderName, msg, false)
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, "="))
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
				"what an idiot"}, 30, false)
			return
		}
		peer.writeln(u.name+" -> ", msg)
		//peer.writeln(u.name+" -> "+peer.name, msg)
		return
	}
	if strings.HasPrefix(line, "./clear") {
		u.term.Write([]byte("\033[H\033[2J"))
		return
	}
	if strings.HasPrefix(line, "./hang") {
		rest := strings.TrimSpace(strings.TrimPrefix(line, "./hang"))
		if len(rest) > 1 {
			u.writeln(u.name, line)
			u.writeln(devbot, "(that word won't show dw)")
			hangGame = &hangman{rest, 15, " "} // default value of guesses so empty space is given away
			b(devbot, u.name+" has started a new game of Hangman! Guess letters with ./hang <letter>")
			b(devbot, "```\n"+hangPrint(hangGame)+"\nTries: "+strconv.Itoa(hangGame.triesLeft)+"\n```")
			return
		}
		if !isSlack {
			b(u.name, line)
		}

		if strings.Trim(hangGame.word, hangGame.guesses) == "" {
			b(devbot, "The game has ended. Start a new game with /hang <word>")
			return
		}
		if len(rest) == 0 {
			b(devbot, "Start a new game with ./hang <word> or guess with /hang <letter>")
			return
		}
		if hangGame.triesLeft == 0 {
			b(devbot, "No more tries! The word was "+hangGame.word)
			return
		}
		if strings.Contains(hangGame.guesses, rest) {
			b(devbot, "You already guessed "+rest)
			return
		}
		hangGame.guesses += rest

		if !(strings.Contains(hangGame.word, rest)) {
			hangGame.triesLeft--
		}

		display := hangPrint(hangGame)
		b(devbot, "```\n"+display+"\nTries: "+strconv.Itoa(hangGame.triesLeft)+"\n```")

		if strings.Trim(hangGame.word, hangGame.guesses) == "" {
			b(devbot, "You got it! The word was "+hangGame.word)
		} else if hangGame.triesLeft == 0 {
			b(devbot, "No more tries! The word was "+hangGame.word)
		}
		return
	}

	if !isSlack { // actually sends the message
		b(u.name, line)
	}

	//if u == nil { // is slack
	//	devbotChat(mainRoom, line, sendToSlack)
	//} else {
	devbotChat(u.room, line, sendToSlack)
	//}

	if strings.HasPrefix(line, "./tic") {
		rest := strings.TrimSpace(strings.TrimPrefix(line, "./tic"))
		if rest == "" {
			b(devbot, "Starting a new game of Tic Tac Toe! The first player is always X.")
			b(devbot, "Play using ./tic <cell num>")
			currentPlayer = tictactoe.X
			tttGame = new(tictactoe.Board)
			b(devbot, "```\n"+" 1 │ 2 │ 3\n───┼───┼───\n 4 │ 5 │ 6\n───┼───┼───\n 7 │ 8 │ 9\n"+"\n```")
			return
		}

		m, err := strconv.Atoi(rest)
		if err != nil {
			b(devbot, "Make sure you're using a number lol")
			return
		}
		if m < 1 || m > 9 {
			b(devbot, "Moves are numbers between 1 and 9!")
			return
		}
		err = tttGame.Apply(tictactoe.Move(m-1), currentPlayer)
		if err != nil {
			b(devbot, err.Error())
			return
		}
		b(devbot, "```\n"+tttPrint(tttGame.Cells)+"\n```")
		if currentPlayer == tictactoe.X {
			currentPlayer = tictactoe.O
		} else {
			currentPlayer = tictactoe.X
		}
		if !(tttGame.Condition() == tictactoe.NotEnd) {
			b(devbot, tttGame.Condition().String())
			currentPlayer = tictactoe.X
			tttGame = new(tictactoe.Board)
			// hasStartedGame = false
		}
		return
	}

	if line == "./users" {
		b("", printUsersInRoom(u.room))
		return
	}
	if line == "./all" {
		names := make([]string, 0, len(allUsers))
		for _, name := range allUsers {
			names = append(names, name)
		}
		sort.Slice(names, func(i, j int) bool {
			return strings.ToLower(stripansi.Strip(names[i])) < strings.ToLower(stripansi.Strip(names[j]))
		})
		b("", fmt.Sprint(names))
		return
	}
	if line == "easter" {
		go func() {
			time.Sleep(time.Second)
			b(devbot, "eggs?")
		}()
		return
	}
	if line == "./exit" && !isSlack {
		u.close(u.name + red.Paint(" has left the chat"))
		return
	}
	if line == "./bell" && !isSlack {
		u.bell = !u.bell
		if u.bell {
			b("", fmt.Sprint("bell on"))
		} else {
			b("", fmt.Sprint("bell off"))
		}
		return
	}
	if strings.HasPrefix(line, "./room") && !isSlack {
		if u.messaging != nil {
			u.messaging = nil
			u.writeln(devbot, "Left private chat")
			return
		}
		rest := strings.TrimSpace(strings.TrimPrefix(line, "./room"))
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
			b("", "Rooms and users  \n"+strings.TrimSpace(roomsInfo))
			return
		}
		if strings.HasPrefix(rest, "#") {
			if v, ok := rooms[rest]; ok {
				u.changeRoom(v, sendToSlack)
			} else {
				rooms[rest] = &room{rest, make([]*user, 0, 10), sync.Mutex{}}
				u.changeRoom(rooms[rest], sendToSlack)
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
		return

	}
	if strings.HasPrefix(line, "./tz") && !isSlack {
		var err error
		tz := strings.TrimSpace(strings.TrimPrefix(line, "./tz"))
		if tz == "" {
			u.timezone = nil
			return
		}
		u.timezone, err = time.LoadLocation(tz)
		if err != nil {
			b(devbot, "Weird timezone you have there, use Continent/City, EST, PST or see nodatime.org/TimeZones!")
			return
		}
		b(devbot, "Done!")
		return
	}
	if strings.HasPrefix(line, "./id") {
		victim, ok := findUserByName(u.room, strings.TrimSpace(strings.TrimPrefix(line, "./id")))
		if !ok {
			b("", "User not found")
			return
		}
		b("", victim.id)
		return
	}
	if strings.HasPrefix(line, "./nick") && !isSlack {
		u.pickUsername(strings.TrimSpace(strings.TrimPrefix(line, "./nick")))
		return
	}
	if strings.HasPrefix(line, "./name") && isSlack {
		u.pickUsername(strings.TrimSpace(strings.TrimPrefix(line, "./name")))
	}
	if strings.HasPrefix(line, "./banIP") && !isSlack {
		if !auth(u) {
			b(devbot, "Not authorized")
			return
		}
		bansMutex.Lock()
		bans = append(bans, strings.TrimSpace(strings.TrimPrefix(line, "./banIP")))
		bansMutex.Unlock()
		saveBansAndUsers()
		return
	}

	if strings.HasPrefix(line, "./ban") && !isSlack {
		victim, ok := findUserByName(u.room, strings.TrimSpace(strings.TrimPrefix(line, "./ban")))
		if !ok {
			b("", "User not found")
			return
		}
		if !auth(u) {
			b(devbot, "Not authorized")
			return
		}
		bansMutex.Lock()
		bans = append(bans, victim.addr)
		bansMutex.Unlock()
		saveBansAndUsers()
		victim.close(victim.name + " has been banned by " + u.name)
		return
	}
	if strings.HasPrefix(line, "./kick") && !isSlack {
		victim, ok := findUserByName(u.room, strings.TrimSpace(strings.TrimPrefix(line, "./kick")))
		if !ok {
			b("", "User not found")
			return
		}
		if !auth(u) {
			b(devbot, "Not authorized")
			return
		}
		victim.close(victim.name + red.Paint(" has been kicked by ") + u.name)
		return
	}
	if strings.HasPrefix(line, "./color") && !isSlack {
		rest := strings.TrimSpace(strings.TrimPrefix(line, "./color"))
		if rest == "which" {
			b(devbot, "you're using "+u.color)
			return
		}
		if err := u.changeColor(rest); err != nil {
			b(devbot, err.Error())
		}
		return
	}
	if line == "./people" {
		b("", `
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
		return
	}

	if line == "./help" || line == "cat README.md" {
		b("", `Welcome to Devzat! Devzat is chat over SSH: github.com/quackduck/devzat  
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
		return
	}
	if strings.HasPrefix(line, "cat") {
		if line == "cat" {
			b("", "usage: cat [-benstuv] [file ...]")
		} else {
			b("", "cat: "+strings.TrimSpace(strings.TrimPrefix(line, "cat"))+": Permission denied")
		}
		return
	}
	if strings.HasPrefix(line, "rm") {
		if line == "rm" {
			b("", `usage: rm [-f | -i] [-dPRrvW] file ...
       unlink file`)
		} else {
			b("", "rm: "+strings.TrimSpace(strings.TrimPrefix(line, "rm"))+": Permission denied, you troll")
		}
		return
	}
	if line == "./example-code" {
		b(devbot, "\n```go\npackage main\nimport \"fmt\"\nfunc main() {\n   fmt.Println(\"Example!\")\n}\n```")
		return
	}
	if line == "./ascii-art" {
		b("", string(artBytes))
		return
	}
	if line == "./shrug" {
		b("", `¯\\_(ツ)_/¯`)
	}
	if line == "./emojis" {
		b(devbot, "Check out github.com/ikatyang/emoji-cheat-sheet")
		return
	}
	if line == "./commands" {
		b("", `Available commands  
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
		return
	}
	if line == "ls" {
		b("", `README.md *users *nick *room *tic *hang *people
*tz *color *all *emojis *exit *help *commands *commands-rest 
*hide *bell *id *ban *kick *ascii-art *shrug *example-code *banIP`)
	}
	if line == "./commands-rest" {
		b("", `All Commands  
   ./hide                   _Hide messages from HC Slack_  
   ./bell                   _Toggle the ANSI bell used in pings_  
   ./id     <user>          _Get a unique ID for a user (hashed IP)_  
   ./ban    <user>          _Ban <user> (admin)_  
   ./kick   <user>          _Kick <user> (admin)_  
   ./ascii-art              _Show some panda art_  
   ./shrug                  _¯\\_(ツ)_/¯_  
   ./example-code           _Example syntax-highlighted code_  
   ./banIP  <IP/ID>         _Ban by IP or ID (admin)_`)
		return
	}
}
