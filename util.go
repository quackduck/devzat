package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"math"
	"math/rand"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/acarl005/stripansi"
	markdown "github.com/quackduck/go-term-markdown"
)

var (
	art = getASCIIArt()
)

func getASCIIArt() string {
	b, _ := ioutil.ReadFile(Config.DataDir + "/art.txt")
	if b == nil {
		return "sorry, no art was found, please slap your developer and tell em to add a " + Config.DataDir + "/art.txt file"
	}
	return string(b)
}

func printUsersInRoom(r *room) string {
	names := ""
	admins := ""
	for _, us := range r.users {
		if auth(us) {
			admins += us.Name + " "
			continue
		}
		names += us.Name + " "
	}
	if len(names) > 0 {
		names = names[:len(names)-1] // cut extra space at the end
	}
	names = "[" + names + "]"
	if len(admins) > 0 {
		admins = admins[:len(admins)-1]
	}
	admins = "[" + admins + "]"
	return names + " Admins: " + admins
}

func lenString(a string) int {
	return len([]rune(stripansi.Strip(a)))
}

func autogenCommands(cmds []cmd) string {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 0, 2, ' ', 0)
	for _, cmd := range cmds {
		w.Write([]byte("   " + cmd.name + "\t" + cmd.argsInfo + "\t_" + cmd.info + "_  \n")) //nolint:errcheck // bytes.Buffer is never going to err out
	}
	w.Flush()
	return b.String()
}

// check if a user is an admin
func auth(u *user) bool {
	_, ok := Config.Admins[u.id]
	return ok
}

// removes arrows, spaces and non-ascii-printable characters
func cleanName(name string) string {
	s := ""
	name = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
		strings.TrimSpace(strings.Split(name, "\n")[0]), // use one trimmed line
		"<-", ""),
		"->", ""),
		" ", "-")
	if len([]rune(name)) > 27 {
		name = string([]rune(name)[:27])
	}
	for i := 0; i < len(name); i++ {
		if 33 <= name[i] && name[i] <= 126 { // ascii printables only: '!' to '~'
			s += string(name[i])
		}
	}
	return s
}

func printPrettyDuration(d time.Duration) string {
	s := strings.TrimSpace(strings.TrimSuffix(d.Round(time.Minute).String(), "0s"))
	if s == "" { // we cut off the seconds so if there's nothing in the string it means it was made of only seconds.
		s = "< 1m"
	}
	return s
}

func mdRender(a string, beforeMessageLen int, lineWidth int) string {
	if strings.Contains(a, "![") && strings.Contains(a, "](") {
		lineWidth = int(math.Min(float64(lineWidth/2), 200)) // max image width is 200
	}
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

// Returns true and the user with the same name if the username is taken, false and nil otherwise
func userDuplicate(r *room, a string) (*user, bool) {
	for i := range r.users {
		if stripansi.Strip(r.users[i].Name) == stripansi.Strip(a) {
			return r.users[i], true
		}
	}
	return nil, false
}

func saveBans() {
	f, err := os.Create(Config.DataDir + "/bans.json")
	if err != nil {
		l.Println(err)
		return
	}
	defer f.Close()
	j := json.NewEncoder(f)
	j.SetIndent("", "   ")
	err = j.Encode(bans)
	if err != nil {
		mainRoom.broadcast(devbot, "error saving bans: "+err.Error())
		l.Println(err)
		return
	}
}

func readBans() {
	f, err := os.Open(Config.DataDir + "/bans.json")
	if err != nil && !os.IsNotExist(err) { // if there is an error and it is not a "file does not exist" error
		l.Println(err)
		return
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&bans)
	if err != nil {
		mainRoom.broadcast(devbot, "error reading bans: "+err.Error())
		l.Println(err)
		return
	}
}

func findUserByName(r *room, name string) (*user, bool) {
	r.usersMutex.Lock()
	defer r.usersMutex.Unlock()
	for _, u := range r.users {
		if stripansi.Strip(u.Name) == name {
			return u, true
		}
	}
	return nil, false
}

func remove(s []*user, a *user) []*user {
	for j := range s {
		if s[j] == a {
			return append(s[:j], s[j+1:]...)
		}
	}
	return s
}

func devbotChat(room *room, line string) {
	if strings.Contains(line, "devbot") {
		if strings.Contains(line, "how are you") || strings.Contains(line, "how you") {
			devbotRespond(room, []string{"How are _you_",
				"Good as always lol",
				"Ah the usual, solving quantum gravity :smile:",
				"Howdy?",
				"Thinking about intergalactic cows",
				"Could maths be different in other universes?",
				""}, 99)
			return
		}
		if strings.Contains(line, "thank") {
			devbotRespond(room, []string{"you're welcome",
				"no problem",
				"yeah dw about it",
				":smile:",
				"no worries",
				"you're welcome man!",
				"lol"}, 93)
			return
		}
		if strings.Contains(line, "good") || strings.Contains(line, "cool") || strings.Contains(line, "awesome") || strings.Contains(line, "amazing") {
			devbotRespond(room, []string{"Thanks haha", ":sunglasses:", ":smile:", "lol", "haha", "Thanks lol", "yeeeeeeeee"}, 93)
			return
		}
		if strings.Contains(line, "bad") || strings.Contains(line, "idiot") || strings.Contains(line, "stupid") {
			devbotRespond(room, []string{"what an idiot, bullying a bot", ":(", ":angry:", ":anger:", ":cry:", "I'm in the middle of something okay", "shut up", "Run ./help, you need it."}, 60)
			return
		}
		if strings.Contains(line, "shut up") {
			devbotRespond(room, []string{"NO YOU", "You shut up", "what an idiot, bullying a bot"}, 90)
			return
		}
		devbotRespond(room, []string{"Hi I'm devbot", "Hey", "HALLO :rocket:", "Yes?", "Devbot to the rescue!", ":wave:"}, 90)
	}
	if line == "./help" || line == "/help" || strings.Contains(line, "help me") {
		devbotRespond(room, []string{"Run help to get help!",
			"Looking for help?",
			"See available commands with cmds or see help with help :star:"}, 100)
	}
	if line == "easter" {
		devbotRespond(room, []string{"eggs?", "bunny?"}, 100)
	}
	if strings.Contains(line, "rm -rf") {
		devbotRespond(room, []string{"rm -rf you", "I've heard rm -rf / can really free up some space!\n\n you should try it on your computer", "evil"}, 100)
		return
	}
	if strings.Contains(line, "where") && strings.Contains(line, "repo") {
		devbotRespond(room, []string{"The repo's at github.com/quackduck/devzat!", ":star: github.com/quackduck/devzat :star:", "# github.com/quackduck/devzat"}, 100)
	}
	if strings.Contains(line, "rocket") || strings.Contains(line, "spacex") || strings.Contains(line, "tesla") {
		devbotRespond(room, []string{"Doge to the mooooon :rocket:",
			"I should have bought ETH before it :rocket:ed to the :moon:",
			":rocket:",
			"I like rockets",
			"SpaceX",
			"Elon Musk OP"}, 80)
	}
	if strings.Contains(line, "elon") {
		devbotRespond(room, []string{"When something is important enough, you do it even if the odds are not in your favor. - Elon",
			"I do think there is a lot of potential if you have a compelling product - Elon",
			"If you're trying to create a company, it's like baking a cake. You have to have all the ingredients in the right proportion. - Elon",
			"Patience is a virtue, and I'm learning patience. It's a tough lesson. - Elon"}, 75)
	}
	if !strings.Contains(line, "start") && strings.Contains(line, "star") {
		devbotRespond(room, []string{"Someone say :star:?",
			"If you like Devzat, give it a star at github.com/quackduck/devzat!",
			":star: github.com/quackduck/devzat", ":star:"}, 90)
	}
	if strings.Contains(line, "cool project") || strings.Contains(line, "this is cool") || strings.Contains(line, "this is so cool") {
		devbotRespond(room, []string{"Thank you :slight_smile:!",
			" If you like Devzat, do give it a star at github.com/quackduck/devzat!",
			"Star Devzat here: github.com/quackduck/devzat"}, 90)
	}
}

func devbotRespond(room *room, messages []string, chance int) {
	if chance == 100 || chance > rand.Intn(100) {
		go func() {
			time.Sleep(time.Second / 2)
			pick := messages[rand.Intn(len(messages))]
			room.broadcast(devbot, pick)
		}()
	}
}

func shasum(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}
