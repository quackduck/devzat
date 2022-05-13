package main

import (
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math/rand"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alecthomas/chroma"
	chromastyles "github.com/alecthomas/chroma/styles"
	markdown "github.com/quackduck/go-term-markdown"
	"github.com/quackduck/term"
	"github.com/shurcooL/tictactoe"
)

type CMD struct {
	name     string
	run      func(line string, u *User)
	argsInfo string
	info     string
}

var (
	MainCMDs = []CMD{
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
		{"devmonk", devmonkCMD, "", "Test your typing speed"},
		{"cd", cdCMD, "#room|user", "Join #room, DM user or run cd to see a list"}, // won't actually run, here just to show in docs
		{"tz", tzCMD, "<zone> [24h]", "Set your IANA timezone (like tz Asia/Dubai) and optionally set 24h"},
		{"nick", nickCMD, "<name>", "Change your username"},
		{"pronouns", pronounsCMD, "@user|pronouns", "Set your pronouns or get another user's"},
		{"theme", themeCMD, "<theme>|list", "Change the syntax highlighting theme"},
		{"rest", commandsRestCMD, "", "Uncommon commands list"}}
	RestCMDs = []CMD{
		// {"people", peopleCMD, "", "See info about nice people who joined"},
		{"bio", bioCMD, "[user]", "Get a user's bio or set yours"},
		{"id", idCMD, "<user>", "Get a unique ID for a user (hashed key)"},
		{"admins", adminsCMD, "", "Print the ID (hashed key) for all admins"},
		{"eg-code", exampleCodeCMD, "[big]", "Example syntax-highlighted code"},
		{"lsbans", listBansCMD, "", "List banned IDs"},
		{"ban", banCMD, "<user>", "Ban <user> (admin)"},
		{"unban", unbanCMD, "<IP|ID> [dur]", "Unban a person and optionally, for a duration (admin)"},
		{"kick", kickCMD, "<user>", "Kick <user> (admin)"},
		{"art", asciiArtCMD, "", "Show some panda art"},
		{"pwd", pwdCMD, "", "Show your current room"},
		//		{"sixel", sixelCMD, "<png url>", "Render an image in high quality"},
		{"shrug", shrugCMD, "", `¯\\\_(ツ)\_/¯`}, // won't actually run, here just to show in docs
	}
	SecretCMDs = []CMD{
		{"ls", lsCMD, "???", "???"},
		{"cat", catCMD, "???", "???"},
		{"rm", rmCMD, "???", "???"},
		{"su", nickCMD, "???", "This is an alias of nick"},
		{"colour", colorCMD, "???", "This is an alias of color"}, // appease the british
	}
)

const (
	MaxRoomNameLen = 30
	MaxBioLen      = 300
)

func init() {
	MainCMDs = append(MainCMDs, CMD{"cmds", commandsCMD, "", "Show this message"}) // avoid initialization loop
}

// runCommands parses a line of raw input from a User and sends a message as
// required, running any commands the User may have called.
// It also accepts a boolean indicating if the line of input is from slack, in
// which case some commands will not be run (such as ./tz and ./exit)
func runCommands(line string, u *User) {
	line = rmBadWords(line)

	if line == "" {
		return
	}
	defer func() { // crash protection
		if i := recover(); i != nil {
			MainRoom.broadcast(Devbot, "Slap the developers in the face for me, the server almost crashed, also tell them this: "+fmt.Sprint(i, "\n"+string(debug.Stack())))
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

	// Now we know it is not a DM, so this is a safe place to add the hook for sending the event to plugins
	sendMessageToPlugins(line, u)

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
		u.room.broadcastNoSlack(u.Name, line)
	} else {
		u.room.broadcast(u.Name, line)
	}

	devbotChat(u.room, line)

	args := strings.TrimSpace(strings.TrimPrefix(line, currCmd))

	if runPluginCMDs(u, currCmd, args) {
		return
	}

	if cmd, ok := getCMD(currCmd); ok {
		cmd.run(args, u)
	}
}

func dmCMD(rest string, u *User) {
	restSplit := strings.Fields(rest)
	if len(restSplit) < 2 {
		u.writeln(Devbot, "You gotta have a message, mate")
		return
	}
	peer, ok := findUserByName(u.room, restSplit[0])
	if !ok {
		u.writeln(Devbot, "No such person lol, who you wanna dm? (you might be in the wrong room)")
		return
	}
	msg := strings.TrimSpace(strings.TrimPrefix(rest, restSplit[0]))
	u.writeln(peer.Name+" <- ", msg)
	if u == peer {
		devbotRespond(u.room, []string{"You must be really lonely, DMing yourself.",
			"Don't worry, I won't judge :wink:",
			"srsly?",
			"what an idiot"}, 30)
		return
	}
	peer.writeln(u.Name+" -> ", msg)
}

func hangCMD(rest string, u *User) {
	if len(rest) > 1 {
		if !u.isSlack {
			u.writeln(u.Name, "hang "+rest)
			u.writeln(Devbot, "(that word won't show dw)")
		}
		hangGame = &hangman{rest, 15, " "} // default value of guesses so empty space is given away
		u.room.broadcast(Devbot, u.Name+" has started a new game of Hangman! Guess letters with hang <letter>")
		u.room.broadcast(Devbot, "```\n"+hangPrint(hangGame)+"\nTries: "+strconv.Itoa(hangGame.triesLeft)+"\n```")
		return
	}
	if !u.isSlack {
		u.room.broadcast(u.Name, "hang "+rest)
	}
	if strings.Trim(hangGame.word, hangGame.guesses) == "" {
		u.room.broadcast(Devbot, "The game has ended. Start a new game with hang <word>")
		return
	}
	if len(rest) == 0 {
		u.room.broadcast(Devbot, "Start a new game with hang <word> or guess with hang <letter>")
		return
	}
	if hangGame.triesLeft == 0 {
		u.room.broadcast(Devbot, "No more tries! The word was "+hangGame.word)
		return
	}
	if strings.Contains(hangGame.guesses, rest) {
		u.room.broadcast(Devbot, "You already guessed "+rest)
		return
	}
	hangGame.guesses += rest
	if !(strings.Contains(hangGame.word, rest)) {
		hangGame.triesLeft--
	}
	display := hangPrint(hangGame)
	u.room.broadcast(Devbot, "```\n"+display+"\nTries: "+strconv.Itoa(hangGame.triesLeft)+"\n```")
	if strings.Trim(hangGame.word, hangGame.guesses) == "" {
		u.room.broadcast(Devbot, "You got it! The word was "+hangGame.word)
	} else if hangGame.triesLeft == 0 {
		u.room.broadcast(Devbot, "No more tries! The word was "+hangGame.word)
	}
}

func clearCMD(_ string, u *User) {
	u.term.Write([]byte("\033[H\033[2J"))
}

func usersCMD(_ string, u *User) {
	u.room.broadcast("", printUsersInRoom(u.room))
}

func dmRoomCMD(line string, u *User) {
	u.writeln(u.messaging.Name+" <- ", line)
	if u == u.messaging {
		devbotRespond(u.room, []string{"You must be really lonely, DMing yourself.",
			"Don't worry, I won't judge :wink:",
			"srsly?",
			"what an idiot"}, 30)
		return
	}
	u.messaging.writeln(u.Name+" -> ", line)
}

// named devmonk at the request of a certain ced
func devmonkCMD(_ string, u *User) {
	sentences := []string{"I really want to go to work, but I am too sick to drive.", "The fence was confused about whether it was supposed to keep things in or keep things out.", "He found the end of the rainbow and was surprised at what he found there.", "He had concluded that pigs must be able to fly in Hog Heaven.", "I just wanted to tell you I could see the love you have for your child by the way you look at her.", "We will not allow you to bring your pet armadillo along.", "The father died during childbirth.", "I covered my friend in baby oil.", "Cursive writing is the best way to build a race track.", "My Mum tries to be cool by saying that she likes all the same things that I do.", "The sky is clear; the stars are twinkling.", "Flash photography is best used in full sunlight.", "The rusty nail stood erect, angled at a 45-degree angle, just waiting for the perfect barefoot to come along.", "People keep telling me \"orange\" but I still prefer \"pink\".", "Peanut butter and jelly caused the elderly lady to think about her past.", "She always had an interesting perspective on why the world must be flat.", "People who insist on picking their teeth with their elbows are so annoying!", "Joe discovered that traffic cones make excellent megaphones.", "They say people remember important moments in their life well, yet no one even remembers their own birth.", "Purple is the best city in the forest.", "The book is in front of the table.", "Everyone was curious about the large white blimp that appeared overnight.", "He wondered if she would appreciate his toenail collection.", "Situps are a terrible way to end your day.", "He barked orders at his daughters but they just stared back with amusement.", "She couldn't decide of the glass was half empty or half full so she drank it.", "It caught him off guard that space smelled of seared steak.", "There are few things better in life than a slice of pie.", "After exploring the abandoned building, he started to believe in ghosts.", "This is a Japanese doll.", "I've never seen a more beautiful brandy glass filled with wine.", "Don't piss in my garden and tell me you're trying to help my plants grow.", "She looked at the masterpiece hanging in the museum but all she could think is that her five-year-old could do better.", "Nobody loves a pig wearing lipstick.", "She always speaks to him in a loud voice.", "The teens wondered what was kept in the red shed on the far edge of the school grounds.", "I'll have you know I've written over fifty novels", "He didn't understand why the bird wanted to ride the bicycle.", "Potato wedges probably are not best for relationships.", "Baby wipes are made of chocolate stardust.", "Lucifer was surprised at the amount of life at Death Valley.", "She was too busy always talking about what she wanted to do to actually do any of it.", "The sudden rainstorm washed crocodiles into the ocean.", "I used to live in my neighbor's fishpond, but the aesthetic wasn't to my taste.", "He kept telling himself that one day it would all somehow make sense.", "The random sentence generator generated a random sentence about a random sentence.", "The reservoir water level continued to lower while we enjoyed our long shower.", "A song can make or ruin a person’s day if they let it get to them.", "He stomped on his fruit loops and thus became a cereal killer.", "I know many children ask for a pony, but I wanted a bicycle with rockets strapped to it."}
	text := sentences[rand.Intn(len(sentences))]
	u.writeln(Devbot, "Okay type this text: \n\n> "+text)
	u.term.SetPrompt("> ")
	defer u.term.SetPrompt(u.Name + ": ")
	start := time.Now()
	line, err := u.term.ReadLine()
	if err == term.ErrPasteIndicator { // TODO: doesn't work for some reason?
		u.room.broadcast(Devbot, "SMH did you know that "+u.Name+" tried to cheat in a typing game?")
		return
	}
	dur := time.Since(start)

	accuracy := 100.0
	// analyze correctness
	if line != text {
		wrongWords := 0
		correct := strings.Fields(line)
		test := strings.Fields(text)
		if len(correct) > len(test) {
			wrongWords += len(correct) - len(test)
			correct = correct[:len(test)]
		} else {
			wrongWords += len(test) - len(correct)
			test = test[:len(correct)]
		}
		for i := 0; i < len(correct); i++ {
			if correct[i] != test[i] {
				wrongWords++
			}
		}
		accuracy -= 100 * float64(wrongWords) / float64(len(test))
		if accuracy < 0.0 {
			accuracy = 0.0
		}
	}

	u.room.broadcast(Devbot, "Okay "+u.Name+", you typed that in "+dur.Truncate(time.Second/10).String()+" so your speed is "+
		strconv.FormatFloat(
			float64(len(strings.Fields(text)))/dur.Minutes(), 'f', 1, 64,
		)+" wpm"+" with accuracy "+strconv.FormatFloat(accuracy, 'f', 1, 64)+"%",
	)
}

func ticCMD(rest string, u *User) {
	if rest == "" {
		u.room.broadcast(Devbot, "Starting a new game of Tic Tac Toe! The first player is always X.")
		u.room.broadcast(Devbot, "Play using tic <cell num>")
		currentPlayer = tictactoe.X
		tttGame = new(tictactoe.Board)
		u.room.broadcast(Devbot, "```\n"+" 1 │ 2 │ 3\n───┼───┼───\n 4 │ 5 │ 6\n───┼───┼───\n 7 │ 8 │ 9\n"+"\n```")
		return
	}
	m, err := strconv.Atoi(rest)
	if err != nil {
		u.room.broadcast(Devbot, "Make sure you're using a number lol")
		return
	}
	if m < 1 || m > 9 {
		u.room.broadcast(Devbot, "Moves are numbers between 1 and 9!")
		return
	}
	err = tttGame.Apply(tictactoe.Move(m-1), currentPlayer)
	if err != nil {
		u.room.broadcast(Devbot, err.Error())
		return
	}
	u.room.broadcast(Devbot, "```\n"+tttPrint(tttGame.Cells)+"\n```")
	if currentPlayer == tictactoe.X {
		currentPlayer = tictactoe.O
	} else {
		currentPlayer = tictactoe.X
	}
	if !(tttGame.Condition() == tictactoe.NotEnd) {
		u.room.broadcast(Devbot, tttGame.Condition().String())
		currentPlayer = tictactoe.X
		tttGame = new(tictactoe.Board)
	}
}

func exitCMD(_ string, u *User) {
	u.close(u.Name + Red.Paint(" has left the chat"))
}

func bellCMD(rest string, u *User) {
	switch rest {
	case "off":
		u.Bell = false
		u.PingEverytime = false
		u.room.broadcast("", "bell off (never)")
	case "on":
		u.Bell = true
		u.PingEverytime = false
		u.room.broadcast("", "bell on (pings)")
	case "all":
		u.Bell = true
		u.PingEverytime = true
		u.room.broadcast("", "bell all (every message)")
	case "", "status":
		if u.PingEverytime {
			u.room.broadcast("", "bell all (every message)")
		} else if u.Bell {
			u.room.broadcast("", "bell on (pings)")
		} else { // bell is off
			u.room.broadcast("", "bell off (never)")
		}
	default:
		u.room.broadcast(Devbot, "your options are off, on and all")
	}
}

func cdCMD(rest string, u *User) {
	if u.messaging != nil {
		u.messaging = nil
		u.writeln(Devbot, "Left private chat")
		if rest == "" || rest == ".." {
			return
		}
	}
	if rest == ".." { // cd back into the main room
		u.room.broadcast(u.Name, "cd "+rest)
		if u.room != MainRoom {
			u.changeRoom(MainRoom)
		}
		return
	}
	if strings.HasPrefix(rest, "#") {
		u.room.broadcast(u.Name, "cd "+rest)
		if len(rest) > MaxRoomNameLen {
			rest = rest[0:MaxRoomNameLen]
			u.room.broadcast(Devbot, "Room name lengths are limited, so I'm shortening it to "+rest+".")
		}
		if v, ok := Rooms[rest]; ok {
			u.changeRoom(v)
		} else {
			Rooms[rest] = &Room{rest, make([]*User, 0, 10), sync.Mutex{}}
			u.changeRoom(Rooms[rest])
		}
		return
	}
	if rest == "" {
		u.room.broadcast(u.Name, "cd "+rest)
		type kv struct {
			roomName   string
			numOfUsers int
		}
		var ss []kv
		for k, v := range Rooms {
			ss = append(ss, kv{k, len(v.users)})
		}
		sort.Slice(ss, func(i, j int) bool {
			return ss[i].numOfUsers > ss[j].numOfUsers
		})
		roomsInfo := ""
		for _, kv := range ss {
			roomsInfo += Blue.Paint(kv.roomName) + ": " + printUsersInRoom(Rooms[kv.roomName]) + "  \n"
		}
		u.room.broadcast("", "Rooms and users  \n"+strings.TrimSpace(roomsInfo))
		return
	}
	name := strings.Fields(rest)[0]
	if len(name) == 0 {
		u.writeln(Devbot, "You think people have empty names?")
		return
	}
	peer, ok := findUserByName(u.room, name)
	if !ok {
		u.writeln(Devbot, "No such person lol, who do you want to dm? (you might be in the wrong room)")
		return
	}
	u.messaging = peer
	u.writeln(Devbot, "Now in DMs with "+peer.Name+". To leave use cd ..")
}

func tzCMD(tzArg string, u *User) {
	var err error
	if tzArg == "" {
		u.Timezone.Location = nil
		return
	}
	tzArgList := strings.Fields(tzArg)
	tz := tzArgList[0]
	switch strings.ToUpper(tz) {
	case "PST", "PDT":
		tz = "PST8PDT"
	case "CST", "CDT":
		tz = "CST6CDT"
	case "EST", "EDT":
		tz = "EST5EDT"
	case "MT":
		tz = "America/Phoenix"
	}
	u.Timezone.Location, err = time.LoadLocation(tz)
	if err != nil {
		u.room.broadcast(Devbot, "Weird timezone you have there, use the format Continent/City, the usual US timezones (PST, PDT, EST, EDT...) or check nodatime.org/TimeZones!")
		return
	}
	if len(tzArgList) == 2 {
		u.FormatTime24 = tzArgList[1] == "24h"
	} else {
		u.FormatTime24 = false
	}
	u.room.broadcast(Devbot, "Changed your timezone!")
}

func bioCMD(line string, u *User) {
	if line == "" {
		u.writeln(Devbot, "Your current bio is:  \n> "+u.Bio)
		u.term.SetPrompt("> ")
		defer u.term.SetPrompt(u.Name + ": ")
		for {
			input, err := u.term.ReadLine()
			if err != nil {
				return
			}
			input = strings.TrimSpace(input)
			if input != "" {
				if len(input) > MaxBioLen {
					u.writeln(Devbot, "Your bio is too long. It shouldn't be more than "+strconv.Itoa(MaxBioLen)+" characters.")
				}
				u.Bio = input
				// make sure it gets saved now so it stays even if the server crashes
				u.savePrefs() //nolint:errcheck // best effort
				return
			}
		}
	}
	target, ok := findUserByName(u.room, line)
	if !ok {
		u.room.broadcast(Devbot, "Who???")
		return
	}
	u.room.broadcast("", target.Bio)
}

func idCMD(line string, u *User) {
	victim, ok := findUserByName(u.room, line)
	if !ok {
		u.room.broadcast("", "User not found")
		return
	}
	u.room.broadcast("", victim.id)
}

func nickCMD(line string, u *User) {
	u.pickUsername(line) //nolint:errcheck // if reading input fails, the next repl will err out
}

func listBansCMD(_ string, u *User) {
	msg := "Bans by ID:  \n"
	for i := 0; i < len(Bans); i++ {
		msg += Cyan.Cyan(strconv.Itoa(i+1)) + ". " + Bans[i].ID + "  \n"
	}
	u.room.broadcast(Devbot, msg)
}

func unbanCMD(toUnban string, u *User) {
	if !auth(u) {
		u.room.broadcast(Devbot, "Not authorized")
		return
	}

	if unbanIDorIP(toUnban) {
		u.room.broadcast(Devbot, "Unbanned person: "+toUnban)
		saveBans()
	} else {
		u.room.broadcast(Devbot, "I couldn't find that person")
	}
}

// unbanIDorIP unbans an ID or an IP, but does NOT save bans to the bans file.
// It returns whether the person was found, and so, whether the bans slice was modified.
func unbanIDorIP(toUnban string) bool {
	for i := 0; i < len(Bans); i++ {
		if Bans[i].ID == toUnban || Bans[i].Addr == toUnban { // allow unbanning by either ID or IP
			// remove this ban
			Bans = append(Bans[:i], Bans[i+1:]...)
			return true
		}
	}
	return false
}

func banCMD(line string, u *User) {
	if !auth(u) {
		u.room.broadcast(Devbot, "Not authorized")
		return
	}
	split := strings.Split(line, " ")
	if len(split) == 0 {
		u.room.broadcast(Devbot, "Which user do you want to ban?")
		return
	}
	victim, ok := findUserByName(u.room, split[0])
	if !ok {
		u.room.broadcast("", "User not found")
		return
	}
	// check if the ban is for a certain duration
	if len(split) > 1 {
		dur, err := time.ParseDuration(split[1])
		if err != nil {
			u.room.broadcast(Devbot, "I couldn't parse that as a duration")
			return
		}
		Bans = append(Bans, Ban{victim.addr, victim.id})
		victim.close(victim.Name + " has been banned by " + u.Name + " for " + dur.String())
		go func(id string) {
			time.Sleep(dur)
			unbanIDorIP(id)
		}(victim.id) // evaluate id now, call unban with that value later
		return
	}
	banUser(u.Name, victim)
}

func banUser(banner string, victim *User) {
	Bans = append(Bans, Ban{victim.addr, victim.id})
	saveBans()
	victim.close(victim.Name + " has been banned by " + banner)
}

func kickCMD(line string, u *User) {
	victim, ok := findUserByName(u.room, line)
	if !ok {
		u.room.broadcast("", "User not found")
		return
	}
	if !auth(u) {
		u.room.broadcast(Devbot, "Not authorized")
		return
	}
	victim.close(victim.Name + Red.Paint(" has been kicked by ") + u.Name)
}

func colorCMD(rest string, u *User) {
	if rest == "which" {
		u.room.broadcast(Devbot, "fg: "+u.Color+" & bg: "+u.ColorBG)
	} else if err := u.changeColor(rest); err != nil {
		u.room.broadcast(Devbot, err.Error())
	}
}

func adminsCMD(_ string, u *User) {
	msg := "Admins by ID:  \n"
	i := 1
	for id, info := range Config.Admins {
		if len(id) > 10 {
			id = id[:10] + "..."
		}
		msg += Cyan.Cyan(strconv.Itoa(i)) + ". " + id + "\t" + info + "  \n"
		i++
	}
	u.room.broadcast(Devbot, msg)
}

func helpCMD(_ string, u *User) {
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

func catCMD(line string, u *User) {
	if line == "" {
		u.room.broadcast("", "usage: cat [-benstuv] [file ...]")
	} else if line == "README.md" {
		helpCMD(line, u)
	} else {
		u.room.broadcast("", "cat: "+line+": Permission denied")
	}
}

func rmCMD(line string, u *User) {
	if line == "" {
		u.room.broadcast("", `usage: rm [-f | -i] [-dPRrvW] file ...
unlink file`)
	} else {
		u.room.broadcast("", "rm: "+line+": Permission denied, sucker")
	}
}

func exampleCodeCMD(line string, u *User) {
	if line == "big" {
		u.room.broadcast(Devbot, "```go\npackage main\n\nimport \"fmt\"\n\nfunc sum(nums ...int) {\n    fmt.Print(nums, \" \")\n    total := 0\n    for _, num := range nums {\n        total += num\n    }\n    fmt.Println(total)\n}\n\nfunc main() {\n\n    sum(1, 2)\n    sum(1, 2, 3)\n\n    nums := []int{1, 2, 3, 4}\n    sum(nums...)\n}\n```")
		return
	}
	u.room.broadcast(Devbot, "\n```go\npackage main\nimport \"fmt\"\nfunc main() {\n   fmt.Println(\"Example!\")\n}\n```")
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

	chromastyles.Register(chroma.MustNewStyle("blackbird", chroma.StyleEntries{chroma.Text: white, chroma.Error: red, chroma.Comment: gray, chroma.Keyword: redItalic, chroma.KeywordNamespace: redItalic, chroma.KeywordType: tealItalic, chroma.Operator: blue, chroma.Punctuation: white, chroma.Name: white, chroma.NameAttribute: white, chroma.NameClass: green, chroma.NameConstant: tealItalic, chroma.NameDecorator: green, chroma.NameException: red, chroma.NameFunction: green, chroma.NameOther: white, chroma.NameTag: yellow, chroma.LiteralNumber: blue, chroma.Literal: yellow, chroma.LiteralDate: yellow, chroma.LiteralString: yellow, chroma.LiteralStringEscape: teal, chroma.GenericDeleted: red, chroma.GenericEmph: "italic", chroma.GenericInserted: green, chroma.GenericStrong: "bold", chroma.GenericSubheading: yellow, chroma.Background: "bg:#000000"}))
}

func themeCMD(line string, u *User) {
	if line == "list" {
		u.room.broadcast(Devbot, "Available themes: "+strings.Join(chromastyles.Names(), ", "))
		return
	}
	for _, name := range chromastyles.Names() {
		if name == line {
			markdown.CurrentTheme = chromastyles.Get(name)
			u.room.broadcast(Devbot, "Theme set to "+name)
			return
		}
	}
	u.room.broadcast(Devbot, "What theme is that? Use theme list to see what's available.")
}

func asciiArtCMD(_ string, u *User) {
	u.room.broadcast("", Art)
}

func pwdCMD(_ string, u *User) {
	if u.messaging != nil {
		u.writeln("", u.messaging.Name)
	} else {
		u.room.broadcast("", u.room.name)
	}
}

func shrugCMD(line string, u *User) {
	u.room.broadcast(u.Name, line+` ¯\\_(ツ)_/¯`)
}

func pronounsCMD(line string, u *User) {
	args := strings.Fields(line)

	if line == "" {
		u.room.broadcast(Devbot, "Set pronouns by providing em or query a user's pronouns!")
		return
	}

	if len(args) == 1 && strings.HasPrefix(args[0], "@") {
		victim, ok := findUserByName(u.room, args[0][1:])
		if !ok {
			u.room.broadcast(Devbot, "Who's that?")
			return
		}
		u.room.broadcast(Devbot, victim.Name+"'s pronouns are "+victim.displayPronouns())
		return
	}

	u.Pronouns = strings.Fields(strings.ReplaceAll(strings.ToLower(line), "\n", ""))
	//u.changeColor(u.Color) // refresh pronouns
	u.room.broadcast(Devbot, u.Name+" now goes by "+u.displayPronouns())
}

func emojisCMD(_ string, u *User) {
	u.room.broadcast(Devbot, "Check out https\\://github.com/ikatyang/emoji-cheat-sheet")
}

func commandsRestCMD(_ string, u *User) {
	u.room.broadcast("", "The rest  \n"+autogenCommands(RestCMDs))
}

func manCMD(rest string, u *User) {
	if rest == "" {
		u.room.broadcast(Devbot, "What command do you want help with?")
		return
	}

	if cmd, ok := getCMD(rest); ok {
		u.room.broadcast(Devbot, "Usage: "+cmd.name+" "+cmd.argsInfo+"  \n"+cmd.info)
	}
	// Plugin commands
	if c, ok := PluginCMDs[rest]; ok {
		u.room.broadcast(Devbot, "Usage: "+rest+" "+c.argsInfo+"  \n"+c.info)
		return
	}

	u.room.broadcast("", "This system has been minimized by removing packages and content that are not required on a system that users do not log into.\n\nTo restore this content, including manpages, you can run the 'unminimize' command. You will still need to ensure the 'man-db' package is installed.")
}

func lsCMD(rest string, u *User) {
	if len(rest) > 0 && rest[0] == '#' {
		if r, ok := Rooms[rest]; ok {
			usersList := ""
			for _, us := range r.users {
				usersList += us.Name + Blue.Paint("/ ")
			}
			u.room.broadcast("", usersList)
			return
		}
	}
	if rest != "" {
		u.room.broadcast("", "ls: "+rest+" Permission denied")
		return
	}
	roomList := ""
	for _, r := range Rooms {
		roomList += Blue.Paint(r.name + "/ ")
	}
	usersList := ""
	for _, us := range u.room.users {
		usersList += us.Name + Blue.Paint("/ ")
	}
	u.room.broadcast("", "README.md "+usersList+roomList)
}

func commandsCMD(_ string, u *User) {
	u.room.broadcast("", "Commands  \n"+autogenCommands(MainCMDs))
}
