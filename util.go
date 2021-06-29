package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
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
	//go:embed art.txt
	artBytes []byte
	admins   = []string{"d84447e08901391eb36aa8e6d9372b548af55bee3799cd3abb6cdd503fdf2d82", // Ishan Goel
		"f5c7f9826b6e143f6e9c3920767680f503f259570f121138b2465bb2b052a85d", // Ella Xu
		"6056734cc4d9fce31569167735e4808382004629a2d7fe6cb486e663714452fc", // Tommy Pujol
		"e9d47bb4522345d019086d0ed48da8ce491a491923a44c59fd6bfffe6ea73317", // Arav Narula
		"1eab2de20e41abed903ab2f22e7ff56dc059666dbe2ebbce07a8afeece8d0424", // Shok
		"12a9f108e7420460864de3d46610f722e69c80b2ac2fb1e2ada34aa952bbd73e", // jmw: github.com/ciearius
		"2433e7c03997d13f9117ded9e36cd2d23bddc4d588b8717c4619bedeb3b7e9ad"} // @epic: github.com/TAG-Epic
)

func printUsersInRoom(r *room) string {
	names := make([]string, 0, len(r.users))
	for _, us := range r.users {
		names = append(names, us.name)
	}
	return fmt.Sprint(names)
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
	//return true
	for _, id := range admins {
		if u.id == id || u.addr == id {
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
	if len(name) > 27 {
		name = name[:27]
	}
	for _, r := range name {
		if unicode.IsPrint(r) {
			s += string(r)
		}
	}
	return s
}

func printPrettyDuration(d time.Duration) string {
	//return strings.TrimSuffix(mainroom.Round(time.Minute).String(), "0s")
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
	if err != nil {
		l.Println(err)
		return
	}
	//bansMutex.Lock()
	json.NewDecoder(f).Decode(&bans)
	//bansMutex.Unlock()
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
	if line == "help" || line == "/help" || strings.Contains(line, "help me") {
		devbotRespond(room, []string{"Run ./help to get help!",
			"Looking for ./help?",
			"See available commands with ./commands or see help with ./help :star:"}, 100)
	}
	if line == "ls" {
		devbotRespond(room, []string{"./help", "Not a shell.", "bruv", "yeah no, this is not your regular ssh server"}, 100)
	}
	if strings.Contains(line, "rm -rf") {
		devbotRespond(room, []string{"rm -rf you", "I've heard rm -rf / can really free up some space!\n\n you should try it on your computer", "evil"}, 100)
		return
	}
	if strings.HasPrefix(line, "rm") {
		devbotRespond(room, []string{"Bad human, bad human", "haha, permission denied", "this is not your regular ssh server", "hehe", "bruh"}, 100)
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

func stringsContain(a []string, s string) bool {
	for i := range a {
		if a[i] == s {
			return true
		}
	}
	return false
}
