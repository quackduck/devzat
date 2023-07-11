package main

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math"
	"math/rand"
	"os"
	"runtime/debug"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/acarl005/stripansi"
	"github.com/caarlos0/sshmarshal"
	"github.com/fatih/color"
	"github.com/gliderlabs/ssh"
	markdown "github.com/quackduck/go-term-markdown"
	cryptoSSH "golang.org/x/crypto/ssh"
)

var (
	Art  = getASCIIArt()
	CMDs []*[]CMD // slice of pointers to slices of commands (this is so updates to sub-slices are reflected in the main slice even if append is used)
)

func init() {
	CMDs = []*[]CMD{&MainCMDs, &RestCMDs, &SecretCMDs}
}

func getCMD(name string) (CMD, bool) {
	for _, cmds := range CMDs {
		if cmds == nil {
			Log.Println("nil slice in CMDs") // should never happen
			continue
		}
		for _, cmd := range *cmds {
			if cmd.name == name {
				return cmd, true
			}
		}
	}
	return CMD{}, false
}

func getASCIIArt() string {
	sep := string(os.PathSeparator)
	b, _ := os.ReadFile(Config.DataDir + sep + "art.txt")
	if b == nil {
		return "sorry, no art was found, please slap your developer and tell em to add a " + Config.DataDir + sep + "art.txt file"
	}
	return string(b)
}

func printUsersInRoom(r *Room) string {
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

func autogenCommands(cmds []CMD) string {
	b := new(bytes.Buffer)
	w := tabwriter.NewWriter(b, 0, 0, 2, ' ', 0)
	for _, cmd := range cmds {
		w.Write([]byte("   " + cmd.name + "\t" + cmd.argsInfo + "\t_" + cmd.info + "_  \n")) //nolint:errcheck // bytes.Buffer is never going to err out
	}
	w.Flush()
	return b.String()
}

// check if a User is an admin
func auth(u *User) bool {
	_, ok := Config.Admins[u.id]
	return ok
}

func keepSessionAlive(s ssh.Session) {
	for {
		time.Sleep(time.Minute * 3)
		_, err := s.SendRequest("keepalive@devzat", true, nil)
		if err != nil {
			return
		}
	}
}

func protectFromPanic() {
	if i := recover(); i != nil {
		MainRoom.broadcast(Devbot, "Slap the developers in the face for me, the server almost crashed, also tell them this: "+fmt.Sprint(i)+", stack: "+string(debug.Stack()))
	}
}

// removes arrows, spaces and non-ascii-printable characters
func cleanName(name string) string {
	s := ""
	name = strings.ReplaceAll(strings.TrimSpace(strings.Split(name, "\n")[0]), // use one trimmed line
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

func mdRender(a string, beforeMessageLen int, lineWidth int) string {
	a = strings.ReplaceAll(a, "https://", "https\\://")
	if strings.Contains(a, "![") && strings.Contains(a, "](") {
		lineWidth = int(math.Min(float64(lineWidth/2), 200)) // max image width is 200
	}
	md := strings.TrimSuffix(string(markdown.Render(a, lineWidth, beforeMessageLen)), "\n")
	if md == "" {
		return ""
	}
	if len(md) < beforeMessageLen {
		return md
	}
	return md[beforeMessageLen:]
}

// Returns true and the User with the same name if the username is taken, false and nil otherwise
func userDuplicate(r *Room, a string) (*User, bool) {
	for i := range r.users {
		if stripansi.Strip(r.users[i].Name) == stripansi.Strip(a) {
			return r.users[i], true
		}
	}
	return nil, false
}

func saveBans() {
	f, err := os.Create(Config.DataDir + string(os.PathSeparator) + "bans.json")
	if err != nil {
		Log.Println(err)
		return
	}
	defer f.Close()
	j := json.NewEncoder(f)
	j.SetIndent("", "   ")
	err = j.Encode(Bans)
	if err != nil {
		MainRoom.broadcast(Devbot, "error saving bans: "+err.Error())
		Log.Println(err)
		return
	}
}

func readBans() {
	f, err := os.Open(Config.DataDir + string(os.PathSeparator) + "bans.json")
	if err != nil {
		if !os.IsNotExist(err) {
			Log.Println(err)
		}
		return
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&Bans)
	if err != nil {
		MainRoom.broadcast(Devbot, "error reading bans: "+err.Error())
		Log.Println(err)
		return
	}
}

func findUserByName(r *Room, name string) (*User, bool) {
	r.usersMutex.RLock()
	defer r.usersMutex.RUnlock()
	for _, u := range r.users {
		if stripansi.Strip(u.Name) == name || "@"+stripansi.Strip(u.Name) == name {
			return u, true
		}
	}
	return nil, false
}

func remove(s []*User, a *User) []*User {
	for j := range s {
		if s[j] == a { // https://github.com/golang/go/wiki/SliceTricks#delete
			copy(s[j:], s[j+1:])
			s[len(s)-1] = nil
			return s[:len(s)-1]
		}
	}
	return s
}

func devbotChat(room *Room, line string) {
	if strings.Contains(line, "devbot") {
		if strings.HasPrefix(line, "kick ") || strings.HasPrefix(line, "ban ") { // devbot already replied in the command function
			return
		}
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
		devbotRespond(room, []string{
			":rocket:",
			"I like rockets",
			"SpaceX",
			"Elon Musk sus"}, 80)
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

func devbotRespond(room *Room, messages []string, chance int) {
	if chance == 100 || chance > rand.Intn(100) {
		go func() {
			time.Sleep(time.Second / 2)
			pick := messages[rand.Intn(len(messages))]
			room.broadcast(Devbot, pick)
		}()
	}
}

func shasum(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func holidaysCheck(u *User) {
	currentMonth := time.Now().Month()
	today := time.Now().Day()

	type holiday struct {
		month time.Month
		day   int
		name  string
		image string
	}

	holidayList := []holiday{
		{time.February, 14, "❤️ - Valentine's Day", "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/81/heavy-black-heart_2764.png"},
		{time.March, 17, "☘️ - St. Patrick's Day", "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/325/shamrock_2618-fe0f.png"},
		{time.April, 22, "🌎 - Earth Day", "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/325/globe-showing-americas_1f30e.png"},
		{time.May, 8, "👩 - Mother's Day", "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/325/woman_1f469.png"},
		{time.June, 19, "👨 - Father's Day", "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/325/man_1f468.png"},
		{time.September, 11, "👴 - Grandparents' Day", "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/325/old-woman_1f475.png"},
		{time.October, 31, "🎃 - Halloween", "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/325/jack-o-lantern_1f383.png"},
		{time.December, 25, "🎅 - Christmas", "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/325/santa-claus_1f385.png"},
		{time.December, 31, "🍾 - New Year's Eve", "https://emojipedia-us.s3.dualstack.us-west-1.amazonaws.com/thumbs/240/apple/325/bottle-with-popping-cork_1f37e.png"},
	}

	for _, h := range holidayList {
		if currentMonth == h.month && today == h.day {
			u.writeln("", "!["+h.name+"]("+h.image+")")
			time.Sleep(time.Second)
			clearCMD("", u)
			break
		}
	}
}

func printPrettyDuration(d time.Duration) string {
	s := d.Round(time.Minute).String()
	s = s[:len(s)-2] // cut off "0s" at the end
	if s == "" {     // we cut off the seconds so if there's nothing in the string it means it was made of only seconds.
		s = "< 1m"
	}
	return s
}

func fmtTime(u *User, lastStamp time.Time) string {
	if u.Timezone.Location == nil {
		diff := lastStamp.Sub(u.joinTime)
		if diff < 0 {
			return printPrettyDuration(-diff) + " earlier"
		}
		return printPrettyDuration(diff) + " in"
	}
	if u.FormatTime24 {
		return lastStamp.In(u.Timezone.Location).Format("15:04")
	}
	return lastStamp.In(u.Timezone.Location).Format("3:04")
}

// Check if the private key is there and if it is not, try to create it.
func checkKey(keyPath string) {
	_, err := os.Stat(keyPath)
	if err == nil {
		// Key exists, everything is fine and dandy.
		return
	}
	if !os.IsNotExist(err) { // the error is not a not-exist error. i.e. the file exists but there's some other problem with it
		Log.Printf("Error while checking for SSH keys in [%v]: %v\n", keyPath, err)
		return
	}

	Log.Printf("Generating new SSH server private key at %v\n", keyPath)
	privkey, pubkey, err := genKey()
	if err != nil {
		Log.Printf("Error while generating keypair: %v\n", err)
		return
	}
	privkeyFile, err := os.Create(keyPath)
	if err != nil {
		Log.Printf("Error while creating a file for the private key: %v\n", err)
		return
	}
	defer privkeyFile.Close()
	blk, err := sshmarshal.MarshalPrivateKey(privkey, "")
	if err != nil {
		Log.Printf("Error while marshalling private key: %v\n", err)
		return
	}
	if err := pem.Encode(privkeyFile, blk); err != nil {
		Log.Printf("Error while encoding private key: %v\n", err)
		return
	}
	Log.Println("Keys successfully generated!\nWhile the public key is not necessary for server operation, it may be useful to save it:\n" + color.YellowString(string(cryptoSSH.MarshalAuthorizedKey(pubkey))))
}

func genKey() (ed25519.PrivateKey, ssh.PublicKey, error) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		return nil, nil, err
	}
	sshPubKey, err := cryptoSSH.NewPublicKey(pub)
	if err != nil {
		return nil, nil, err
	}
	return priv, sshPubKey, nil
}
