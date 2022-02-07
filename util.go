package main

import (
	"bytes"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"text/tabwriter"
	"time"
	"unicode"

	"github.com/acarl005/stripansi"
	markdown "github.com/quackduck/go-term-markdown"
)

var (
	art    = getASCIIArt()
	admins = getAdmins()
)

func getAdmins() []string {
	data, err := ioutil.ReadFile("admins.json")
	if err != nil {
		fmt.Println("Error reading admins.json:", err, ". Make an admins.json file to add admins.")
		return []string{}
	}
	var adminsList map[string]string // id to info
	err = json.Unmarshal(data, &adminsList)
	if err != nil {
		return []string{}
	}
	ids := make([]string, 0, len(adminsList))
	for id := range adminsList {
		ids = append(ids, id)
	}
	return ids
}

func getASCIIArt() string {
	b, _ := ioutil.ReadFile("art.txt")
	if b == nil {
		return "sowwy, no art was found, please slap your developer and tell em to add an art.txt file"
	}
	return string(b)
}

func printUsersInRoom(r *room) string {
	names := ""
	if len(r.users) == 0 {
		return names
	}
	for _, us := range r.users {
		names += us.name + " "
	}
	names = names[:len(names)-1] // cut extra space at the end
	names = "[" + names + "]"
	return names
}

func lenString(a string) int {
	return len([]rune(stripansi.Strip(a)))
}

func autogenCommands(cmds []cmd) string {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 0, 2, ' ', 0)
	for _, cmd := range cmds {
		w.Write([]byte("   " + cmd.name + "\t" + cmd.argsInfo + "\t_" + cmd.info + "_  \n"))
	}
	w.Flush()
	return b.String()
}

// check if a user is an admin
func auth(u *user) bool {
	for _, id := range admins {
		if u.id == id {
			return true
		}
	}
	return false
}

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
	for _, r := range name {
		if unicode.IsPrint(r) {
			s += string(r)
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
func userDuplicate(r *room, a string) bool {
	for i := range r.users {
		if stripansi.Strip(r.users[i].name) == stripansi.Strip(a) {
			return true
		}
	}
	return false
}

func saveBans() {
	f, err := os.Create("bans.json")
	if err != nil {
		l.Println(err)
		return
	}
	j := json.NewEncoder(f)
	j.SetIndent("", "   ")
	j.Encode(bans)
	f.Close()
}

func readBans() {
	f, err := os.Open("bans.json")
	if err != nil && !os.IsNotExist(err) { // if there is an error and it is not a "file does not exist" error
		l.Println(err)
		return
	}
	json.NewDecoder(f).Decode(&bans)
	f.Close()
}

func findUserByName(r *room, name string) (*user, bool) {
	r.usersMutex.Lock()
	defer r.usersMutex.Unlock()
	for _, u := range r.users {
		if stripansi.Strip(u.name) == name {
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
	if canDevbotReply(line, []string{"devbot"}, []string{"how are you", "how you"}, []string{}) {
		devbotRespond(room, []string{"How are _you_",
			"Good as always lol",
			"Ah the usual, solving quantum gravity :smile:",
			"Howdy?",
			"Thinking about intergalactic cows",
			"Could maths be different in other universes?",
			""}, 99)
		return
	}
	if canDevbotReply(line, []string{"devbot", "thanks"}, []string{}, []string{}) {
		devbotRespond(room, []string{"you're welcome",
			"no problem",
			"yeah dw about it",
			":smile:",
			"no worries",
			"you're welcome man!",
			"lol"}, 93)
		return
	}
	if canDevbotReply(line, []string{"devbot"}, []string{"good", "cool", "awesome", "amazing"}, []string{}) {
		devbotRespond(room, []string{"Thanks haha", ":sunglasses:", ":smile:", "lol", "haha", "Thanks lol", "yeeeeeeeee"}, 93)
		return
	}
	if canDevbotReply(line, []string{"devbot"}, []string{"bad", "idiot", "stupid"}, []string{}) {
		devbotRespond(room, []string{"what an idiot, bullying a bot", ":(", ":angry:", ":anger:", ":cry:", "I'm in the middle of something okay", "shut up", "Run ./help, you need it."}, 60)
		return
	}
	if canDevbotReply(line, []string{"devbot", "shut up"}, []string{}, []string{}) {
		devbotRespond(room, []string{"NO YOU", "You shut up", "what an idiot, bullying a bot"}, 90)
		return
	}
	if canDevbotReply(line, []string{"devbot"}, []string{}, []string{}) {
		devbotRespond(room, []string{"Hi I'm devbot", "Hey", "HALLO :rocket:", "Yes?", "Devbot to the rescue!", ":wave:"}, 90)
	}
	if canDevbotReply(line, []string{}, []string{"help", "/help", "helm me"}, []string{}) {
		devbotRespond(room, []string{"Run help to get help!",
			"Looking for help?",
			"See available commands with cmds or see help with help :star:"}, 100)
	}
	if canDevbotReply(line, []string{"easter"}, []string{}, []string{}) {
		devbotRespond(room, []string{"eggs?", "bunny?"}, 100)
	}
	if canDevbotReply(line, []string{"rm -rf"}, []string{}, []string{}) {
		devbotRespond(room, []string{"rm -rf you", "I've heard rm -rf / can really free up some space!\n\n you should try it on your computer", "evil"}, 100)
		return
	}
	if canDevbotReply(line, []string{"where", "repo"}, []string{}, []string{}) {
		devbotRespond(room, []string{"The repo's at github.com/quackduck/devzat!", ":star: github.com/quackduck/devzat :star:", "# github.com/quackduck/devzat"}, 100)
	}
	if canDevbotReply(line, []string{}, []string{"rocket", "spacex", "tesla"}, []string{}) {
		devbotRespond(room, []string{"Doge to the mooooon :rocket:",
			"I should have bought ETH before it :rocket:ed to the :moon:",
			":rocket:",
			"I like rockets",
			"SpaceX",
			"Elon Musk OP"}, 80)
	}
	if canDevbotReply(line, []string{}, []string{"elon", "Elon"}, []string{}) {
		devbotRespond(room, []string{"When something is important enough, you do it even if the odds are not in your favor. - Elon",
			"I do think there is a lot of potential if you have a compelling product - Elon",
			"If you're trying to create a company, it's like baking a cake. You have to have all the ingredients in the right proportion. - Elon",
			"Patience is a virtue, and I'm learning patience. It's a tough lesson. - Elon"}, 75)
	}
	if canDevbotReply(line, []string{"star"}, []string{}, []string{"start"}) {
		devbotRespond(room, []string{"Someone say :star:?",
			"If you like Devzat, give it a star at github.com/quackduck/devzat!",
			":star: github.com/quackduck/devzat", ":star:"}, 90)
	}
	if canDevbotReply(line, []string{}, []string{"cool project", "this is cool", "this is so cool"}, []string{}) {
		devbotRespond(room, []string{"Thank you :slight_smile:!",
			" If you like Devzat, do give it a star at github.com/quackduck/devzat!",
			"Star Devzat here: github.com/quackduck/devzat"}, 90)
	}
}

// This function returns true if `line` contains all elements from
// `containsAllOf`, any element from `andAnyOf`, but no elements from
// `butNotAnyOf`. If any of the list is empty, the condition associated with
// it will default to true.
func canDevbotReply(line string, containsAllOf []string, andAnyOf []string, butNotAnyOf []string) bool {
	if len(containsAllOf) != 0 {
		for _, s := range containsAllOf {
			if !strings.Contains(line, s) {
				return false
			}
		}
	}
	if len(butNotAnyOf) != 0 {
		for _, s := range butNotAnyOf {
			if strings.Contains(line, s) {
				return false
			}
		}
	}
	if len(andAnyOf) != 0 {
		for _, s := range andAnyOf {
			if strings.Contains(line, s) {
				return true
			}
		}
	} else {
		return true
	}
	return false
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
